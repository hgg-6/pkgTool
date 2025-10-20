package netX

import "net"

// GetOutboundIP 获得对外发送消息的 IP 地址
func GetOutboundIP() string {
	conn, err := net.Dial("udp", "8.8.8.8:80") // 8.8.8.8 is google's DNS
	if err != nil {
		return ""
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr) // *net.UDPAddr is a struct
	return localAddr.IP.String()
}
