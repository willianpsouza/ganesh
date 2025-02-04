package main

import (
	"encoding/json"
	"fmt"
	"runtime"
	"sync/atomic"
	"time"

	localSetup "ganesh.provengo.io/internal/setup"
	localStructs "ganesh.provengo.io/internal/structs"
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
		UniqueUser := uuid.New().String()
		users = append(users, localStructs.DataLogin{
			UUID:      UniqueUser,
			Username:  fkrV4.Username(),
			Password:  fkrV4.Password(),
			Timestamp: time.Now().UnixMilli(),
			Sequence:  int64(i),
		})
	}
	return users
}

func GoUsers(reqChan chan localStructs.DataLogin, totalUsers *uint32) {
	totalMessages := 0
	for _, value := range GenerateUsers(localSetup.UsersGenerate) {
		reqChan <- value
		_ = value
		totalMessages += 1
		atomic.AddUint32(totalUsers, 1)
	}

	log.Printf("%d users generated", totalMessages)
}

func NatsWorker(idGg int, reqChannel chan localStructs.DataLogin, wg *sync.WaitGroup, totalMessages *uint32) {
	log.Printf("Starting NATS worker: %d\n", idGg)
	nc, err := nats.Connect(nats.DefaultURL)

	if err != nil {
		log.Fatal("error connecting to NATS")
	}

	defer func() {
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
		atomic.AddUint32(totalMessages, 1)
	}
}

func startWorkTasks(wg *sync.WaitGroup, reqChannel chan localStructs.DataLogin, totalMesssages *uint32) {
	for i := 0; i < totalTasks; i++ {
		wg.Add(1)
		go NatsWorker(i, reqChannel, wg, totalMesssages)
	}
	defer func() {
		wg.Done()
	}()
}

func runningServer() {
	wg := sync.WaitGroup{}
	loginQueue := make(chan localStructs.DataLogin)
	var totalMessages uint32
	var totalUsers uint32

	wg.Add(1)
	go startWorkTasks(&wg, loginQueue, &totalMessages)

	for i := 0; i < 4; i++ {
		GoUsers(loginQueue, &totalUsers)
	}

	//WAITING FOR ALL MESSAGES HAS BEEN SENT BEFORE CLOSE CHANNEL
	for {
		if totalMessages == totalUsers {
			close(loginQueue)
			break
		}
		time.Sleep(100 * time.Millisecond)
	}

	defer func() {
		fmt.Printf("Total Messages: %d\n", totalMessages)
		fmt.Printf("Total Users: %d\n", totalUsers)
	}()
}

func main() {
	runtime.GOMAXPROCS(2)
	runningServer()
}
