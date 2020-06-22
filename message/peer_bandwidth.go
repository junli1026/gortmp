package message

import (
	"encoding/binary"

	"github.com/junli1026/gortmp/logging"
	utils "github.com/junli1026/gortmp/utils"
)

type SetPeerBandwidthMessage struct {
	messageHeader
	ackWindowSize uint32
	limitType     byte
}

func NewSetPeerBandwidthMessage(ackWindowSize uint32, limitType byte) *SetPeerBandwidthMessage {
	m := &SetPeerBandwidthMessage{}
	m.StreamID = 0
	m.ChunkStreamID = 2
	m.MsgType = 6
	m.ackWindowSize = ackWindowSize
	m.limitType = limitType
	return m
}

func (msg SetPeerBandwidthMessage) toRaw() (*RawMessage, error) {
	raw := &RawMessage{
		messageHeader: msg.messageHeader,
		Raw:           make([]byte, 5),
	}
	binary.BigEndian.PutUint32(raw.Raw, uint32(msg.ackWindowSize))
	raw.Raw[4] = msg.limitType
	return raw, nil
}

func deserializeSetPeerBandwidth(msg *RawMessage) (Message, error) {
	m := &SetPeerBandwidthMessage{}
	m.messageHeader = msg.messageHeader
	m.ackWindowSize = utils.ReadUint32(msg.Raw[0:4])
	m.limitType = msg.Raw[4]
	logging.Logger.Info("window size ", m.ackWindowSize)
	return m, nil
}
