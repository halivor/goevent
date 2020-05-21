package consul

import (
	svc "github.com/halivor/goutil/service"
)

type data struct {
	ev   svc.EventType
	data []byte
}

func (x *data) Event() svc.EventType {
	return x.ev
}

func (x *data) Data() []byte {
	return x.data
}

func (x *data) String() string {
	return ""
}
