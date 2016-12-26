package mux

import (
	"io"
	"math"
	"errors"
	"fmt"
	"encoding/binary"
)

// A part of a larger frame which has been broken up to alleviate head of line blocking
type Fragment struct {
	data []byte
}

// A simple key value pair
type Header struct {
	Key   []byte
	Value []byte
}

////////////////////////////
// decoding methods
////////////////////////////

func decodeHeaders(input io.Reader) ([]Header, error) {
	var count int16

	err := binary.Read(input, binary.BigEndian, &count)
	if err != nil {
		return []Header{}, err
	}

	headers := make([]Header, 0, count)

	for i := int16(0); i < count; i++ {
		header, err := decodeHeader(input)
		if err != nil {
			return headers, err
		}

		headers = append(headers, header)
	}

	return headers, nil
}

func decodeHeader(input io.Reader) (Header, error) {
	key, err := decodeInt16Bytes(input)
	if err != nil {
		return Header{}, err
	}

	value, err := decodeInt16Bytes(input)
	if err != nil {
		return Header{}, err
	}

	header := Header {
		Key: key,
		Value: value,
	}

	return header, nil
}

////////////////////////////
// encoding methods
////////////////////////////

func encodeHeader(header *Header, writer io.Writer) error {
	err := encodeBytesInt16(writer, header.Key)

	if err != nil {
		return err
	}

	return encodeBytesInt16(writer, header.Value)
}

func encodeHeaders(writer io.Writer, headers []Header) (err error) {
	// write the count
	l := len(headers)

	if l > math.MaxInt16 {
		return errors.New(fmt.Sprintf("Too many headers: %d", l))
	}

	int16length := int16(l)

	err = binary.Write(writer, binary.BigEndian, &int16length)

	if err != nil {
		return
	}

	// write each header
	for _, header := range headers {
		err = encodeHeader(&header, writer)
		if err != nil {
			return
		}
	}

	return nil
}

func headersSize(headers []Header) int {
	acc := 2 // int16 for header count

	for _, h := range headers {
		acc += 4 // int16 for key and value
		acc += len(h.Key)
		acc += len(h.Value)
	}

	return acc
}