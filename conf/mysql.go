package conf

type Mysql struct {
	Addr string `json:"addr"` // format => ip:port
	User string `json:"user"`
	Pwd  string `json:"password"`
	DB   string `json:"db"`
}

func (mc *Mysql) String() string {
	if len(mc.Addr) == 0 {
		mc.Addr = "127.0.0.1:3306"
	}
	if len(mc.User) == 0 {
		panic("user not exist")
	}
	// user:password@tcp(ip:port)/database?charset=utf8
	return mc.User + ":" + mc.Pwd +
		"@tcp(" + mc.Addr + ")/" +
		mc.DB
}
