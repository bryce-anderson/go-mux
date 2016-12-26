package mux

import (
	"io"
	"encoding/binary"
)

// dispatch message sent by the client and received by the server
type Tdispatch struct {
	Contexts []Header
	Dest 	 string
	Dtabs    []Header
	Body     []byte
}

// response to a dispatch sent by a server and received by the client
type Rdispatch struct {
	Status   int8
	Contexts []Header
	Body     []byte
}

////////////////////////////
// Tdispatch Message methods
////////////////////////////

func (t *Tdispatch) Size() int {
	acc := headersSize(t.Contexts)
	acc += 2 + len(t.Dest)
	acc += headersSize(t.Dtabs)
	acc += len(t.Body)
	return acc
}

func (t *Tdispatch) Type() int8 {
	return TdispatchTpe
}

func (t *Tdispatch) Encode(writer io.Writer) (err error) {
	err = encodeHeaders(writer, t.Contexts)
	if err != nil {
		return
	}

	err = encodeBytesInt16(writer, []byte(t.Dest))
	if err != nil {
		return
	}

	err = encodeHeaders(writer, t.Dtabs)
	if err != nil {
		return
	}

	_, err = writer.Write(t.Body)

	return
}


////////////////////////////
// Rdispatch Message methods
////////////////////////////

func (r *Rdispatch) Type() int8 {
	return RdispatchTpe
}

func (r *Rdispatch) Size() int {
	acc := 1	// status
	acc += headersSize(r.Contexts)
	acc += len(r.Body)
	return acc
}

func (r *Rdispatch) Encode(writer io.Writer) (err error) {
	// encode the status
	binary.Write(writer, binary.BigEndian, r.Status)
	// encode headers
	err = encodeHeaders(writer, r.Contexts)

	if err != nil {
		return
	}

	_, err = writer.Write(r.Body)

	return
}

////////////////////////////
// decoding methods
////////////////////////////

// decode a Tdispatch frame from the input. The size is the payload of the Tdispatch frame
func decodeTdispatch(input io.Reader, size int32) (msg *Tdispatch, err error) {
	var contexts, dtabs []Header
	var dest []byte
	contexts, err = decodeHeaders(input)
	if err != nil {
		return
	}

	dest, err = decodeInt16Bytes(input)
	if err != nil {
		return
	}

	dtabs, err = decodeHeaders(input)
	if err != nil {
		return
	}

	bodysize := int(size) - headersSize(contexts) - headersSize(dtabs)
	body := make([]byte, bodysize)
	io.ReadFull(input, body)
	msg = &Tdispatch {
		Contexts: contexts,
		Dest: string(dest),
		Dtabs: dtabs,
		Body: body,
	}
	return
}

func decodeRdispatch(input io.Reader, size int32) (msg *Rdispatch, err error) {
	var status int8
	var contexts []Header

	err = binary.Read(input, binary.BigEndian, &status)
	if err != nil {
		return
	}

	contexts, err = decodeHeaders(input)
	if err != nil {
		return
	}

	// The body size is the size minus status, and headersize
	bodysize := int(size) - 1 - headersSize(contexts)
	body := make([]byte, bodysize)
	_, err = io.ReadFull(input, body)
	if err != nil {
		return
	}

	// Successful message read.
	msg = &Rdispatch {
		Contexts: contexts,
		Body: body,
	}

	return
}