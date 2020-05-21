package etcd

import (
	svc "github.com/halivor/goutil/service"
	"go.etcd.io/etcd/mvcc/mvccpb"
)

type data struct {
	event svc.EventType
	kv    *mvccpb.KeyValue
}

func (x *data) Data() []byte {
	return x.kv.Value
}
func (x *data) Event() svc.EventType {
	return x.event
}

func (x *data) String() string {
	return string(x.kv.Value)
}
