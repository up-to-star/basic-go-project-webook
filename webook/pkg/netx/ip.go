package netx

import "net"

func GetOutBoundIP() string {
	conn, err := net.Dial("udp", "114.114.114.114:80")
	if err != nil {
		return ""
	}
	defer conn.Close()
	return conn.LocalAddr().(*net.UDPAddr).IP.String()
}
