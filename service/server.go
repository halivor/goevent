package service

import (
	"encoding/json"
)

type http struct {
	Name   string `json:"name"`
	Method string `json:"method"`
	Path   string `json:"path"`
}

type server struct {
	// TODO: 增加接口调用超时
	Key       string                     `json:"key"`
	Service   string                     `json:"service"`
	Host      string                     `json:"host"`
	Type      string                     `json:"type"`
	Interface map[string]json.RawMessage `json:"interface"`
}

type Server interface {
	AddGrpc(name string)
	AddHttp(name, method, path string)
	Data() string
	GetKey() string
	GetService() string
	GetHost() string
	GetType() string
	GetInterface() map[string]json.RawMessage
}

func NewServer(key, service, typ, host string) Server {
	return &server{
		Key:       key,
		Service:   service,
		Host:      host,
		Type:      typ,
		Interface: make(map[string]json.RawMessage, 32),
	}
}

func ParseServer(data []byte) Server {
	s := &server{}
	json.Unmarshal(data, s)
	return s
}

func (s *server) AddHttp(name, method, path string) {
	s.Interface[name], _ = json.Marshal(&http{name, method, path})
}

func (s *server) AddGrpc(name string) {
	s.Interface[name] = []byte("{}")
}

func (s *server) Data() string {
	pb, _ := json.MarshalIndent(s, "", "    ")
	return string(pb)
}

func (s *server) GetKey() string {
	return s.Key
}

func (s *server) GetService() string {
	return s.Service
}

func (s *server) GetHost() string {
	return s.Host
}

func (s *server) GetType() string {
	return s.Type
}

func (s *server) GetInterface() map[string]json.RawMessage {
	return s.Interface
}
