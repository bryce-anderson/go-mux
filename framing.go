package mux

import (
	"io"
	"encoding/binary"
)

// TODO: this doesn't seem like it should be exported
type Frame struct {
	streamId  	int32
	message   	Message
}

func EncodeFrame(writer io.Writer, frame *Frame) error {
	err := encodeFrameHeader(writer, frame)
	if err != nil {
		return err
	}
	return frame.message.Encode(writer)
}

func DecodeFrame(input io.Reader) (Frame, error) {
	var size int32

	err := binary.Read(input, binary.BigEndian, &size)
	if err != nil {
		return Frame{}, err
	}

	// Slice a frame off the reader. The entire frame must be read or else we will be sad
	return decodeFrame(input, size)
}

func encodeFrameHeader(writer io.Writer, frame *Frame) (err error) {
	// Write the message size as a prefix
	size := frame.message.Size() + 4

	err = binary.Write(writer, binary.BigEndian, int32(size))
	if err != nil {
		return err
	}

	header := int32(frame.message.Type()) << 24 | frame.streamId
	err = binary.Write(writer, binary.BigEndian, header)
	return
}

