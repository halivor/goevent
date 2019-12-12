package net

import (
	"net"
	"os"
)

func InIpv4() (string, error) {
	var ip string
	switch addrs, err := net.InterfaceAddrs(); {
	case err != nil:
		return "", err
	default:
		for _, address := range addrs {
			// 检查ip地址判断是否回环地址
			if ipnet, ok := address.(*net.IPNet); ok {
				if ipnet.IP.To4() != nil && !ipnet.IP.IsLoopback() {
					ip = ipnet.IP.To4().String()
				}
			}
		}
		if len(ip) == 0 {
			return "", os.ErrNotExist
		}
	}
	return ip, nil
}
