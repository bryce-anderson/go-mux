package mux

import "io"

////////////////////////////
// Message types
////////////////////////////

// Ping message sent by the client to the server
type Tping struct {}

// Ping response sent by the server to the client
type Rping struct {}


////////////////////////////
// Tping Message methods
////////////////////////////

// Tping
func (t *Tping) Type() int8 {
	return TpingTpe
}

func (_ *Tping) Size() int {
	return 0
}

func (_ *Tping) Encode(write io.Writer) error {
	return nil
}

////////////////////////////
// Rping Message methods
////////////////////////////

func (_ *Rping) Encode(write io.Writer) error {
	return nil
}

func (_ *Rping) Type() int8 {
	return RpingTpe
}

func (_ *Rping) Size() int {
	return 0
}
