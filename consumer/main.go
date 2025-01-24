package main

import (
	"encoding/json"
	"fmt"
	locasStructs "ganesh.provengo.io/local_structs"
	localEncrypt "ganesh.provengo.io/tools/encrypt"
	"github.com/nats-io/nats.go"
	"log"
	"os"
	"runtime"
	"sync"
	"time"
)

const (
	remoteQueue  = "ganesh_provengo_io_login"
	totalTasks   = 2
	queueGroup   = "login_workers"
	consumerType = "workers" // pubSub
)

func UserPassword(c *locasStructs.DataLogin) string {
	algo := os.Getenv("ALGORITHM")
	str := fmt.Sprintf("%s:%s", c.Username, c.Password)
	encrypt := localEncrypt.CalculateChecksum([]byte(str), algo)
	return encrypt
}

func PasswordWorkers(idGg int, reqChannel chan locasStructs.DataLogin, wg *sync.WaitGroup) {
	defer wg.Done()
	fmt.Printf("Starting worker %d\n", idGg)
	for req := range reqChannel {
		password := UserPassword(&req)
		totalTime := time.Now().UnixMilli() - req.Timestamp
		fmt.Printf("Seq: %d Username %s Password %s, Total Time %d\n", req.Sequence, req.Username, password, totalTime)
	}
}

func NatsWorker(idGg int, reqChannel chan locasStructs.DataLogin, wg *sync.WaitGroup) {
	log.Printf("Starting NATS worker: %d\n", idGg)
	nc, err := nats.Connect(nats.DefaultURL)
	if err != nil {
		log.Fatal("error connecting to NATS")
	}
	defer func() {
		nc.Close()
		wg.Done()
	}()

	switch consumerType {
	case "pub_sub":
		//PUB/SUB PATTERN GROUP QUEUE
		_, err = nc.Subscribe(remoteQueue, func(msg *nats.Msg) {
			var req locasStructs.DataLogin
			_err := json.Unmarshal(msg.Data, &req)
			if _err != nil {
				log.Printf("error unmarshalling data from NATS: %v", err)
			}
			reqChannel <- req
		})
	case "workers":
		//WORKER PROCESS PATTERN
		subject := fmt.Sprintf("%s.*", remoteQueue)
		_, err = nc.QueueSubscribe(subject, queueGroup, func(msg *nats.Msg) {
			var req locasStructs.DataLogin
			_err := json.Unmarshal(msg.Data, &req)
			if _err != nil {
				log.Printf("error unmarshalling data from NATS: %v", err)
			}
			reqChannel <- req
		})

		if err != nil {
			log.Fatalf("error subscribing to NATS: %v", err)
		}
	}

	select {}
}

func startWorkTasks(wg *sync.WaitGroup, reqChannel chan locasStructs.DataLogin) {
	log.Printf("Starting work tasks")
	for i := 0; i < totalTasks; i++ {
		wg.Add(1)
		go NatsWorker(i, reqChannel, wg)
	}

	for i := 0; i < totalTasks; i++ {
		wg.Add(1)
		go PasswordWorkers(i+1, reqChannel, wg)
	}
}

func runningServer() {
	wg := sync.WaitGroup{}
	loginQueue := make(chan locasStructs.DataLogin)

	startWorkTasks(&wg, loginQueue)

}

func main() {
	runtime.GOMAXPROCS(2)
	runningServer()
	select {}
}
