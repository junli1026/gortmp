package rtmp

import (
	"net"

	"github.com/junli1026/rtmp-server/logging"
	"github.com/junli1026/rtmp-server/message"
	"github.com/sirupsen/logrus"
)

type LogLevel int

const (
	PanicLevel LogLevel = iota
	FatalLevel
	ErrorLevel
	WarnLevel
	InfoLevel
	DebugLevel
	TraceLevel
)

var loglevelMap map[LogLevel]logrus.Level = map[LogLevel]logrus.Level{
	PanicLevel: logrus.PanicLevel,
	FatalLevel: logrus.FatalLevel,
	ErrorLevel: logrus.ErrorLevel,
	WarnLevel:  logrus.WarnLevel,
	InfoLevel:  logrus.InfoLevel,
	DebugLevel: logrus.DebugLevel,
	TraceLevel: logrus.TraceLevel,
}

//LogSetting is the setting for logger
type LogSetting struct {
	LogLevel   LogLevel
	Filename   string
	MaxSize    int
	MaxBackups int
	MaxAge     int
}

type rtmpServer struct {
	*baseServer
	flvHeaderCb     FlvCallback
	flvScriptDataCb FlvCallback
	flvVideoDataCb  FlvCallback
	flvAudioDataCb  FlvCallback
}

func NewServer(addr string) *rtmpServer {
	return newRtmpServer(addr)
}

func (s *rtmpServer) ConfigLog(setting *LogSetting) {
	config := &logging.LogConfig{
		LogLevel:   loglevelMap[setting.LogLevel],
		Filename:   setting.Filename,
		MaxSize:    setting.MaxSize,
		MaxBackups: setting.MaxBackups,
		MaxAge:     setting.MaxAge,
	}
	logging.ConfigLogger(config)
}

func (s *rtmpServer) Run() error {
	return s.run()
}

func (s *rtmpServer) Stop() {
	s.stop()
}

type FlvCallback func(meta *StreamMeta, timestamp uint32, data []byte) error

func newRtmpServer(addr string) *rtmpServer {
	s := &rtmpServer{}
	s.baseServer = newBaseServer(addr, s)
	return s
}

func (s *rtmpServer) OnFlvHeader(cb FlvCallback) *rtmpServer {
	s.flvHeaderCb = cb
	return s
}

func (s *rtmpServer) OnFlvScriptData(cb FlvCallback) *rtmpServer {
	s.flvScriptDataCb = cb
	return s
}

func (s *rtmpServer) OnFlvVideoData(cb FlvCallback) *rtmpServer {
	s.flvVideoDataCb = cb
	return s
}

func (s *rtmpServer) OnFlvAudioData(cb FlvCallback) *rtmpServer {
	s.flvAudioDataCb = cb
	return s
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
