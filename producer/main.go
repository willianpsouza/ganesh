package main

import (
	"encoding/json"
	"fmt"
	"time"

	localStructs "ganesh.provengo.io/local_structs"
	localSetup "ganesh.provengo.io/tools/setup"
	fkrV4 "github.com/go-faker/faker/v4"
	"github.com/google/uuid"
	"github.com/nats-io/nats.go"
	"log"

	"sync"
)

const (
	remoteQueue = localSetup.QueueName
	totalTasks  = localSetup.TotalTasks
)

func GenerateUsers(total int) []localStructs.DataLogin {
	var users []localStructs.DataLogin
	for i := 0; i < total; i++ {
		_uudi := uuid.New()
		users = append(users, localStructs.DataLogin{
			UUID:      _uudi.String(),
			Username:  fkrV4.Username(),
			Password:  fkrV4.Password(),
			Timestamp: time.Now().UnixMilli(),
			Sequence:  int64(i),
		})
	}
	return users
}

func GoUsers(reqChan chan localStructs.DataLogin) {

	for _, value := range GenerateUsers(localSetup.UsersGenerate) {
		reqChan <- value
	}

}

func NatsWorker(idGg int, reqChannel chan localStructs.DataLogin, wg *sync.WaitGroup) {
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

func startWorkTasks(wg *sync.WaitGroup, reqChannel chan localStructs.DataLogin) {
	log.Printf("Starting work tasks")
	for i := 0; i < totalTasks; i++ {
		wg.Add(1)
		go NatsWorker(i, reqChannel, wg)
	}
}

func runningServer() {
	wg := sync.WaitGroup{}
	loginQueue := make(chan localStructs.DataLogin)
	startWorkTasks(&wg, loginQueue)
	total := 4
	for {
		if total == 0 {
			break
		}
		GoUsers(loginQueue)
		total -= 1
	}
}

func main() {
	runningServer()
}
