package message

import (
	"encoding/binary"
	"fmt"

	"github.com/junli1026/gortmp/logging"
)

// RawMessage reprsents raw message
type RawMessage struct {
	messageHeader
	Raw []byte
}

// Message interface
type Message interface {
	GetStreamID() int
	GetType() byte
	GetChunkStreamID() int
	GetTimestamp() uint32
	toRaw() (*RawMessage, error)
}

type messageHeader struct {
	MsgType       byte
	StreamID      int
	ChunkStreamID int
	Timestamp     uint32
}

func (header messageHeader) GetStreamID() int {
	return header.StreamID
}

func (header messageHeader) GetType() byte {
	return header.MsgType
}

func (header messageHeader) GetChunkStreamID() int {
	return header.ChunkStreamID
}

func (header messageHeader) GetTimestamp() uint32 {
	return header.Timestamp
}

type deserializer func(*RawMessage) (Message, error)

var deserializerList [23]deserializer

func init() {
	deserializerList[1] = deserializeSetChunkSize
	deserializerList[3] = deserializeAcknowledgement
	deserializerList[5] = deserializeAckWindowSize
	deserializerList[6] = deserializeSetPeerBandwidth
	deserializerList[8] = deserializeAudioMessage
	deserializerList[9] = deserializeVideoMessage
	deserializerList[18] = deserializeDataMessage
	deserializerList[20] = deserializeAmf0CommandMessage
}

// Deserialize function deserialize message
func Deserialize(raw *RawMessage) (Message, error) {
	if int(raw.MsgType) >= len(deserializerList) {
		return nil, fmt.Errorf("msg type %v out of range", raw.MsgType)
	}

	f := deserializerList[int(raw.MsgType)]
	if f == nil {
		return nil, fmt.Errorf("deserializer for msg type %v not implemented yet", raw.MsgType)
	}
	return f(raw)
}

type nonRecognizedMessage struct {
	messageHeader
}

func (msg nonRecognizedMessage) toRaw() (*RawMessage, error) {
	raw := &RawMessage{
		messageHeader: msg.messageHeader,
		Raw:           make([]byte, 4),
	}
	return raw, nil
}

// Serialize serialize message to byte slice
func Serialize(chunkSize int, msg Message) ([]byte, error) {
	raw, err := msg.toRaw()
	if err != nil {
		return nil, err
	}
	if raw.ChunkStreamID >= 64 {
		return nil, fmt.Errorf("chunk stream id %v not supported", raw.ChunkStreamID)
	}

	body := make([]byte, 0)
	index := 0
	for len(raw.Raw[index:]) > chunkSize {
		logging.Logger.Debug("do chunking")
		body = append(body, raw.Raw[index:index+chunkSize]...)
		body = append(body, 0xC0|byte(raw.ChunkStreamID))
		index += chunkSize
	}
	if len(raw.Raw[index:]) > 0 {
		body = append(body, raw.Raw[index:]...)
	} else {
		body = body[0 : len(body)-1] //truncate trailing 0xC0|csid
	}

	header := make([]byte, 12)
	header[0] = byte(raw.ChunkStreamID)

	// TODO :timestamp

	// message length
	l := len(raw.Raw)
	header[4] = byte(l >> 16)
	header[5] = byte(l >> 8)
	header[6] = byte(l)

	// message tyoe
	header[7] = byte(raw.MsgType)

	// stream id
	binary.LittleEndian.PutUint32(header[8:12], uint32(raw.StreamID))

	return append(header, body...), nil
}
