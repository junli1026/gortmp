package message

import (
	"encoding/binary"

	"github.com/junli1026/rtmp-server/logging"
	utils "github.com/junli1026/rtmp-server/utils"
)

type SetChunkSizeMessage struct {
	messageHeader
	ChunkSize int
}

func NewSetChunkSizeMessage(chunkSize int) *SetChunkSizeMessage {
	m := &SetChunkSizeMessage{}
	m.StreamID = 0
	m.ChunkStreamID = 2
	m.MsgType = 1
	m.ChunkSize = chunkSize
	return m
}

func (msg SetChunkSizeMessage) toRaw() (*RawMessage, error) {
	raw := &RawMessage{
		messageHeader: msg.messageHeader,
		Raw:           make([]byte, 4),
	}
	binary.BigEndian.PutUint32(raw.Raw, uint32(msg.ChunkSize))
	return raw, nil
}

func deserializeSetChunkSize(msg *RawMessage) (Message, error) {
	m := &SetChunkSizeMessage{}
	m.messageHeader = msg.messageHeader
	m.ChunkSize = int(utils.ReadUint32(msg.Raw[0:4]))
	logging.Logger.Info("chunk size ", m.ChunkSize)
	return m, nil
}
