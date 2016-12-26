package mux

import (
	"io"
	"bufio"
	"sync"
	"errors"
	"log"
	"io/ioutil"
	"encoding/binary"
	"bytes"
	"fmt"
)

const LatestMuxVersion = 0x0001

type Stream interface {
	io.ReadWriteCloser
}

type streamTracker struct {
	nextId      int32
	openStreams map[int32]*streamState
}

func newStreamTracker() streamTracker {
	return streamTracker{
		nextId:      1,
		openStreams: make(map[int32]*streamState),
	}
}

// unsynchronized: must be called from a location protected by the mutex
func (s *streamTracker) nextStreamId() (int32, error) {
	start := s.nextId

	for {
		i := s.nextId
		_, occupied := s.openStreams[i]

		s.nextId += 1
		if s.nextId > MaxStreamId {
			s.nextId = 1
		}

		if !occupied {
			return i, nil // success
		}

		if s.nextId == start {
			return 0, errors.New("Exhausted tag numbers")
		}

	}
}

type ClientSession interface {
	NewStream() (Stream, error)
	Close() error
	Dispatch(data []byte) ([]byte, error)
}

type streamState struct {
	id int32 // stream id associated with this stream
	parent *clientSession
	input chan []byte
	buffer []byte
}

func (s *streamState) Close() error {
	panic("not implemented")
}

func (s *streamState) Read(p []byte) (n int, err error) {
	if len(s.buffer) == 0 {
		// need to read some data from the queue
		data, ok := <- s.input
		if !ok {
			err = io.EOF
			return
		}
		s.buffer = data
	}

	n = copy(p, s.buffer)
	s.buffer = s.buffer[n:]	// lop off our remainder

	return
}

func (s *streamState) Write(p []byte) (n int, err error) {
	dispatch := Tdispatch{
		Contexts: []Header{},
		Dest: 	  "/a/good/path",
		Dtabs:    []Header{},
		Body: p,
	}

	frame := Frame{
		frameType: dispatch.Type(),
		streamId:  s.id,
		message: &dispatch,
	}

	s.parent.lock.Lock()
	defer s.parent.lock.Unlock()

	err = EncodeFrame(s.parent.rw, &frame)
	if err != nil {
		return
	}

	// Gotta make sure to flush so that the line gets the bytes
	err = s.parent.rw.Flush()
	if err != nil {
		return
	}

	n = len(p)
	return
}

// TODO: we don't have any information about error in this
type clientSession struct {
	lock        sync.Mutex
	nextId      int32
	raw         io.ReadWriteCloser
	rw    	    *bufio.ReadWriter
	streams     streamTracker
}

func (s *clientSession) Dispatch(in []byte) (out []byte, err error) {
	stream, err := s.NewStream()

	if err != nil {
		return
	}

	_, err = stream.Write(in)
	if err != nil {
		return
	}

	out, err = ioutil.ReadAll(stream)
	return
}


func (s *clientSession) NewStream() (stream Stream, err error) {
	s.lock.Lock()
	defer s.lock.Unlock()
	nextId, err := s.streams.nextStreamId()
	if err != nil {
		return
	}

	str := makeStreamState(nextId, s)
	stream = str
	s.streams.openStreams[nextId] = str

	return
}

func clientSessionReadLoop(s *clientSession) {
	for {
		frame, err := DecodeFrame(s.rw)
		if err != nil {
			panic("Graceful shutdown not implemented. Error: " + err.Error())
		}

		switch msg := frame.message.(type) {
		case *Rdispatch:
			clientSessionRdispatch(s, &frame, msg)

		default:
			panic("I don't know what this is!")
		}
	}
}

func streamRdispatch(tracker *streamTracker, frame *Frame, d *Rdispatch) (err error) {

	stream, ok := tracker.openStreams[frame.streamId]
	if !ok {
		log.Printf("Received Rdispatch frame from non-existant stream %d", frame.streamId)
		return
	}

	stream.input <- d.Body

	if !frame.isFragment {
		close(stream.input)
		delete(tracker.openStreams, frame.streamId)
	}

	return
}

func clientSessionRdispatch(session *clientSession, frame *Frame, d *Rdispatch) (err error) {
	session.lock.Lock()
	defer session.lock.Unlock()

	return streamRdispatch(&session.streams, frame, d)
}

func (s *clientSession) Close() error {
	panic("not implemented")
}

func makeStreamState(id int32, parent *clientSession) *streamState {
	return &streamState{
		id: id,
		parent: parent,
		input: make(chan []byte),
	}
}

func NewClientSession(raw io.ReadWriteCloser) ClientSession {
	canInit, _ := canTinit(raw)

	if canInit {
		fmt.Printf("We can tinit!\n")
		err := negotiate(raw)
		if err != nil {
			panic("Failed to negotiate: " + err.Error())
		}
	}

	// We use buffered reader and writers since they are more efficient
	rw :=  bufio.NewReadWriter(bufio.NewReader(raw), bufio.NewWriter(raw))

	session := &clientSession{
		nextId: 1,
		raw: raw,
		rw: rw,
		streams: newStreamTracker(),
	}

	go clientSessionReadLoop(session)
	return session
}

func negotiate(rw io.ReadWriter) (err error) {
	headers := ClientHeaders(1000)
	err = sendInit(rw, LatestMuxVersion, headers)
	if err != nil {
		return
	}

	var frame Frame
	frame, err = DecodeFrame(rw)
	if err != nil {
		return
	}

	switch rmsg := frame.message.(type) {
	case *Rinit:
		fmt.Printf("Found rinit msg. Headers:\n")
		for _, h := range rmsg.headers {
			fmt.Printf("('%s', '%s')\n", string(h.Key), string(h.Value))
		}

	default:
		fmt.Printf("Failed to init! Type: %d\n", frame.frameType)
	}

	return
}

func sendInit(writer io.Writer, version int16, headers []Header) (err error) {
	msg := Tinit{
		version: version,
		headers: headers,
	}

	err = EncodeFrame(writer, &Frame{
		frameType:	msg.Type(),
		isFragment:	false,
		streamId:  	1,
		message:   	&msg,
	})

	return
}

func canTinit(rw io.ReadWriter) (canInit bool, err error) {

	msg := Rerr{
		error: "tinit check",
	}

	err = EncodeFrame(rw, &Frame{
		frameType:	msg.Type(),
		isFragment:	false,
		streamId:  	1,
		message:   	&msg,
	})
	if err != nil {
		return
	}

	var frame Frame
	frame, err = DecodeFrame(rw)

	if frame.streamId != 1 {
		canInit = false
		return
	}

	switch rmsg := frame.message.(type) {
	case *Rerr:
		canInit = msg.error == rmsg.error

	default:
		canInit = false
	}

	return
}

func ClientHeaders(maxFrameSize int32) []Header {
	buffer := bytes.NewBuffer([]byte{})
	binary.Write(buffer, binary.BigEndian, maxFrameSize)

	header := Header{
		Key: []byte("mux-framer"),
		Value: buffer.Bytes(),
	}

	return []Header {header}
}
