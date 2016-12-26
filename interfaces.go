package mux

import "io"

////////////////////////////
// Message interface
////////////////////////////
type Message interface {
	Type() int8                   // Message type
	Size() int                    // Message size not counting type tag or length prefix
	Encode(write io.Writer) error // Encode message into the stream. Doesn't include the type type, streamid, or frame length prefix
}
