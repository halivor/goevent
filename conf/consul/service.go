package consul

import (
	"github.com/hashicorp/consul/api"
)

func (c *Consul) AddService() {
	c.Agent().ServiceRegister(&api.AgentServiceRegistration{
		ID:                c.ID,
		Name:              c.Name,
		Tags:              c.Tags,
		Address:           c.IP,
		Port:              c.Port,
		Meta:              c.Meta,
		EnableTagOverride: false,
	})
}
