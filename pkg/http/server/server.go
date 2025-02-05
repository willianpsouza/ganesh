package server

import (
	"fmt"
	"ganesh.provengo.io/api"
	"ganesh.provengo.io/pkg/ipaddr"
	"github.com/gin-gonic/gin"
)

func validatePort(port int) error {
	if port >= 1024 && port <= 64000 {
		return nil
	}
	return fmt.Errorf("invalid port %d", port)
}

func aliveAndKicking(port int, isSSL bool) {
	_port := fmt.Sprintf(":%d", port)
	_ipaddr := ipaddr.DetectLocalIP()
	_httpProto := "http://"

	if isSSL {
		_httpProto = "https://"
	}
	for _i := range _ipaddr {
		fmt.Printf("Listening on: %s%s%s\n", _httpProto, _ipaddr[_i], _port)
	}
}

func Run(port int, isSSL bool) {
	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()
	r.GET("/ping", api.Ping())
	r.GET("/keep-alive/:id", api.KeepAlive())
	r.GET("/", api.Default())
	_port := fmt.Sprintf(":%d", port)

	aliveAndKicking(port, isSSL)

	if isSSL {
		_ = r.RunTLS(_port, "./assets/certs/server.crt", "./assets/certs/server.key")
	} else {
		_ = r.Run(_port)
	}
}
