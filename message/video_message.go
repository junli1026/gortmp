package message

type VideoMessage struct {
	RawMessage
}

func (msg VideoMessage) toRaw() (*RawMessage, error) {
	return &msg.RawMessage, nil
}

func deserializeVideoMessage(msg *RawMessage) (Message, error) {
	return &VideoMessage{
		RawMessage: *msg,
	}, nil
}
