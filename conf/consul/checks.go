package consul

import (
	"strconv"

	"github.com/hashicorp/consul/api"
)

func (c *Consul) AddTcpCheck(note string) {
	c.Agent().CheckRegister(&api.AgentCheckRegistration{
		ID:        c.ID,
		Name:      c.Name,
		Notes:     note,
		ServiceID: c.ID,
		AgentServiceCheck: api.AgentServiceCheck{
			CheckID:  c.ID,
			Name:     c.Name,
			Interval: "1m",
			Timeout:  "1m",
			TCP:      c.IP + ":" + strconv.Itoa(c.Port),

			DeregisterCriticalServiceAfter: "5m",
		},
	})
}

func (c *Consul) AddHttpCheck(note string, path string,
	method string, header map[string][]string, body string) {
	c.Agent().CheckRegister(&api.AgentCheckRegistration{
		ID:        c.ID,
		Name:      c.Name,
		Notes:     note,
		ServiceID: c.ID,
		AgentServiceCheck: api.AgentServiceCheck{
			CheckID:  c.ID,
			Name:     c.Name,
			Interval: "1m",
			Timeout:  "1m",
			HTTP:     "http://" + c.IP + ":" + strconv.Itoa(c.Port) + path,
			Method:   method,
			Header:   header,
			Body:     body,

			DeregisterCriticalServiceAfter: "5m",
		},
	})
}
