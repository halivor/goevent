package service

import "encoding/json"

type Client struct {
	Service   string                     `json:"service"`
	Interface map[string]json.RawMessage `json:"interface"`
	Type      string                     `json:"type"`
	Host      string                     `jsno:"host"`
	This      interface{}                `json:"-"`
}
