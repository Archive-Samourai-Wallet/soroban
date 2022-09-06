package p2p

import (
	"encoding/json"
	"errors"
)

type Message struct {
	Context string
	Payload []byte
}

func NewMessage(context string, obj interface{}) (Message, error) {
	if len(context) == 0 {
		return Message{}, errors.New("invalid context")
	}
	if obj == nil {
		return Message{}, errors.New("invalid payload")
	}
	payload, err := json.Marshal(obj)
	if err != nil {
		return Message{}, err
	}

	return Message{
		Context: context,
		Payload: payload,
	}, nil
}

func MessageFromBytes(data []byte) (Message, error) {
	if len(data) == 0 {
		return Message{}, errors.New("invalid data")
	}

	var result Message
	err := json.Unmarshal(data, &result)
	if err != nil {
		return Message{}, errors.New("failed to Unmarshal Message")
	}

	return result, nil
}

func (p *Message) ToBytes() ([]byte, error) {
	if len(p.Context) == 0 {
		return nil, errors.New("invalid context")
	}
	if len(p.Payload) == 0 {
		return nil, errors.New("invalid payload")
	}

	return json.Marshal(&p)
}

func (p *Message) ParsePayload(obj interface{}) error {
	if obj == nil {
		return errors.New("invalid obj")
	}
	if len(p.Context) == 0 {
		return errors.New("invalid context")
	}
	if len(p.Payload) == 0 {
		return errors.New("invalid payload")
	}

	return json.Unmarshal(p.Payload, obj)
}
