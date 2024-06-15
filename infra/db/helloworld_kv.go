package db

import (
	"encoding/json"
	"errors"
)

type HelloWorldKv struct {
	messages []string
}

func NewHelloWorldKv() *HelloWorldKv {
	return &HelloWorldKv{}
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

func (kv *HelloWorldKv) CreateSnap() []byte {
	if len(kv.messages) == 0 {
		return nil
	}

	bz, _ := json.Marshal(kv.messages)
	return bz
}
