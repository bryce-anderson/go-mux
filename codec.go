package mux

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
)

const MaxStreamId = 0x007fffff	// 23 bits
const FragmentMask = 1 << 23	// 24th bit signals fragment

const TdispatchTpe = 2
const RdispatchTpe = -2

const TpingTpe = 65
const RpingTpe = -65

type Frame struct {
	frameType int8
	streamId  int32
	message   Message
}

func EncodeFrame(frame *Frame, writer io.Writer) error {
	// Write the message size as a prefix
	size := frame.message.Size() + 4

	err := binary.Write(writer, binary.BigEndian, size)
	if err != nil {
		return err
	}

	header := int32(frame.frameType)<<24 | frame.streamId
	err = binary.Write(writer, binary.BigEndian, header)
	if err != nil {
		return err
	}

	return frame.message.Encode(writer)
}

type Message interface {
	Type() int8                   // Message type
	Size() int                    // Message size not counting type tag or length prefix
	Encode(write io.Writer) error // Encode message into the stream. Doesn't include the type type, streamid, or frame length prefix
}

type Header struct {
	Key   []byte
	Value []byte
}

type Tdispatch struct {
	Contexts []Header
	Dtabs    []Header
	Body     []byte
}

type Rdispatch struct {
	Contexts []Header
	Body     []byte
}

type Tping struct {}

type Rping struct {}

func (t *Tping) Type() int8 {
	return TpingTpe
}

func (_ *Tping) Size() int {
	return 0
}

func (_ *Tping) Encode(write io.Writer) error {
	return nil
}

func (_ *Rping) Encode(write io.Writer) error {
	return nil
}

func (_ *Rping) Type() int8 {
	return RpingTpe
}

func (_ *Rping) Size() int {
	return 0
}

func (r *Rdispatch) Type() int8 {
	return RdispatchTpe
}

func (r *Rdispatch) Size() int {
	acc := headersSize(r.Contexts)
	acc += len(r.Body)
	return acc
}

func (r *Rdispatch) Encode(writer io.Writer) (err error) {
	// encode headers
	err = encodeHeaders(r.Contexts, writer)

	if err != nil {
		return
	}

	_, err = writer.Write(r.Body)

	return
}

func (t *Tdispatch) Size() int {
	acc := headersSize(t.Contexts)
	acc += headersSize(t.Dtabs)
	acc += len(t.Body)
	return acc
}

func (t *Tdispatch) Type() int8 {
	return TdispatchTpe
}

func (t *Tdispatch) Encode(writer io.Writer) (err error) {
	err = encodeHeaders(t.Contexts, writer)
	if err != nil {
		return
	}

	err = encodeHeaders(t.Dtabs, writer)
	if err != nil {
		return
	}

	_, err = writer.Write(t.Body)

	return
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

func encodeHeaders(headers []Header, writer io.Writer) (err error) {
	// write the count
	l := len(headers)

	if l > binary.MaxVarintLen16 {
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

func encodeHeader(header *Header, writer io.Writer) error {
	err := encodeBytesInt16(header.Key, writer)

	if err != nil {
		return err
	}

	return encodeBytesInt16(header.Value, writer)
}

func encodeBytesInt16(bytes []byte, writer io.Writer) (err error) {
	length := len(bytes)

	if length > binary.MaxVarintLen16 {
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

func DecodeFrame(input io.Reader) (Frame, error) {
	var size int32

	err := binary.Read(input, binary.LittleEndian, &size)
	if err != nil {
		return Frame{}, err
	}

	// Slice a frame off the reader. The entire frame must be read or else

	limitReader := io.LimitReader(input, int64(size))

	tpe, stream, msg, err := decodeMessage(limitReader, size)
	if err != nil {
		return Frame{}, err
	}

	frame := Frame{
		frameType: tpe,
		streamId:  stream,
		message:   msg,
	}

	return frame, nil
}

func decodeMessage(input io.Reader, size int32) (tpe int8, stream int32, msg Message, err error) {
	var header int32

	err = binary.Read(input, binary.LittleEndian, &header)
	if err != nil {
		return
	}

	tpe = int8((header >> 24) & 0xff) // most significant byte is the type
	stream = 0x00ffffff & header

	switch tpe {
	case TdispatchTpe:
		contexts, _err := decodeHeaders(input)
		if _err != nil {
			err = _err
			return
		}
		dtabs, _err := decodeHeaders(input)
		if _err != nil {
			err = _err
			return
		}

		bodysize := int(size) - 4 - headersSize(contexts) - headersSize(dtabs)
		body := make([]byte, bodysize)
		io.ReadFull(input, body)
		msg = &Tdispatch {
			Contexts: contexts,
			Dtabs: dtabs,
			Body: body,
		}

		return

	case RdispatchTpe:
		contexts, _err := decodeHeaders(input)
		if _err != nil {
			err = _err
			return
		}

		bodysize := int(size) - 4 - headersSize(contexts)
		body := make([]byte, bodysize)
		io.ReadFull(input, body)
		msg = &Rdispatch {
			Contexts: contexts,
			Body: body,
		}

		return

	case TpingTpe:
		msg = &Tping{}

	case RpingTpe:
		msg = &Rping{}


	default:
		err = errors.New(fmt.Sprintf("Found invalid frame type: %d", tpe))
		return
	}

	return
}

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
