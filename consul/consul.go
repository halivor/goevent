package consul

import (
	"time"

	"github.com/hashicorp/consul/api"
)

type Consul struct {
	ID       string
	Name     string
	Tags     []string
	Meta     map[string]string
	WaitTime time.Duration

	IP   string
	Port int

	*api.Client
	kv
}

type Config struct {
	ID   string
	Name string
	IP   string
	Port int
	Tags []string
	Meta map[string]string
}

func New(address, proto string) *Consul {
	addr := "127.0.0.1:8500"
	if len(address) != 0 {
		addr = address
	}
	scheme := "http"
	if len(proto) == 0 {
		scheme = proto
	}
	clnt, e := api.NewClient(&api.Config{
		Address: addr,
		Scheme:  scheme,
	})
	if e != nil {
		panic(e)
	}
	return &Consul{
		WaitTime: time.Minute * 15,
		Client:   clnt,
	}
}

func (c *Consul) AddConf(cfg *Config) {
	c.ID = cfg.ID
	c.Name = cfg.Name
	c.Tags = cfg.Tags
	c.IP = cfg.IP
	c.Port = cfg.Port
	c.Meta = cfg.Meta
}

func (c *Consul) AddTag(tag string) {
	c.Tags = append(c.Tags, tag)
}

func (c *Consul) SetID(id string) {
	c.ID = id
}

func (c *Consul) SetName(name string) {
	c.Name = name
}

func (c *Consul) SetWaitTime(waitTime time.Duration) {
	if waitTime > time.Minute && waitTime < time.Hour {
		c.WaitTime = waitTime
	}
}

func (c *Consul) AddMetaData(key, val string) {
	c.Meta[key] = val
}

func (c *Consul) SetIP(ip string) {
	c.IP = ip
}

func (c *Consul) SetPort(port int) {
	c.Port = port
}
