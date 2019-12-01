package message

import (
	"errors"
	"fmt"
	"reflect"
)

type Amf0CommandMessage struct {
	messageHeader
	Name          string
	TransactionID int
	CommandObject interface{}
	Others        []interface{}
}

// NewAmf0CommandMessage create a new instance of Amf0CommandMessage
func NewAmf0CommandMessage(name string, transactionID int) *Amf0CommandMessage {
	msg := &Amf0CommandMessage{
		Name:          name,
		TransactionID: transactionID,
		CommandObject: nil,
		Others:        nil,
	}
	msg.MsgType = 20
	msg.StreamID = 0
	msg.ChunkStreamID = 3
	return msg
}

// SetCommandObject set command object
func (msg *Amf0CommandMessage) SetCommandObject(o interface{}) {
	msg.CommandObject = o
}

// AddOther add object to others
func (msg *Amf0CommandMessage) AddOther(other interface{}) {
	msg.Others = append(msg.Others, other)
}

func (msg Amf0CommandMessage) toRaw() (raw *RawMessage, err error) {
	raw = &RawMessage{
		messageHeader: msg.messageHeader,
	}

	objects := []interface{}{msg.Name, msg.TransactionID, msg.CommandObject}
	if len(msg.Others) > 0 {
		objects = append(objects, msg.Others...)
	}
	if raw.Raw, err = serializeAMF0(objects); err != nil {
		return nil, err
	}
	return raw, nil
}

func getCommandName(arr []interface{}) (string, error) {
	if len(arr) < 1 {
		return "", errors.New("missing command name")
	}
	var cmdname string
	var ok bool
	if cmdname, ok = arr[0].(string); !ok {
		return "", fmt.Errorf("expect string as command name, while get %v", reflect.TypeOf(arr[0]))
	}
	return cmdname, nil
}

func getTransactionID(arr []interface{}) (int, error) {
	if len(arr) < 2 {
		return 0, errors.New("missing transaction id")
	}
	var transactionID float64
	var ok bool
	if transactionID, ok = arr[1].(float64); !ok {
		return 0, fmt.Errorf("expect number as transaction id, while get %v", reflect.TypeOf(arr[1]))
	}
	return int(transactionID), nil
}

func validateCommand(arr []interface{}) (bool, error) {
	if len(arr) < 3 {
		return false, fmt.Errorf("invalid message %v", arr)
	}
	if _, ok := arr[0].(string); !ok {
		return false, fmt.Errorf("invalid message %v, expect string as command name, while get %v", arr, reflect.TypeOf(arr[0]))
	}
	if _, ok := arr[1].(float64); !ok {
		return false, fmt.Errorf("invalid message %v, expect number as transaction id, while get %v", arr, reflect.TypeOf(arr[1]))
	}
	return true, nil
}

func deserializeAmf0CommandMessage(raw *RawMessage) (Message, error) {
	arr, err := deserializeAMF0(raw.Raw)
	if valid, err := validateCommand(arr); !valid {
		return nil, err
	}

	m := &Amf0CommandMessage{}
	m.messageHeader = raw.messageHeader
	m.Name = arr[0].(string)
	m.TransactionID = int(arr[1].(float64))
	m.CommandObject = arr[2]
	m.Others = arr[3:]
	return m, err
}
