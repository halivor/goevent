package etcd

import (
	"context"
	"fmt"
	"reflect"

	cp "co.mplat.com/packet"
	ce "co.mplat.com/util/errno"
	"github.com/golang/protobuf/proto"
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

func (c *conn) Call(service, api string, uid int64, req, rsp proto.Message) ce.Errno {
	//fmt.Println("call service", service)
	if _, ok := c.mcs[service]; !ok {
		fmt.Println("service", service, "not found")
		return ce.SRV_ERR
	}
	for clnt, _ := range c.mcs[service] {
		call := reflect.ValueOf(clnt.This).MethodByName(api)
		if !call.IsValid() {
			return ce.NO_SRV_ERR
		}
		rets := call.Call([]reflect.Value{
			reflect.ValueOf(context.TODO()),
			reflect.ValueOf(cp.NewRequest(uid, req)),
		})

		for _, ret := range rets {
			// TODO: 确认各种error信息
			switch v := ret.Interface().(type) {
			case *cp.Response:
				if en := ce.Errno(v.GetErrno()); en != ce.SUCC {
					return en
				}
				proto.Unmarshal(v.GetBody(), rsp)
				return ce.SUCC
			case error:
				if v != nil {
					fmt.Println("call failed", v)
					return ce.SRV_ERR
				}
			default:
				return ce.BAD_REQ
			}
		}
		return ce.SUCC
	}
	return ce.SUCC
}

func (c *conn) InCall(service, api string, req *cp.Request, rsp proto.Message) ce.Errno {
	//fmt.Println("in call service", service)
	if _, ok := c.mcs[service]; !ok {
		fmt.Println("service", service, "not found")
		return ce.SRV_ERR
	}
	for clnt, _ := range c.mcs[service] {
		call := reflect.ValueOf(clnt.This).MethodByName(api)
		if !call.IsValid() {
			return ce.NO_SRV_ERR
		}
		rets := call.Call([]reflect.Value{
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
					fmt.Println("in call failed", v)
					return ce.SRV_ERR
				}
			default:
				return ce.BAD_REQ
			}
		}
	}
	return ce.SUCC
}
