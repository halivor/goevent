package etcd

import (
	"reflect"

	"github.com/golang/protobuf/proto"
	cp "github.com/halivor/common/golang/packet"
	ce "github.com/halivor/common/golang/util/errno"
	svc "github.com/halivor/goutil/service"
)

func (c *conn) NewClnt(srv svc.Server, ptr interface{}) {
	service := srv.GetService()
	nclnt := &svc.Client{
		Service:   service,
		Interface: srv.GetInterface(),
		Type:      srv.GetType(),
		Host:      srv.GetHost(),
		This:      ptr,
	}
	if _, ok := c.mcs[service]; !ok {
		c.mcs[service] = make(map[*svc.Client]struct{}, 8)
	}
	for clnt, _ := range c.mcs[service] {
		if clnt.Host == nclnt.Host && clnt.Type == nclnt.Type {
			// TODO: 处理历史连接
			break
		}
	}
	c.mcs[service][nclnt] = struct{}{}
}

func (c *conn) Call(srv, service string, req, rsp proto.Message) ce.Errno {
	if _, ok := c.mcs[srv]; !ok {
		return ce.SRV_ERR
	}
	for clnt, _ := range c.mcs[srv] {
		rets := reflect.ValueOf(clnt.This).MethodByName(service).
			Call([]reflect.Value{reflect.ValueOf(cp.NewRequest(req))})
		for _, ret := range rets {
			switch v := ret.Interface().(type) {
			case *cp.Response:
				if en := ce.Errno(v.GetErrno()); en != ce.SUCC {
					return en
				}
				proto.Unmarshal(v.GetBody(), rsp)
			case error:
				if v != nil {
					return ce.SRV_ERR
				}
			}
		}
	}
	return ce.SUCC
}
