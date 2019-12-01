package message

type Amf0DataMessage struct {
	messageHeader
	CommandName  string
	CallbackName string
	Parameters   map[string]interface{}
	Raw          []byte
}

func (msg Amf0DataMessage) toRaw() (*RawMessage, error) {
	raw := &RawMessage{}
	raw.messageHeader = msg.messageHeader
	raw.Raw = msg.Raw
	return raw, nil
}

func deserializeDataMessage(msg *RawMessage) (Message, error) {
	arr, err := deserializeAMF0(msg.Raw)
	if err != nil {
		return nil, err
	}
	m := &Amf0DataMessage{}
	m.messageHeader = msg.messageHeader
	if len(arr) >= 1 {
		m.CommandName = arr[0].(string)
	}
	if len(arr) >= 2 {
		m.CallbackName = arr[1].(string)
	}
	if len(arr) >= 3 {
		m.Parameters = arr[2].(map[string]interface{})
	}
	m.Raw = msg.Raw
	return m, nil
}
