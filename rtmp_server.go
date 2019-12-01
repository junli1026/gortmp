package rs

import (
	"net"

	"github.com/junli1026/rtmp-server/message"
)

type rtmpServer struct {
	*baseServer
	flvHeaderCb     FlvCallback
	flvScriptDataCb FlvCallback
	flvVideoDataCb  FlvCallback
	flvAudioDataCb  FlvCallback
}

type FlvCallback func(meta *streamMeta, timestamp uint32, data []byte) error

func newRtmpServer(addr string) *rtmpServer {
	s := &rtmpServer{}
	s.baseServer = newBaseServer(addr, s)
	return s
}

func (s *rtmpServer) OnFlvHeader(cb FlvCallback) {
	s.flvHeaderCb = cb
}

func (s *rtmpServer) OnFlvScriptData(cb FlvCallback) {
	s.flvScriptDataCb = cb
}

func (s *rtmpServer) OnFlvVideoData(cb FlvCallback) {
	s.flvVideoDataCb = cb
}

func (s *rtmpServer) OnFlvAudioData(cb FlvCallback) {
	s.flvAudioDataCb = cb
}

func (s *rtmpServer) newContext(conn net.Conn) interface{} {
	return newRtmpContext(s)
}

func (*rtmpServer) read(data []byte, context interface{}) (consumed int, reply []byte, err error) {
	ctx := context.(*rtmpContext)
	if !ctx.hs.done() {
		return ctx.hs.handshake(data)
	}

	var rawMessage *message.RawMessage = nil
	if rawMessage, consumed, err = ctx.chunkReader.read(data); err != nil {
		return 0, nil, err
	}
	ctx.updateReceived(uint32(consumed))

	if rawMessage == nil {
		return consumed, nil, nil
	}

	var msg message.Message
	if msg, err = message.Deserialize(rawMessage); err != nil {
		return 0, nil, err
	}

	var resp []message.Message
	resp, err = ctx.handle(msg)
	if err != nil {
		return 0, nil, err
	}
	for _, r := range resp {
		buf, err := message.Serialize(ctx.chunkReader.chunkSize, r)
		if err != nil {
			return 0, nil, err
		}
		reply = append(reply, buf...)
	}

	return consumed, reply, nil
}
