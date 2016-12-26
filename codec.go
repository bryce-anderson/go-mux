package mux

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"math"
	"io/ioutil"
	"sync"
)

type codec interface {
	// Run serially. Reads a single frame from the wire
	decodeFrame(in io.Reader) (Frame, error)

	// May be run in parallel. The types of Messages in frames are determined by the protocol version
	encodeFrame(out io.Writer, frame *Frame) error
}

// Base codec implementation. Doesn't support message fragmenting
type baseCodec struct {
	writeLock sync.Mutex
}

// simple decoder function
func (c *baseCodec) decodeFrame(in io.Reader) (Frame, error) {
	return DecodeFrame(in)
}

// simple encoder function
func (c *baseCodec) encodeFrame(out io.Writer, frame *Frame) error {
	c.writeLock.Lock()
	defer c.writeLock.Unlock()
	return EncodeFrame(out, frame)
}

func isFrameFragment(streamId int32) bool {
	return (streamId & FragmentMask) != 0
}

func encodeBytesInt16(writer io.Writer, bytes []byte) (err error) {
	length := len(bytes)

	if length > math.MaxInt16 {
		return errors.New(fmt.Sprintf("Context field overflow. Length: %d", length))
	}

	i16len := int16(length)

	err = binary.Write(writer, binary.BigEndian, &i16len)
	if err != nil {
		return
	}

	_, err = writer.Write(bytes)

	return
}

func decodeFrame(in io.Reader, size int32) (frame Frame, err error) {
	limitReader := io.LimitReader(in, int64(size))

	// Read the header of the frame
	var header int32
	err = binary.Read(limitReader, binary.BigEndian, &header)
	if err != nil {
		return
	}

	frameTpe := int8((header >> 24) & 0xff) // most significant byte is the type
	frame.streamId = MaxStreamId & header

	// subtract 4 bytes for the header
	frame.message, err = decodeStandardFrame(limitReader, frameTpe, size - 4)
	return
}

// Expects the reader to signal EOF at the end of the frame
func decodeStandardFrame(in io.Reader, tpe int8, size int32) (msg Message, err error) {
	switch tpe {
	// TODO: fragments are handled incorrectly
	case TdispatchTpe:
		msg, err = decodeTdispatch(in, size)

	case RdispatchTpe:
		msg, err = decodeRdispatch(in, size)

	case TpingTpe:
		msg = &Tping{}

	case RpingTpe:
		msg = &Rping{}

	case TinitTpe:
		var headers []Header
		var version int16
		version, headers, err = decodeInit(in)
		msg = &Tinit{
			version: version,
			headers: headers,
		}

	case RinitTpe:
		var headers []Header
		var version int16
		version, headers, err = decodeInit(in)
		msg = &Rinit{
			version: version,
			headers: headers,
		}

	case RerrTpe:
		fallthrough
	case BadRerrTpe:
		var bytes []byte
		bytes, err = ioutil.ReadAll(in)
		msg = &Rerr{
			error: string(bytes),
		}

	default:
		err = errors.New(fmt.Sprintf("Found invalid frame type: %d", tpe))
		return
	}

	return
}

func readInt32Slice(input io.Reader) (data []byte, err error) {
	var fieldLen int32
	err = binary.Read(input, binary.BigEndian, &fieldLen)
	if err != nil {
		return
	}
	data = make([]byte, int(fieldLen), int(fieldLen))
	_, err = io.ReadFull(input, data)
	return
}

func decodeInt16Bytes(input io.Reader) ([]byte, error) {
	var len int16
	err := binary.Read(input, binary.BigEndian, len)
	if err != nil {
		return []byte{}, err
	}

	bytes := make([]byte, len)
	_, err = io.ReadFull(input, bytes)

	return bytes, err
}
