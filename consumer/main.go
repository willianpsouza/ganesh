package main

import (
	"context"
	"encoding/json"
	"fmt"
	localPostgres "ganesh.provengo.io/lib/postgres"
	localRedis "ganesh.provengo.io/lib/redis"
	localStructs "ganesh.provengo.io/local_structs"
	localEncrypt "ganesh.provengo.io/tools/encrypt"
	localSetup "ganesh.provengo.io/tools/setup"
	"github.com/nats-io/nats.go"
	"log"
	"os"
	"runtime"
	"sync"
	"time"
)

const (
	remoteQueue  = localSetup.QueueName
	totalTasks   = localSetup.TotalTasks
	queueGroup   = localSetup.QueueGroup
	consumerType = localSetup.ConsumerType // pubSub
)

type Channels struct {
	passwordQueue chan localStructs.DataLogin
	redisQueue    chan localStructs.DataLogin
	postgresQueue chan localStructs.DataLogin
}

func UserPassword(c *localStructs.DataLogin) string {
	algo := os.Getenv("ALGORITHM")
	str := fmt.Sprintf("%s:%s", c.Username, c.Password)
	encrypt := localEncrypt.CalculateChecksum([]byte(str), algo)
	return encrypt
}

func PasswordWorkers(idGg int, reqChannel chan localStructs.DataLogin, wg *sync.WaitGroup) {
	defer wg.Done()
	fmt.Printf("Starting Password Worker %d\n", idGg)
	for req := range reqChannel {
		password := UserPassword(&req)
		totalTime := time.Now().UnixMilli() - req.Timestamp
		fmt.Printf("Seq: %d Username %s Password %s, Total Time %d\n", req.Sequence, req.Username, password, totalTime)
	}
}

func NatsWorker(idGg int, reqChannel Channels, wg *sync.WaitGroup) {
	log.Printf("Starting NATS worker: %d\n", idGg)
	nc, err := nats.Connect(localSetup.NatsAddress)
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
			var req localStructs.DataLogin
			_err := json.Unmarshal(msg.Data, &req)
			if _err != nil {
				log.Printf("error unmarshalling data from NATS: %v", err)
			}
			reqChannel.redisQueue <- req
			reqChannel.postgresQueue <- req
			reqChannel.passwordQueue <- req

		})
	case "workers":
		//WORKER PROCESS PATTERN
		subject := fmt.Sprintf("%s.*", remoteQueue)
		_, err = nc.QueueSubscribe(subject, queueGroup, func(msg *nats.Msg) {
			var req localStructs.DataLogin
			_err := json.Unmarshal(msg.Data, &req)
			if _err != nil {
				log.Printf("error unmarshalling data from NATS: %v", err)
			}
			reqChannel.redisQueue <- req
			reqChannel.postgresQueue <- req
			reqChannel.passwordQueue <- req
		})

		if err != nil {
			log.Fatalf("error subscribing to NATS: %v", err)
		}
	}

	select {}
}

func RedisWorker(idGg int, reqChannel chan localStructs.DataLogin, wg *sync.WaitGroup, ctx context.Context) {

	fmt.Printf("Starting Redis worker: %d\n", idGg)

	conn, _err := localRedis.RedisConnection(ctx)
	if _err != nil {
		log.Fatalf("Error connecting to Redis %v", _err)
	}

	defer func() {
		wg.Done()
		conn.Close()
	}()

	_randoTime := localSetup.RedisDefaultTTL

	for req := range reqChannel {
		_data := localRedis.RedisData{
			Key:   req.UUID,
			Value: fmt.Sprintf("Username %s Password %s", req.Username, req.Password),
			TTL:   _randoTime,
		}

		go func() {
			_err := conn.Set(ctx, _data)
			if _err != nil {
				log.Fatalf("Error setting data in Redis %v", _err)
			}
		}()
	}
}

func PostgresWorker(idGg int, reqChannel chan localStructs.DataLogin, wg *sync.WaitGroup, ctx context.Context) {
	fmt.Printf("Starting Postgres Worker: %d\n", idGg)

	conn, err := localPostgres.PostgresConnection(ctx, localSetup.PostgresURI)

	if err != nil {
		log.Fatalf("Error connecting to Postgres %v", err)
	}

	err = conn.Ping(ctx)
	if err != nil {
		log.Fatalf("Error pinging Postgres %v", err)
	}

	defer func() {
		wg.Done()
		conn.Close()
	}()

	for req := range reqChannel {
		wg.Add(1)
		go func() {
			err := conn.InsertUserPassword(ctx, req)
			if err != nil {
				log.Fatalf("Error inserting user password %v", err)
			}
		}()
	}
}

func startWorkTasks(wg *sync.WaitGroup, reqChannel Channels, ctx context.Context) {

	log.Printf("Starting work tasks")
	for i := 0; i < totalTasks; i++ {

		//Nats workers
		wg.Add(1)
		go NatsWorker(i, reqChannel, wg)

		//Passwords workers
		wg.Add(1)
		go PasswordWorkers(i, reqChannel.passwordQueue, wg)
	}

	for i := 0; i < totalTasks+2; i++ {
		//Redis Workers
		wg.Add(1)
		go RedisWorker(i, reqChannel.redisQueue, wg, ctx)

		//Postgres Workers
		wg.Add(1)
		go PostgresWorker(i, reqChannel.postgresQueue, wg, ctx)

	}
}

func runningServer(ctx context.Context, wg *sync.WaitGroup) {

	var channels Channels
	channels.passwordQueue = make(chan localStructs.DataLogin, totalTasks)
	channels.redisQueue = make(chan localStructs.DataLogin, totalTasks)
	channels.postgresQueue = make(chan localStructs.DataLogin, totalTasks)

	startWorkTasks(wg, channels, ctx)

}

func main() {
	var ctx = context.Background()
	wg := sync.WaitGroup{}
	runtime.GOMAXPROCS(2)
	runningServer(ctx, &wg)
	select {}
}
