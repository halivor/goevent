package service

type Client struct {
	This   interface{} `json:"-"`
	Module string      `json:"module"`
	Type   string      `json:"type"`
	Addr   string      `jsno:"addr"`
}
