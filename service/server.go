package service

import "encoding/json"

type http struct {
	Name   string `json:"name"`
	Method string `json:"method"`
	Path   string `json:"path"`
}

type server struct {
	key       string
	Module    string                 `json:"module"`
	Type      string                 `json:"type"`
	Host      string                 `json:"host"`
	Interface map[string]interface{} `json:"interface"`
}

type Server interface {
	AddGrpc(name string)
	AddHttp(name, method, path string)
	Key() string
	Data() string
}

func NewServer(key, module, typ, host string) Server {
	return &server{
		key:       key,
		Type:      typ,
		Host:      host,
		Module:    module,
		Interface: make(map[string]interface{}, 32),
	}
}

func (s *server) AddHttp(name, method, path string) {
	s.Interface[name] = &http{name, method, path}
}

func (s *server) AddGrpc(name string) {
	s.Interface[name] = struct{}{}
}

func (s *server) Key() string {
	return s.key
}

func (s *server) Data() string {
	pb, _ := json.Marshal(s)
	return string(pb)
}
