package conf

import (
	"encoding/json"
)

type Service interface {
	Add(string, *SvcCnf)
	Commit()
}

func NewSvc(name string, host string) Service {
	return &service{
		Service: name,
		Host:    host,
		Api:     make(map[string][]*SvcCnf, 32),
	}
}

type service struct {
	Service string               `json:"service"`
	Host    string               `json:"host"`
	Api     map[string][]*SvcCnf `json:"api"`
}

type SvcCnf struct {
	Type    string `json:"type"`  // grpc, http
	Level   string `json:"level"` // priority: low 1 - 5 high
	Call    string `json:"call"`  // call: http{GET/POST/DEL/PUT}
	Service string `json:"path"`  // http{/usr/login} grpc{name}
	Pack    string `json:"pack"`  // http{json, proto} grpc{proto}
	Load    string `json:"load"`  // load: default 0
}

func (svc *service) Add(name string, cnf *SvcCnf) {
	if _, ok := svc.Api[name]; !ok {
		svc.Api[name] = make([]*SvcCnf, 0, 2)
	}
	svc.Api[name] = append(svc.Api[name], cnf)
}

func (svc *service) Commit() {
	pb, _ := json.Marshal(svc)
	sc.Put(svc.Host, string(pb))
}
