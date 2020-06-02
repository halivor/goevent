package consul

import (
	"sync"
	"time"

	"github.com/golang/protobuf/proto"
	ce "github.com/halivor/common/golang/util/errno"
	svc "github.com/halivor/goutil/service"
	"github.com/hashicorp/consul/api"
)

type Consul struct {
	Proj     string
	ID       string
	Name     string
	Tags     []string
	Meta     map[string]string
	WaitTime time.Duration

	IP   string
	Port int

	cc *api.Client
	kv
	stat
	Index sync.Map
}

type Config struct {
	Proj string
	ID   string
	Name string
	IP   string
	Port int
	Tags []string
	Meta map[string]string
}

func init() {
	svc.Register("consul", New)
}

func New() svc.Service {
	return &Consul{
		WaitTime: time.Minute * 15,
	}
}

func (c *Consul) Init(param interface{}) {
	params := param.(map[string]interface{})
	addr := "127.0.0.1:8500"
	if address, ok := params["addr"].(string); ok {
		addr = address
	}
	scheme := "http"
	if proto, ok := params["proto"].(string); ok {
		scheme = proto
	}
	var e error
	if c.cc, e = api.NewClient(&api.Config{
		Address: addr,
		Scheme:  scheme,
	}); e != nil {
		panic(e)
	}

	if tmo, ok := params["waittime"].(time.Duration); ok && tmo > 0 {
		c.WaitTime = tmo
	}
}

func (c *Consul) AddConf(cfg *Config) {
	c.Proj = cfg.Proj
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

func (c *Consul) getIdx(key string) uint64 {
	if idx, ok := c.Index.Load(key); ok {
		return idx.(uint64)
	}

	return 1
}

func (c *Consul) SetUp(name string, m svc.Method) {
}
func (c *Consul) Call(name string, req proto.Message, rsp proto.Message) ce.Errno {
	return ce.SUCC
}
func (c *Consul) SignUp(svc.Server) {
}
