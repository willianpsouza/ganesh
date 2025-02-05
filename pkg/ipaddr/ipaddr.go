package ipaddr

import "net"

func DetectLocalIP() []string {
	var data []string

	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return data
	}
	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				data = append(data, ipnet.IP.String())
			}
		}
	}
	return data
}
