package main

import (
	GaneshHttpClient "ganesh.provengo.io/pkg/http/client"
	"time"
)

func main() {
	url := "http://192.168.86.45:8080/keep-alive/lalalalla"

	GaneshHttpClient.StartClient(url, 100, 3600*time.Second)

}
