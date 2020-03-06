package net

import (
	"net"
	"os"
)

func InIpv4() ([]string, error) {
	var ips []string
	switch addrs, err := net.InterfaceAddrs(); {
	case err != nil:
		return nil, err
	default:
		for _, address := range addrs {
			// 检查ip地址判断是否回环地址
			if ipnet, ok := address.(*net.IPNet); ok {
				if ipnet.IP.To4() != nil && !ipnet.IP.IsLoopback() {
					ip := ipnet.IP.To4().String()
					ips = append(ips, ip)
				}
			}
		}
		if len(ips) == 0 {
			return nil, os.ErrNotExist
		}
	}
	return ips, nil
}
