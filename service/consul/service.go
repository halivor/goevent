package consul

import (
	svc "github.com/halivor/goutil/service"
	"github.com/hashicorp/consul/api"
)

func (c *Consul) AddService() {
	c.cc.Agent().ServiceRegister(&api.AgentServiceRegistration{
		ID:                c.ID,
		Name:              c.Name,
		Tags:              c.Tags,
		Address:           c.IP,
		Port:              c.Port,
		Meta:              c.Meta,
		EnableTagOverride: false,
	})
}

func (c *Consul) NewClnt(svc.Server, interface{}) {
}
