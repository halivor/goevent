package etcd

import (
	svc "github.com/halivor/goutil/service"
)

func (c *conn) SignUp(srv svc.Server) {
	c.Put(srv.Key(), srv.Data())
}
