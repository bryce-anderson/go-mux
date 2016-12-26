package mux

import "io"

// Error response sent from the server to the client
type Rerr struct {
	error string
}

////////////////////////////
// Rerr Message methods
////////////////////////////

func (r *Rerr) Type() int8 {
	return RerrTpe
}

func (r *Rerr) Size() int {
	return len([]byte(r.error))
}

func (r *Rerr) Encode(writer io.Writer) error {
	_, err := writer.Write([]byte(r.error))
	return err
}

////////////////////////////
// Rerr error method
////////////////////////////

func (r *Rerr) Error() string {
	return r.error
}