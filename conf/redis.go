package conf

import "strconv"

type Redis struct {
	Addr string `json:"addr"`
	Auth string `json:auth`
	DB   int    `json:"db"`
}

func (rc *Redis) String() string {

	return rc.Addr + ":" + rc.Auth + "/" + strconv.Itoa(rc.DB)
}
