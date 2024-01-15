package main

import (
	"fmt"
	"net"
)

// 获取本机内网IP
func getLocalIP() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		fmt.Println("获取本地IP错误, err = ", err.Error())
		return ""
	}

	for _, addr := range addrs {
		ip, ok := addr.(*net.IPNet)
		if ok {
			ipAddr := ip.IP
			if ipAddr.IsPrivate() && ipAddr.To4() != nil && !ipAddr.IsLoopback() {
				return ip.IP.String()
			}
		}
	}

	return ""
}
