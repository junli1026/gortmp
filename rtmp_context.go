package rtmp

import (
	"encoding/binary"
	"errors"
	"fmt"
	"strings"

	"github.com/junli1026/gortmp/logging"
	"github.com/junli1026/gortmp/message"
)

type rtmpContext struct {
	streams           []*StreamMeta
	app               string
	tcURL             string
	swfURL            string
	flashVer          string
	hs                *handshakeState
	windowSize        int
	chunkReader       *chunkReader
	createStreamCount int
	received          uint32

	flvHeaderWritten bool
	s                *RtmpServer
}

func newRtmpContext(s *RtmpServer) *rtmpContext {
	ctx := &rtmpContext{}
	ctx.hs = newHandshakeState()
	ctx.windowSize = 2500000
	ctx.chunkReader = newChunkReader()
	ctx.streams = make([]*StreamMeta, 0)
	ctx.s = s
	ctx.received = 0
	return ctx
}

func (ctx *rtmpContext) updateReceived(delta uint32) {
	ctx.received += delta
}

func (ctx *rtmpContext) handle(msg message.Message) (reply []message.Message, err error) {
	switch v := msg.(type) {
	case *message.SetChunkSizeMessage:
		ctx.chunkReader.setChunkSize(int(v.ChunkSize))
	case *message.AckWindowSizeMessage:
		ctx.windowSize = v.WindowSize
	case *message.Amf0CommandMessage:
		reply, err = ctx.handleCommand(v)
	case *message.Amf0DataMessage:
		reply, err = ctx.handleData(v)
	case *message.VideoMessage:
		reply, err = ctx.onVideoData(v)
	case *message.AudioMessage:
		reply, err = ctx.onAudioData(v)
	default:
		logging.Logger.Warnf("unhandled message, type: %v", msg.GetType())
	}
	if ctx.received >= uint32(ctx.windowSize) {
		ack := message.NewAcknowledgementMessage(uint32(ctx.received))
		ctx.received = 0
		reply = append([]message.Message{ack}, reply...)
	}
	return
}

func (ctx *rtmpContext) handleData(cmd *message.Amf0DataMessage) ([]message.Message, error) {
	if cmd.CommandName == "@setDataFrame" &&
		(cmd.CallbackName == "onMetaData" || cmd.CallbackName == "onmetadata") {
		return ctx.onMetaData(cmd)
	}
	return nil, nil
}

func (ctx *rtmpContext) handleCommand(cmd *message.Amf0CommandMessage) ([]message.Message, error) {
	switch cmd.Name {
	case "connect":
		return ctx.onConnect(cmd)
	case "publish":
		return ctx.onPublish(cmd)
	case "FCPublish":
		return ctx.onFCPublish(cmd)
	case "releaseStream":
		return ctx.emptyResult(cmd)
	case "createStream":
		return ctx.onCreateStream(cmd)
	default:
		return ctx.emptyResult(cmd)
	}
}

func (ctx *rtmpContext) onCreateStream(cmd *message.Amf0CommandMessage) ([]message.Message, error) {
	result := message.NewAmf0CommandMessage("_result", cmd.TransactionID)
	ctx.createStreamCount++
	result.AddOther(ctx.createStreamCount)
	return []message.Message{result}, nil
}

func (ctx *rtmpContext) onConnect(cmd *message.Amf0CommandMessage) ([]message.Message, error) {
	logging.Logger.Infof(
		"connect stream-id:%v objects:%v", cmd.GetStreamID(), cmd.CommandObject)
	kv := cmd.CommandObject.(map[string]interface{})
	if v, ok := kv["tcUrl"].(string); ok {
		ctx.tcURL = v
	}
	if v, ok := kv["swfUrl"].(string); ok {
		ctx.swfURL = v
	}
	if v, ok := kv["flashVer"].(string); ok {
		ctx.flashVer = v
	}

	reply := make([]message.Message, 0)
	reply = append(reply, message.NewAckWindowSizeMessage(ctx.windowSize))
	reply = append(reply, message.NewSetPeerBandwidthMessage(2500000, 2))
	reply = append(reply, message.NewSetChunkSizeMessage(ctx.chunkReader.chunkSize))

	result := message.NewAmf0CommandMessage("_result", cmd.TransactionID)
	result.SetCommandObject(map[string]interface{}{
		"rtmpVer":      "RS/1.0",
		"capabilities": 255,
		"mode":         1,
	})
	result.AddOther(map[string]interface{}{
		"level":          "status",
		"code":           "NetConnection.Connect.Success",
		"description":    "Connection succeeded.",
		"objectEncoding": 0,
	})
	reply = append(reply, result)

	onBWDone := message.NewAmf0CommandMessage("onBWDone", 0)
	reply = append(reply, onBWDone)
	return reply, nil
}

