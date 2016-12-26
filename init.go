package mux

import (
	"io"
	"encoding/binary"
)

// Handshake message sent by the client to the server
type Tinit struct {
	version int16
	headers []Header
}

// Handshake response sent by the server to the client
type Rinit struct {
	version int16
	headers []Header
}

////////////////////////////
// Tinit Message methods
////////////////////////////

func (t *Tinit) Size() int {
	return initSize(t.headers)
}

func (t *Tinit) Type() int8 {
	return TinitTpe
}

func (t *Tinit) Encode(writer io.Writer) error {
	return encodeInit(writer, t.version, t.headers)
}

////////////////////////////
// Rinit Message methods
////////////////////////////

func (t *Rinit) Size() int {
	return initSize(t.headers)
}

func (t *Rinit) Type() int8 {
	return TinitTpe
}

func (t *Rinit) Encode(writer io.Writer) error {
	return encodeInit(writer, t.version, t.headers)
}

////////////////////////////
// init decoding methods
////////////////////////////

// Requires a limited stream since EOF determines the end of the init frame
func decodeInit(input io.Reader) (version int16, headers []Header, err error) {
	err = binary.Read(input, binary.BigEndian, &version)
	if err != nil {
		return
	}

	for {
		var header Header
		header.Key, err = readInt32Slice(input)
		if err != nil{
			if err == io.EOF { // not really an error
				err = nil
			}
			return
		}

		header.Value, err = readInt32Slice(input)
		if err != nil {
			return
		}


		headers = append(headers, header)
	}
}

////////////////////////////
// Helper methods
////////////////////////////

func initSize(headers []Header) int {
	acc := 2	// 16 bit version
	for _, h := range headers {
		acc += 8	// length fields
		acc += len(h.Key)
		acc += len(h.Value)
	}
	return acc
}

func encodeInit(writer io.Writer, version int16, headers []Header) (err error) {
	err = binary.Write(writer, binary.BigEndian, version)
	if err != nil {
		return
	}

	for _, h := range headers {
		err = binary.Write(writer, binary.BigEndian, int32(len(h.Key)))
		if err != nil {
			return
		}
		_, err = writer.Write(h.Key)
		if err != nil {
			return
		}

		err = binary.Write(writer, binary.BigEndian, int32(len(h.Value)))
		if err != nil {
			return
		}
		_, err = writer.Write(h.Value)
		if err != nil {
			return
		}
	}

	// success
	return
}