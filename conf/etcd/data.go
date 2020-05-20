package etcd

import (
	"github.com/halivor/goutil/conf"
	"go.etcd.io/etcd/mvcc/mvccpb"
)

type data struct {
	event conf.EventType
	kv    *mvccpb.KeyValue
}

func (x *data) Data() []byte {
	return x.kv.Value
}
func (x *data) Event() conf.EventType {
	return x.event
}

func (x *data) String() string {
	return string(x.kv.Value)
}
