package consul

import (
	"github.com/halivor/goutil/conf"
)

type data struct {
	ev   conf.EventType
	data []byte
}

func (x *data) Event() conf.EventType {
	return x.ev
}

func (x *data) Data() []byte {
	return x.data
}

func (x *data) String() string {
	return ""
}
