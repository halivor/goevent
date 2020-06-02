package etcd

import (
	"github.com/golang/protobuf/proto"
	ce "github.com/halivor/common/golang/util/errno"
	svc "github.com/halivor/goutil/service"
)

func (c *conn) NewClnt(ptr interface{}, module, typ, addr string) {
	nclnt := &svc.Client{This: ptr, Module: module, Type: typ, Addr: addr}
	if _, ok := c.mcs[module]; !ok {
		c.mcs[module] = make(map[*svc.Client]struct{}, 8)
	}
	for clnt, _ := range c.mcs[module] {
		if clnt.Addr == nclnt.Addr {
			// TODO: 处理历史连接
			break
		}
	}
	c.mcs[module][nclnt] = struct{}{}
}

func (c *conn) CallR(module, method string, req, rsp proto.Message) ce.Errno {
	return ce.SUCC
}
