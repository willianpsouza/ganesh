package main

import (
	"encoding/json"
	"fmt"
	"time"

	locasStructs "ganesh.provengo.io/local_structs"
	fakerV4 "github.com/go-faker/faker/v4"
	"github.com/google/uuid"
	"github.com/nats-io/nats.go"
	"log"

	"sync"
)

const (
	remoteQueue = "ganesh_provengo_io_login"
	totalTasks  = 2
)

func GenerateUsers(total int) []locasStructs.DataLogin {
	var users []locasStructs.DataLogin
	for i := 0; i < total; i++ {
		_uudi := uuid.New()
		users = append(users, locasStructs.DataLogin{
			UUID:      _uudi.String(),
			Username:  fakerV4.Username(),
			Password:  fakerV4.Password(),
			Timestamp: time.Now().UnixMilli(),
			Sequence:  int64(i),
		})
	}
	return users
}

func GoUsers(reqChan chan locasStructs.DataLogin) {

	for _, value := range GenerateUsers(100000) {
		reqChan <- value
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

	for req := range reqChannel {
		msg, err := json.Marshal(&req)
		if err != nil {
			log.Fatal("Error marshalling data:", err)
		}

		subject := fmt.Sprintf("%s.%s", remoteQueue, req.UUID)
		err = nc.Publish(subject, msg)
		if err != nil {
			log.Fatal("Error publishing message to NATS", err)

		}
	}
}

func startWorkTasks(wg *sync.WaitGroup, reqChannel chan locasStructs.DataLogin) {
	log.Printf("Starting work tasks")
	for i := 0; i < totalTasks; i++ {
		wg.Add(1)
		go NatsWorker(i, reqChannel, wg)
	}
}

func runningServer() {
	wg := sync.WaitGroup{}
	loginQueue := make(chan locasStructs.DataLogin)

	startWorkTasks(&wg, loginQueue)
	GoUsers(loginQueue)
}

func main() {
	runningServer()
}