func (ctx *rtmpContext) findStream(streamID int) *StreamMeta {
	var stream *StreamMeta = nil
	for _, s := range ctx.streams {
		if s.streamID == streamID {
			stream = s
			break
		}
	}
	return stream
}

func (ctx *rtmpContext) onPublish(cmd *message.Amf0CommandMessage) ([]message.Message, error) {
	var publishingName, publishingType string
	if len(cmd.Others) < 2 {
		return nil, fmt.Errorf("invalid publish meesage %v", *cmd)
	}
	if v, ok := cmd.Others[0].(string); ok {
		publishingName = v
	}
	if v, ok := cmd.Others[1].(string); ok {
		publishingType = v
	}
	logging.Logger.Info("publish(\"", publishingName, "\")")
	if strings.ToLower(publishingType) != "live" {
		return nil, fmt.Errorf("Only support publishing type live, while get %v", publishingType)
	}

	/* set stream info */
	stream := ctx.findStream(cmd.StreamID)
	if stream == nil {
		stream = &StreamMeta{}
		stream.streamID = cmd.StreamID
		ctx.streams = append(ctx.streams, stream)
	}
	stream.streamName = publishingName

	/* prepare reply */
	result := message.NewAmf0CommandMessage("onStatus", 0)
	result.StreamID = cmd.StreamID
	result.AddOther(map[string]interface{}{
		"level":       "status",
		"code":        "NetStream.Publish.Start",
		"description": "publishing " + publishingName,
	})
	return []message.Message{result}, nil
}

func (ctx *rtmpContext) onFCPublish(cmd *message.Amf0CommandMessage) ([]message.Message, error) {
	var streamName string
	if v, ok := cmd.Others[0].(string); ok {
		streamName = v
	} else {
		return nil, errors.New("FCPublish stream name empty")
	}

	msg := message.NewAmf0CommandMessage("onFCPublish", 0)
	msg.AddOther(map[string]interface{}{
		"code":        "NetStream.Publish.Start",
		"description": streamName,
	})

	/* prepare reply */
	result := message.NewAmf0CommandMessage("_result", cmd.TransactionID)

	return []message.Message{msg, result}, nil
}

func (ctx *rtmpContext) emptyResult(cmd *message.Amf0CommandMessage) ([]message.Message, error) {
	result := message.NewAmf0CommandMessage("_result", cmd.TransactionID)
	return []message.Message{result}, nil
}

func (ctx *rtmpContext) onMetaData(cmd *message.Amf0DataMessage) ([]message.Message, error) {
	logging.Logger.Debugf("@setDataFrame %v", cmd.Parameters)
	stream := ctx.findStream(cmd.StreamID)
	if stream == nil {
		return nil, fmt.Errorf("failed to find stream with id %v", cmd.StreamID)
	}

	ctx.setStreamMeta(stream, cmd.Parameters)

	if !ctx.flvHeaderWritten && ctx.s.flvHeaderCb != nil {
		err := ctx.s.flvHeaderCb(stream, 0, []byte{
			'F', 'L', 'V', 0x01, 0x05, 0x00, 0x00, 0x00, 0x09,
			0x00, 0x00, 0x00, 0x00, // previous tag size
		})
		if err != nil {
			return nil, err
		}
		ctx.flvHeaderWritten = true
	}

	if ctx.s.flvScriptDataCb != nil {
		metaData := cmd.Raw[16:] // skip @setDataFrame
		scriptData := make([]byte, 0)
		l := make([]byte, 4)
		binary.BigEndian.PutUint32(l, uint32(len(metaData)))
		scriptData = append(scriptData,
			0x12,             //tag type script data
			l[1], l[2], l[3], //data size
			0x00, 0x00, 0x00, 0x00, //timestamp
			0x00, 0x00, 0x00, //stream id
		)
		scriptData = append(scriptData, metaData[:]...)
		binary.BigEndian.PutUint32(l, uint32(len(scriptData)))
		scriptData = append(scriptData, l[:]...)

		err := ctx.s.flvScriptDataCb(stream, 0, scriptData)
		if err != nil {
			return nil, err
		}
	}
	return nil, nil
}

