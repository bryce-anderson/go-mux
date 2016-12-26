package mux

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"math"
	"io/ioutil"
)

const MaxStreamId = 0x007fffff	// 23 bits
const FragmentMask = 1 << 23	// 24th bit signals fragment

const TdispatchTpe = 2
const RdispatchTpe = -2

const TpingTpe = 65
const RpingTpe = -65

const TinitTpe = 68
const RinitTpe = -68

const BadRerrTpe = 127	// Old implementation fluke... Two's complement and all...
const RerrTpe = -128

type Frame struct {
	frameType 	int8
	isFragment	bool
	streamId  	int32
	message   	Message
}

func isFrameFragment(streamId int32) bool {
	return (streamId & FragmentMask) != 0
}

func EncodeFrame(writer io.Writer, frame *Frame) error {
	// Write the message size as a prefix
	size := frame.message.Size() + 4

	err := binary.Write(writer, binary.BigEndian, int32(size))
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
	Dest 	 string
	Dtabs    []Header
	Body     []byte
}

type Rdispatch struct {
	Status   int8
	Contexts []Header
	Body     []byte
}

type Tinit struct {
	version int16
	headers []Header
}

func (t *Tinit) Size() int {
	return initSize(t.headers)
}

func (t *Tinit) Type() int8 {
	return TinitTpe
}

func initSize(headers []Header) int {
	acc := 2	// 16 bit version
	for _, h := range headers {
		acc += 8	// length fields
		acc += len(h.Key)
		acc += len(h.Value)
	}
	return acc
}

func (t *Tinit) Encode(writer io.Writer) error {
	return encodeInit(writer, t.version, t.headers)
}

type Rinit struct {
	version int16
	headers []Header
}

func (t *Rinit) Size() int {
	return initSize(t.headers)
}

func (t *Rinit) Type() int8 {
	return TinitTpe
}

func (t *Rinit) Encode(writer io.Writer) error {
	return encodeInit(writer, t.version, t.headers)
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

type Tping struct {}

type Rping struct {}

type Rerr struct {
	error string
}

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

func headersSize(headers []Header) int {
	acc := 2 // int16 for header count

	for _, h := range headers {
		acc += 4 // int16 for key and value
		acc += len(h.Key)
		acc += len(h.Value)
	}

	return acc
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

func encodeHeader(header *Header, writer io.Writer) error {
	err := encodeBytesInt16(writer, header.Key)

	if err != nil {
		return err
	}

	return encodeBytesInt16(writer, header.Value)
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

func DecodeFrame(input io.Reader) (Frame, error) {
	var size int32

	err := binary.Read(input, binary.BigEndian, &size)
	if err != nil {
		return Frame{}, err
	}

	// Slice a frame off the reader. The entire frame must be read or else we will be sad

	return decodeMessage(input, size)
}

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

	bodysize := int(size) - 4 - headersSize(contexts) - headersSize(dtabs)
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

	// The body size is the size minus streamId+tpe, status, and headersize
	bodysize := int(size) - 4 - 1 - headersSize(contexts)
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

func decodeMessage(in io.Reader, size int32) (frame Frame, err error) {
	limitReader := io.LimitReader(in, int64(size))

	var header int32

	err = binary.Read(limitReader, binary.BigEndian, &header)
	if err != nil {
		return
	}

	frame.frameType = int8((header >> 24) & 0xff) // most significant byte is the type
	frame.streamId = MaxStreamId & header
	frame.isFragment = isFrameFragment(header)	// only makes sense for dispatch frames...

	switch frame.frameType {
	// TODO: fragments are handled incorrectly
	case TdispatchTpe:
		frame.message, err = decodeTdispatch(limitReader, size)

	case RdispatchTpe:
		frame.message, err = decodeRdispatch(limitReader, size)

	case TpingTpe:
		frame.message = &Tping{}

	case RpingTpe:
		frame.message = &Rping{}

	case TinitTpe:
		var headers []Header
		var version int16
		version, headers, err = decodeInit(limitReader)
		frame.message = &Tinit{
			version: version,
			headers: headers,
		}

	case RinitTpe:
		var headers []Header
		var version int16
		version, headers, err = decodeInit(limitReader)
		frame.message = &Rinit{
			version: version,
			headers: headers,
		}

	case RerrTpe:
		fallthrough
	case BadRerrTpe:
		var bytes []byte
		bytes, err = ioutil.ReadAll(limitReader)
		frame.message = &Rerr{
			error: string(bytes),
		}

	default:
		err = errors.New(fmt.Sprintf("Found invalid frame type: %d", frame.frameType))
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

// Requires a limited stream.
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
