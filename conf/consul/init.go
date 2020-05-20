package consul

import (
	"github.com/halivor/goutil/conf"
)

func init() {
	conf.Register("consul", New())
}
