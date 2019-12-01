package message

import (
	"encoding/binary"

	"github.com/junli1026/rtmp-server/logging"
	utils "github.com/junli1026/rtmp-server/utils"
)

type AckWindowSizeMessage struct {
	messageHeader
	WindowSize int
}

func (msg AckWindowSizeMessage) toRaw() (*RawMessage, error) {
	raw := &RawMessage{
		messageHeader: msg.messageHeader,
		Raw:           make([]byte, 4),
	}
	binary.BigEndian.PutUint32(raw.Raw, uint32(msg.WindowSize))
	return raw, nil
}

func NewAckWindowSizeMessage(windowSize int) *AckWindowSizeMessage {
	m := &AckWindowSizeMessage{}
	m.StreamID = 0
	m.ChunkStreamID = 2
	m.WindowSize = windowSize
	m.MsgType = 5
	return m
}

func deserializeAckWindowSize(msg *RawMessage) (Message, error) {
	m := &AckWindowSizeMessage{}
	m.messageHeader = msg.messageHeader
	m.WindowSize = int(utils.ReadUint32(msg.Raw[0:4]))
	logging.Logger.Info("window size ", m.WindowSize)
	return m, nil
}
