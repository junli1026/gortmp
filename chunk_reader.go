package rtmp

import (
	"encoding/binary"
	"errors"
	"fmt"

	"github.com/junli1026/rtmp-server/logging"
	"github.com/junli1026/rtmp-server/message"
	utils "github.com/junli1026/rtmp-server/utils"
)

var chunkHeaderSize = [4]int{11, 7, 3}

type chunkStream struct {
	chunkStreamID int
	prev          *chunkHeader
	curr          *chunkHeader
	payload       []byte
	remain        int
}

type chunkReader struct {
	streams   map[int]*chunkStream
	chunkSize int
}

func newChunkReader() *chunkReader {
	return &chunkReader{
		streams:   make(map[int]*chunkStream),
		chunkSize: 128,
	}
}

func (r *chunkReader) setChunkSize(chunkSize int) {
	r.chunkSize = chunkSize
}

func (r *chunkReader) read(data []byte) (*message.RawMessage, int, error) {
	h, length, err := readHeader(data)
	if length == 0 || err != nil || h == nil {
		return nil, 0, err
	}

	var cs *chunkStream = nil
	if st, ok := r.streams[int(h.chunkStreamID)]; ok {
		cs = st
		cs.curr = h
	} else {
		cs = &chunkStream{
			chunkStreamID: int(h.chunkStreamID),
			prev:          nil,
			curr:          h,
			payload:       make([]byte, 0),
			remain:        0,
		}
		r.streams[int(h.chunkStreamID)] = cs
	}
	updateHeader(cs.curr, cs.prev)

	/* process chunk message payload */
	var consumed int
	sz := int(cs.curr.messageLength) - len(cs.payload)

	if sz > int(r.chunkSize) {
		sz = int(r.chunkSize)
	}
	if len(data[length:]) < sz {
		return nil, 0, nil
	}

	cs.payload = append(cs.payload, data[length:length+sz]...)
	cs.remain = int(cs.curr.messageLength) - len(cs.payload)
	cs.prev = cs.curr
	cs.curr = nil
	consumed = length + sz

	/* message is complete */
	if cs.remain == 0 {
		msg := &message.RawMessage{}
		msg.Raw = cs.payload
		msg.MsgType = cs.prev.typeID
		msg.StreamID = int(cs.prev.streamID)
		msg.ChunkStreamID = int(cs.prev.chunkStreamID)
		msg.Timestamp = cs.prev.timestamp
		cs.payload = make([]byte, 0)
		return msg, consumed, nil
	}
	return nil, consumed, nil
}

type chunkHeader struct {
	/* basic header */
	format        byte
	chunkStreamID uint32

	/* message header */
	timestamp      uint32
	typeID         byte
	streamID       uint32
	timestampDelta uint32
	messageLength  uint32
}

func updateHeader(curr *chunkHeader, prev *chunkHeader) error {
	if curr.format != 0 && prev == nil {
		return errors.New("first message fmt is not 0")
	}

	switch curr.format {
	case 1:
		curr.streamID = prev.streamID
		curr.timestamp = prev.timestamp + curr.timestampDelta
	case 2:
		curr.streamID = prev.streamID
		curr.messageLength = prev.messageLength
		curr.timestamp = prev.timestamp + curr.timestampDelta
		curr.typeID = prev.typeID
	case 3:
		if curr.chunkStreamID != prev.chunkStreamID {
			return fmt.Errorf("Unexpected chunk stream id %v, expect %v", curr.chunkStreamID, prev.chunkStreamID)
		}
		curr.streamID = prev.streamID
		curr.messageLength = prev.messageLength
		curr.timestamp = prev.timestamp
		curr.typeID = prev.typeID
	}

	logging.Logger.Debugf(
		"fmt:%v cs-id:%v msg-type:%v stream-id:%v timestamp:%v body-size:%v",
		curr.format,
		curr.chunkStreamID,
		curr.typeID,
		curr.streamID,
		curr.timestamp,
		curr.messageLength,
	)
	return nil
}

func readHeader(data []byte) (h *chunkHeader, length int, err error) {
	h = &chunkHeader{
		format:         0xFF,
		chunkStreamID:  0,
		timestamp:      0,
		typeID:         0,
		timestampDelta: 0,
		messageLength:  0,
	}

	length, err = h.readBasicHeader(data)
	if length == 0 || err != nil {
		return
	}

	readFunc := []func([]byte) (int, error){
		h.readMessageType0,
		h.readMessageType1,
		h.readMessageType2,
	}

	if h.format != 3 {
		l, err2 := readFunc[h.format](data[length:])
		if l == 0 || err2 != nil {
			return nil, l, err2
		}
		length += l
	}
	return
}

func (h *chunkHeader) readBasicHeader(data []byte) (int, error) {
	if len(data) == 0 {
		return 0, nil
	}

	h.format = data[0] >> 6
	if h.format > 3 {
		return 0, errors.New("invalid fmt value")
	}

	chunkStreamID := data[0] & 0x3F
	switch chunkStreamID {
	case 0x00:
		if len(data) < 2 {
			return 0, nil
		}
		h.chunkStreamID = uint32(data[1]) + 64
		return 2, nil
	case 0x3F:
		if len(data) < 3 {
			return 0, nil
		}
		h.chunkStreamID = uint32(data[2])*256 + uint32(data[1])
		return 3, nil
	default:
		h.chunkStreamID = uint32(chunkStreamID)
		return 1, nil
	}
}

func (h *chunkHeader) readMessageType0(data []byte) (int, error) {
	if len(data) < chunkHeaderSize[0] {
		return 0, nil
	}
	h.timestamp = utils.ReadUint32(data[0:3])
	h.messageLength = utils.ReadUint32(data[3:6])

	h.typeID = data[6]
	h.streamID = binary.LittleEndian.Uint32(data[7:11])

	// read extended timestamp
	if h.timestamp >= 0xFFFFFF {
		if len(data) < chunkHeaderSize[0]+4 {
			return 0, nil
		}
		h.timestamp = utils.ReadUint32(data[11:15])
		return chunkHeaderSize[0] + 4, nil
	}
	return chunkHeaderSize[0], nil
}

func (h *chunkHeader) readMessageType1(data []byte) (int, error) {
	if len(data) < chunkHeaderSize[1] {
		return 0, nil
	}

	// message stream id is kept the same as last one
	h.timestampDelta = utils.ReadUint32(data[0:3])
	h.messageLength = utils.ReadUint32(data[3:6])
	h.typeID = data[6]

	// read extended timestamp delta
	if h.timestampDelta >= 0xFFFFFF {
		if len(data) < chunkHeaderSize[1]+4 {
			return 0, nil
		}
		h.timestampDelta = utils.ReadUint32(data[7:11])
		return chunkHeaderSize[1] + 4, nil
	}
	return chunkHeaderSize[1], nil
}

func (h *chunkHeader) readMessageType2(data []byte) (int, error) {
	if len(data) < chunkHeaderSize[2] {
		return 0, nil
	}
	h.timestampDelta = utils.ReadUint32(data[0:3])

	// read extended timestamp delta
	if h.timestampDelta >= 0xFFFFFF {
		if len(data) < chunkHeaderSize[2]+4 {
			return 0, nil
		}
		h.timestampDelta = utils.ReadUint32(data[3:7])
		return chunkHeaderSize[2] + 4, nil
	}
	return chunkHeaderSize[2], nil
}
