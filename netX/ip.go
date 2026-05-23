package netX

import (
	"fmt"
	"net"
)

// GetOutboundIP 获得对外发送消息的 IP 地址
// 如果获取失败，返回空字符串
func GetOutboundIP() string {
	ip, _ := GetOutboundIPWithError()
	return ip
}

// GetOutboundIPWithError 获得对外发送消息的 IP 地址，并返回错误信息
func GetOutboundIPWithError() (string, error) {
	conn, err := net.Dial("udp", "8.8.8.8:80") // 8.8.8.8 is google's DNS
	if err != nil {
		return "", err
	}
	defer conn.Close()

	localAddr, ok := conn.LocalAddr().(*net.UDPAddr)
	if !ok {
		return "", fmt.Errorf("unexpected LocalAddr type: %T", conn.LocalAddr())
	}
	return localAddr.IP.String(), nil
}
