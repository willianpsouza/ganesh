package main

import (
	GaneshHTTPServer "ganesh.provengo.io/pkg/http/server"
	"runtime"
)

const (
	ServicePort = 8080
	IsSSL       = false
)

func main() {
	runtime.GOMAXPROCS(2)
	GaneshHTTPServer.Run(ServicePort, IsSSL)
}
