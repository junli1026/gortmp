package message

type AudioMessage struct {
	RawMessage
}

func (msg AudioMessage) toRaw() (*RawMessage, error) {
	return &msg.RawMessage, nil
}

func deserializeAudioMessage(msg *RawMessage) (Message, error) {
	return &AudioMessage{
		RawMessage: *msg,
	}, nil
}
