package http_client

import (
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"sync"
	"time"
)

type StatusQuery struct {
	Worker         int
	QueryID        int
	Sequence       int
	StatusCode     int
	ConnectionTime time.Duration
	TotalTime      time.Duration
	TotalBytes     int
}

func httpWorker(goId int, url string, totalQueries int, responseChannel chan StatusQuery, wg *sync.WaitGroup, startSignal *sync.WaitGroup) {
	defer wg.Done()

	startSignal.Wait()

	tlsConfig := &tls.Config{
		InsecureSkipVerify: true,
	}

	transport := &http.Transport{
		DialContext: (&net.Dialer{
			Timeout:   5 * time.Second,
			KeepAlive: 3 * time.Second,
		}).DialContext,
		TLSHandshakeTimeout: 5 * time.Second,
		TLSClientConfig:     tlsConfig,
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 100,
		MaxConnsPerHost:     1000,
		IdleConnTimeout:     5 * time.Second,
	}

	client := &http.Client{
		Transport: transport,
		Timeout:   5 * time.Second,
	}

	for i := 0; i < totalQueries; i++ {
		startTime := time.Now()
		bodyLength := 0

		resp, err := client.Get(url)
		connectionTime := time.Since(startTime)
		if err != nil {
			log.Println("Error:", err)
		} else {
			if resp.StatusCode >= 200 && resp.StatusCode <= 299 {
				body, _ := ioutil.ReadAll(resp.Body)
				bodyLength = len(body)
				_ = resp.Body.Close()
			}
		}
		totalTime := time.Since(startTime)

		responseChannel <- StatusQuery{
			Worker:         goId,
			QueryID:        i,
			Sequence:       i,
			StatusCode:     resp.StatusCode,
			ConnectionTime: connectionTime,
			TotalTime:      totalTime,
			TotalBytes:     bodyLength,
		}

	}
}

func LogStatus(responseChannel chan StatusQuery, wg *sync.WaitGroup, startSignal *sync.WaitGroup) {
	defer wg.Done()
	startSignal.Wait()
	fmt.Printf("Starting log consumer")

	defaultPrint := "Worker id %d -- Query %d ConnTime %v TotalTime %v TotalBytes %v"

	func() {
		for metric := range responseChannel {
			fmt.Printf(defaultPrint, metric.Worker, metric.QueryID, metric.Sequence, metric.TotalTime, metric.TotalBytes)
		}
	}()
}

func StartClient(url string, totalQueries int, TotalWorkers int) {
	wg := sync.WaitGroup{}
	startSignal := sync.WaitGroup{}
	responseChan := make(chan StatusQuery)

	startSignal.Add(1)

	wg.Add(1)
	go LogStatus(responseChan, &wg, &startSignal)

	for i := 0; i < TotalWorkers; i++ {
		wg.Add(1)
		go httpWorker(i, url, totalQueries, responseChan, &wg, &startSignal)
		fmt.Printf("Worker %d started\n", i)
	}

	startSignal.Done()

	go func() {
		wg.Wait()
		close(responseChan)
	}()
}
