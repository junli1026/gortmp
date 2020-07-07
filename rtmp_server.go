package rtmp

import (
	"net"

	"github.com/junli1026/gortmp/logging"
	"github.com/junli1026/gortmp/message"
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

type StreamDataHandler func(meta *StreamMeta, data *StreamData) error

type StreamCloseHandler func(meta *StreamMeta, err error)

type StreamDataType int

const (
	FlvHeader StreamDataType = iota
	FlvScript
	FlvVideo
	FlvAudio
)

type StreamData struct {
	Type      StreamDataType
	Timestamp uint32
	Data      []byte
}

type RtmpServer struct {
	*baseServer
	streamDataHandler  StreamDataHandler
	streamCloseHandler StreamCloseHandler
}

func NewServer() *RtmpServer {
	s := newRtmpServer()
	return s
}

func (s *RtmpServer) ConfigLog(setting *LogSetting) {
	config := &logging.LogConfig{
		LogLevel:   loglevelMap[setting.LogLevel],
		Filename:   setting.Filename,
		MaxSize:    setting.MaxSize,
		MaxBackups: setting.MaxBackups,
		MaxAge:     setting.MaxAge,
	}
	logging.ConfigLogger(config)
}

func (s *RtmpServer) Run(addr string) error {
	return s.baseServer.listenAndServe(addr)
}

func (s *RtmpServer) Stop() {
	s.stop()
}

func newRtmpServer() *RtmpServer {
	s := &RtmpServer{}
	s.baseServer = newBaseServer(s)
	return s
}

func (s *RtmpServer) OnStreamData(handler StreamDataHandler) {
	s.streamDataHandler = handler
}

func (s *RtmpServer) OnStreamClose(handler StreamCloseHandler) {
	s.streamCloseHandler = handler
}

func (s *RtmpServer) newContext(conn net.Conn) interface{} {
	return newRtmpContext(s)
}

func (*RtmpServer) read(data []byte, context interface{}) (consumed int, reply []byte, err error) {
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

func (s *RtmpServer) close(err error, context interface{}) {
	ctx := context.(*rtmpContext)
	for _, stream := range ctx.streams {
		s.streamCloseHandler(stream, err)
	}
}
