package ip

import "net"

func GetIpv4() string {
	address, err := net.InterfaceAddrs()
	if err != nil {
		return ""
	}
	for _, address := range address {
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String()
			}
		}
	}
	return ""
}
