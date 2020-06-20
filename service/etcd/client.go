package etcd

import (
	"context"
	_ "fmt"
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
	//fmt.Println("add service", service)
	c.mcs[service][nclnt] = struct{}{}
}

func (c *conn) Call(service, api string, req, rsp proto.Message) ce.Errno {
	//fmt.Println("call service", service)
	if _, ok := c.mcs[service]; !ok {
		//fmt.Println("service", service, "not found")
		return ce.SRV_ERR
	}
	for clnt, _ := range c.mcs[service] {
		rets := reflect.ValueOf(clnt.This).MethodByName(api).
			Call([]reflect.Value{
				reflect.ValueOf(context.TODO()),
				reflect.ValueOf(cp.NewRequest(req)),
			})
		for _, ret := range rets {
			switch v := ret.Interface().(type) {
			case *cp.Response:
				if en := ce.Errno(v.GetErrno()); en != ce.SUCC {
					return en
				}
				proto.Unmarshal(v.GetBody(), rsp)
				return ce.SUCC
			case error:
				if v != nil {
					goto CONTINUE
				}
			default:
				// TODO: WARNING
				return ce.BAD_REQ
			}
		}
	CONTINUE:
	}
	return ce.SUCC
}

func (c *conn) InCall(service, api string, req *cp.Request, rsp proto.Message) ce.Errno {
	//fmt.Println("call service", service)
	if _, ok := c.mcs[service]; !ok {
		//fmt.Println("service", service, "not found")
		return ce.SRV_ERR
	}
	for clnt, _ := range c.mcs[service] {
		rets := reflect.ValueOf(clnt.This).MethodByName(api).
			Call([]reflect.Value{
				reflect.ValueOf(context.TODO()),
				reflect.ValueOf(req),
			})
		for _, ret := range rets {
			switch v := ret.Interface().(type) {
			case *cp.Response:
				if en := ce.Errno(v.GetErrno()); en != ce.SUCC {
					return en
				}
				proto.Unmarshal(v.GetBody(), rsp)
				return ce.SUCC
			case error:
				if v != nil {
					goto CONTINUE
				}
			default:
				// TODO: WARNING
				return ce.BAD_REQ
			}
		}
	CONTINUE:
	}
	return ce.SUCC
}
