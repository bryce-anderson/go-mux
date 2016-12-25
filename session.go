package mux

import (
	"io"
	"bufio"
	"sync"
	"errors"
	"log"
	"io/ioutil"
)

type Stream interface {
	io.ReadWriteCloser
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
	dispatch := Tdispatch {
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

	err = EncodeFrame(&frame, s.parent.rw)
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
	openStreams map[int32]*streamState
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

	nextId, err := s.nextStreamId()
	if err != nil {
		return
	}

	str := makeStreamState(nextId, s)
	stream = str
	s.openStreams[nextId] = str


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

func clientSessionRdispatch(session *clientSession, frame *Frame, d *Rdispatch) (err error) {
	session.lock.Lock()
	defer session.lock.Unlock()

	stream, ok := session.openStreams[frame.streamId]
	if !ok {
		log.Printf("Received Rdispatch frame from non-existant stream %d", frame.streamId)
		return
	}

	stream.input <- d.Body

	if !frame.isFragment {
		close(stream.input)
		delete(session.openStreams, frame.streamId)
	}

	return
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

// unsynchronized: must be called from a location protected by the mutex
func (s *clientSession) nextStreamId() (int32, error) {
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

func NewClientSession(raw io.ReadWriteCloser) ClientSession {
	rw :=  bufio.NewReadWriter(bufio.NewReader(raw), bufio.NewWriter(raw))
	session := &clientSession{
		nextId: 1,
		raw: raw,
		rw: rw,
		openStreams: make(map[int32]*streamState),
	}

	go clientSessionReadLoop(session)
	return session
}
