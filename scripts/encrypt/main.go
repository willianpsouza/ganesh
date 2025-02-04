package main

import (
	"fmt"
	localEncrypt "ganesh.provengo.io/internal/encrypt"
	localSetup "ganesh.provengo.io/internal/setup"
	"runtime"
	"sync"
	"sync/atomic"
)

type PostData struct {
	worker int
	id     uint64
	data   string
}

func main() {
	runtime.GOMAXPROCS(4)
	var wg sync.WaitGroup
	var channel = make(chan PostData)
	var total uint64

	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			for j := 0; j < 100000; j++ {
				data, _ := localEncrypt.EncryptString(localSetup.PostgresURI, "5e5e83befe0b49fc5aa40b0a058a30b9")

				atomic.AddUint64(&total, 1)

				channel <- PostData{
					worker: i,
					id:     total,
					data:   data,
				}
			}
			defer wg.Done()
		}()
	}

	go func() {
		for msg := range channel {
			fmt.Println(msg.worker, msg.id, msg.data)
		}

	}()

	wg.Wait()
	close(channel)

}
