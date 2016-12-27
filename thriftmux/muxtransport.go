package thriftmux

import (
	"git.apache.org/thrift.git/lib/go/thrift"
	"github.com/bryce-anderson/mux"
	"bytes"
)

func NewThriftMuxTransport(session mux.ClientSession) thrift.TTransport {
	muxt := &MuxTransport{
		session: session,
	}

	return muxt
}

type MuxTransport struct {
	session mux.ClientSession
	buffer bytes.Buffer
	response bytes.Reader
}

func (m *MuxTransport) Close() error {
	panic("not implemented")
}

func (m *MuxTransport) Read(data []byte) (int, error) {
	return m.response.Read(data)
}

func (m *MuxTransport) Write(data []byte) (int, error) {
	return m.buffer.Write(data)
}

func (m *MuxTransport) Flush() (err error) {
	// Now we need to write the data to the wire

	var resp []byte
	resp, err = mux.SimpleDispatch(m.session, m.buffer.Bytes())

	// Clear our outbound buffer
	m.buffer.Reset()

	if err != nil {
		return
	}

	m.response.Reset(resp)

	return
}

func (m *MuxTransport) Open() error {
	panic("not implemented")
}

func (m *MuxTransport) IsOpen() bool {
	panic("not implemented")
}

func (m *MuxTransport) RemainingBytes() uint64 {
	return uint64(m.response.Len())
}
