package gvalue

import (
	"net"
	"os"
	"strings"
)

// IdcName 获取idc名称
func IdcName() string {
	// 优先尝试从容器环境变量获取idc
	var envKeys = []string{"SYS_IDC_NAME", "QIHOO_IDC"} // 新旧容日环境变量不一致
	for _, key := range envKeys {
		idc := os.Getenv(key)
		if len(idc) > 0 {
			return idc
		}
	}

	// 再尝试重虚拟机hostname拆解
	hostname, err := os.Hostname()
	if err != nil {
		return ""
	}

	// 公司的虚拟机hostname统一第三位是idc信息，例如：app33.add.bjmd.qihoo.net
	hostArr := strings.Split(hostname, ".")
	if len(hostArr) > 3 {
		return hostArr[2]
	}

	return ""
}

// Hostname 获取hostname
func Hostname() string {
	hostname, _ := os.Hostname()
	
	return hostname
}

// LocalIP 获取机器本地ip地址
func LocalIP() string {
	ipList := []string{"101.226.4.6:80", "218.30.118.6:80", "123.125.81.6:80", "114.114.114.114:80", "8.8.8.8:80"}
	for _, ip := range ipList {
		conn, err := net.Dial("udp", ip)
		if err != nil {
			continue
		}
		localAddr := conn.LocalAddr().(*net.UDPAddr)
		conn.Close()
		return localAddr.IP.String()
	}

	return ""
}