func (ctx *rtmpContext) setStreamMeta(stream *StreamMeta, meta map[string]interface{}) {
	logging.Logger.Info(meta)
	stream.url = ctx.tcURL
	for key, value := range meta {
		switch key {
		case "width":
			if v, ok := value.(float64); ok {
				stream.width = int(v)
			}
		case "height":
			if v, ok := value.(float64); ok {
				stream.height = int(v)
			}
		case "videocodecid":
			if v, ok := value.(string); ok {
				stream.videoCodec = v
			}
			stream.hasVideo = true
		case "videodatarate":
			if v, ok := value.(float64); ok {
				stream.videoDataRate = int(v)
			}
		case "framerate":
			if v, ok := value.(float64); ok {
				stream.frameRate = int(v)
			}
		case "audiocodecid":
			if v, ok := value.(string); ok {
				stream.audioCodec = v
			}
			stream.hasAudio = true
		case "audiodatarate":
			if v, ok := value.(float64); ok {
				stream.audioDataRate = int(v)
			}
		case "audiosamplerate":
			if v, ok := value.(float64); ok {
				stream.audioSampleRate = int(v)
			}
		case "audiosamplesize":
			if v, ok := value.(float64); ok {
				stream.audioSampleSize = int(v)
			}
		case "audiochannels":
			if v, ok := value.(float64); ok {
				stream.audioChannels = int(v)
			}
		case "stereo":
			if v, ok := value.(bool); ok {
				stream.stereo = v
			}
		case "encoder":
			if v, ok := value.(string); ok {
				stream.encoder = v
			}
		}
	}
}

func (ctx *rtmpContext) onVideoData(msg *message.VideoMessage) ([]message.Message, error) {
	if ctx.s.flvVideoDataCb != nil {
		if err := ctx.onMediaData(msg.RawMessage, 0x09); err != nil {
			return nil, err
		}
	}
	return nil, nil
}

func (ctx *rtmpContext) onAudioData(msg *message.AudioMessage) ([]message.Message, error) {
	if ctx.s.flvAudioDataCb != nil {
		if err := ctx.onMediaData(msg.RawMessage, 0x08); err != nil {
			return nil, err
		}
	}
	return nil, nil
}

func (ctx *rtmpContext) onMediaData(msg message.RawMessage, tagTye byte) error {
	stream := ctx.findStream(msg.StreamID)
	if stream == nil {
		return fmt.Errorf("failed to find stream with id %v", msg.StreamID)
	}

	l := make([]byte, 4)
	binary.BigEndian.PutUint32(l, uint32(len(msg.Raw)))

	mediaData := make([]byte, 0)

	timestamp := make([]byte, 4)
	binary.BigEndian.PutUint32(timestamp, msg.Timestamp)
	mediaData = append(mediaData,
		tagTye,           //tag type script data
		l[1], l[2], l[3], //data size
		timestamp[1], timestamp[2], timestamp[3], timestamp[0], //timestamp
		0x00, 0x00, 0x00, //stream id
	)

	mediaData = append(mediaData, msg.Raw[:]...)
	binary.BigEndian.PutUint32(l, uint32(len(mediaData)))
	mediaData = append(mediaData, l[:]...) //previous tag size

	var err error
	if tagTye == 0x09 {
		err = ctx.s.flvVideoDataCb(stream, msg.Timestamp, mediaData)
	} else {
		err = ctx.s.flvAudioDataCb(stream, msg.Timestamp, mediaData)
	}
	if err != nil {
		return err
	}
	return nil
}
