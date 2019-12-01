package message

import (
	"encoding/binary"

	"github.com/junli1026/rtmp-server/logging"
	utils "github.com/junli1026/rtmp-server/utils"
)

type AcknowledgementMessage struct {
	messageHeader
	Sequence uint32
}

func (msg AcknowledgementMessage) toRaw() (*RawMessage, error) {
	raw := &RawMessage{
		messageHeader: msg.messageHeader,
		Raw:           make([]byte, 4),
	}
	binary.BigEndian.PutUint32(raw.Raw, uint32(msg.Sequence))
	return raw, nil
}

func NewAcknowledgementMessage(seq uint32) *AcknowledgementMessage {
	m := &AcknowledgementMessage{}
	m.StreamID = 0
	m.ChunkStreamID = 2
	m.Sequence = seq
	m.MsgType = 3
	return m
}

func deserializeAcknowledgement(msg *RawMessage) (Message, error) {
	m := &AcknowledgementMessage{}
	m.messageHeader = msg.messageHeader
	m.Sequence = utils.ReadUint32(msg.Raw[0:4])
	logging.Logger.Info(" size ", m.Sequence)
	return m, nil
}
