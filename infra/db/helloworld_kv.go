package db

import (
	"encoding/json"
	"errors"
)

type HelloWorldKv struct {
	messages []string
}

var _ Repository = (*HelloWorldKv)(nil)

func NewHelloWorldKv() *HelloWorldKv {
	return &HelloWorldKv{}
}

func (kv *HelloWorldKv) GetName() string {
	return "HelloWorldKV"
}

func (kv *HelloWorldKv) Append(msg string) {
	kv.messages = append(kv.messages, msg)
}

func (kv *HelloWorldKv) GetAll() []string {
	return kv.messages
}

func (kv *HelloWorldKv) LoadSnap(bz []byte) error {
	if len(bz) == 0 {
		return nil
	}

	var restore []string
	if err := json.Unmarshal(bz, &restore); err != nil {
		return errors.New("can not unmarshal load snap")
	}

	kv.messages = restore
	return nil
}

func (kv *HelloWorldKv) CreateSnap() ([]byte, error) {
	if len(kv.messages) == 0 {
		return []byte{}, nil
	}

	bz, err := json.Marshal(kv.messages)
	if err != nil {
		return []byte{}, err
	}

	return bz, nil
}
