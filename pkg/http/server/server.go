package server

import (
	"fmt"
	"ganesh.provengo.io/api"
	localStructs "ganesh.provengo.io/internal/structs"
	"ganesh.provengo.io/pkg/ipaddr"
	"github.com/gin-gonic/gin"
	"log"
	"sync"
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

func customerConsumer(channel chan localStructs.DataLogin, wg *sync.WaitGroup) {
	defer wg.Done()
	for msg := range channel {
		fmt.Println(msg)
	}
}

func Run(port int, isSSL bool) {
	if validatePort(port) != nil {
		log.Fatalf("invalid port %d", port)
		return
	}

	wg := &sync.WaitGroup{}
	customerChannel := make(chan localStructs.DataLogin)

	wg.Add(1)
	go customerConsumer(customerChannel, wg)

	defer func() {
		wg.Done()
		close(customerChannel)
	}()

	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()
	r.GET("/ping", api.Ping())
	r.GET("/keep-alive/:id", api.KeepAlive())
	r.GET("/", api.Default())
	r.POST("/send-user", api.SendUser(customerChannel))
	_port := fmt.Sprintf(":%d", port)

	aliveAndKicking(port, isSSL)

	if isSSL {
		_ = r.RunTLS(_port, "./assets/certs/server.crt", "./assets/certs/server.key")
	} else {
		_ = r.Run(_port)
	}
}
