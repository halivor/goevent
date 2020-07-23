package service

import (
	"context"

	cp "co.mplat.com/packet"
	ce "co.mplat.com/util/errno"
	"github.com/golang/protobuf/proto"
)

type Entry struct {
	Type string      `json:"type"`
	Host string      `json:"host"`
	Port int         `json:"port"`
	Grpc Method      `json:"-"`
	Http interface{} `json:"-"`
}

func (e *Entry) Call(ctx context.Context, name string, params []interface{}) (en ce.Errno) {
	switch e.Type {
	case "grpc":
		if len(params) != 3 {
			return ce.BAD_REQ
		}
		req := params[0].(proto.Message)
		rsp := params[1].(proto.Message)
		rst, er := e.Grpc(ctx, cp.NewRequest(0, req))
		if er != nil {
			return ce.SRV_ERR
		}
		if ce.Errno(rst.GetErrno()) != ce.SUCC {
			return ce.Errno(rst.GetErrno())
		}
		proto.Unmarshal(rst.GetBody(), rsp)
	case "http":
	}
	return ce.SUCC
}
