package mux

import (
	"io"
	"bufio"
	"bytes"
	"sync"
	"errors"
	"log"
	"encoding/binary"
	"fmt"
	"time"
	"math"
)

const LatestMuxVersion = 0x0001

type ClientSession interface {
	Close() error
	Dispatch(msg Tdispatch) (Rdispatch, error)
	Ping() (time.Duration, error)
}

func SimpleDispatch(s ClientSession, data []byte) (out []byte, err error) {
	disp := Tdispatch{
		Body: data,
	}
	var rdisp Rdispatch
	rdisp,err = s.Dispatch(disp)
	if err != nil {
		return
	}
	out = rdisp.Body
	return
}

// Basic client dispatcher
type baseStreamState struct {
	streamId int32          // stream id associated with this stream
	parent   *baseClientSession
	input    chan Message // only a single element will go in the channel
}

func (b *baseStreamState) discard() {
	close(b.input)
	session := b.parent

	session.lock.Lock()
	defer session.lock.Unlock()

	delete(session.streams.openStreams, b.streamId)
}

// TODO: we don't have any information about error in this
type baseClientSession struct {
	lock    sync.RWMutex
	rw      *bufio.ReadWriter
	streams baseTracker
	c       codec
}

func (s *baseClientSession) Ping() (out time.Duration, err error) {
	now := time.Now()

	panic("This is not a real function...")
	end := time.Now()

	out = now.Sub(end)
	return
}

func (s *baseClientSession) Dispatch(in Tdispatch) (out Rdispatch, err error) {
	var stream *baseStreamState

	stream, err = makeStreamState(s)
	if err != nil {
		return
	}

	defer stream.discard()

	frame := Frame{
		streamId: stream.streamId,
		message: &in,
	}

	err = s.c.encodeFrame(s.rw, &frame)
	if err != nil {
		return
	}

	message, ok := <- stream.input

	if !ok {
		err = io.EOF
		return
	}

	switch msg := message.(type) {
	case *Rdispatch:
		out = *msg

	case *Rerr:
		err = errors.New(msg.error)

	default:
		err = errors.New(fmt.Sprintf("Unexpcted message type: %d", msg.Type()))
	}

	return
}


func baseReadLoop(s *baseClientSession) {
	for {
		frame, err := s.c.decodeFrame(s.rw)
		if err != nil {
			log.Println("Graceful shutdown not implemented. Error: " + err.Error())
			break
		}

		switch msg := frame.message.(type) {
		case *Rdispatch:
			baseSessionRdispatch(s, &frame, msg)

		default:
			panic("I don't know what this is!")
		}
	}
}

func baseSessionRdispatch(session *baseClientSession, frame *Frame, d *Rdispatch) (err error) {
	tracker := &session.streams

	session.lock.RLock()
	defer session.lock.RUnlock()

	stream, ok := tracker.openStreams[frame.streamId]
	if !ok {
		log.Printf("Received Rdispatch frame from non-existant stream %d", frame.streamId)
		return
	}

	stream.input <- d
	return
}

func (s *baseClientSession) Close() error {
	panic("not implemented")
}

// Intended to be a simple constructor function
func makeStreamState(parent *baseClientSession) (state *baseStreamState, err error) {
	parent.lock.Lock()
	defer parent.lock.Unlock()

	state = &baseStreamState{}

	state.streamId, err = parent.streams.nextStreamId()
	if err != nil {
		return
	}

	state.parent = parent
	state.input = make(chan Message)

	parent.streams.openStreams[state.streamId] = state

	return
}

func NewClientSession(raw io.ReadWriteCloser, maxFrameSize int32) (session ClientSession, err error) {

	if maxFrameSize == math.MaxInt32 {
		// No need to negotiate framing, its slower anyway
		session = newBaseClientSession(raw)
		return
	}

	var canInit bool
	canInit, err = canTinit(raw)
	if err != nil {
		return
	}

	if canInit {
		headers := ClientHeaders(maxFrameSize)
		err = sendInit(raw, LatestMuxVersion, headers)
		if err != nil {
			return
		}

		var frame Frame
		frame, err = DecodeFrame(raw)
		if err != nil {
			return
		}

		switch rmsg := frame.message.(type) {
		case *Rinit:
			fmt.Printf("Found rinit msg. Headers:\n")
			for _, h := range rmsg.headers {
				fmt.Printf("('%s', '%s')\n", string(h.Key), string(h.Value))
			}
			// TODO: we have nogotiated to mux v1... We should use it.
			panic("Not implemented")

		default:
			err = errors.New(fmt.Sprintf("Failed to init! Type: %d\n", frame.message.Type()))
		}
	} else {
		// Don't support protocol nogotiation, just use the default.
		session = newBaseClientSession(raw)
	}
	return
}

func newBaseClientSession(raw io.ReadWriteCloser) ClientSession {
	rw :=  bufio.NewReadWriter(bufio.NewReader(raw), bufio.NewWriter(raw))

	bSession := &baseClientSession{
		rw: rw,
		c: &baseCodec{},
		streams: newBaseStreamTracker(),
	}

	go baseReadLoop(bSession)
	return bSession
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

func sendInit(writer io.Writer, version int16, headers []Header) (err error) {
	msg := Tinit{
		version: version,
		headers: headers,
	}

	err = EncodeFrame(writer, &Frame{
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

type baseTracker struct {
	nextId      int32
	openStreams map[int32]*baseStreamState
}

func newBaseStreamTracker() baseTracker {
	return baseTracker{
		nextId:      1,
		openStreams: make(map[int32]*baseStreamState),
	}
}

func (s *baseTracker) nextStreamId() (int32, error) {
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