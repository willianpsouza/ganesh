package http_client

import (
	"crypto/tls"
	"fmt"
	"github.com/google/uuid"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"sort"
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
	TimesStamp     time.Time
	TotalBytes     int
	Singnature     string
}

func httpWorker(goId int, url string, executionTime time.Duration, responseChannel chan StatusQuery, wg *sync.WaitGroup, startSignal *sync.WaitGroup) {
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

	executionStart := time.Now()
	for {
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
			QueryID:        0,
			Sequence:       0,
			StatusCode:     resp.StatusCode,
			ConnectionTime: connectionTime,
			TotalTime:      totalTime,
			TotalBytes:     bodyLength,
			Singnature:     uuid.New().String(),
			TimesStamp:     time.Now(),
		}
		if time.Since(executionStart) >= executionTime {
			break
		}
	}
}

func LogStatus(responseChannel chan StatusQuery, returnChannel chan []StatusQuery, wg *sync.WaitGroup, startSignal *sync.WaitGroup) {

	startSignal.Wait()
	var returnData []StatusQuery

	for metric := range responseChannel {
		returnData = append(returnData, metric)
	}

	returnChannel <- returnData

}

func calculateAverageAndP95(values []float64) (float64, float64, float64) {
	if len(values) == 0 {
		return 0, 0, 0
	}

	sum := 0.0
	for _, v := range values {
		sum += v
	}
	average := sum / float64(len(values))

	sort.Float64s(values)

	p95Index := int(0.95 * float64(len(values)-1))
	p95 := values[p95Index]

	p99Index := int(0.99 * float64(len(values)-1))
	p99 := values[p99Index]

	return average, p95, p99
}

func AccessMetrics(data []StatusQuery) {
	var totalDuration []float64
	var connectionTime []float64
	var totalAccess int
	var totalAccessStatusCode200 int
	var totalAccessStatusCode300 int
	var totalAccessStatusCode400 int
	var totalAccessStatusCode500 int
	minTime := time.Now()
	maxTime := time.Now()

	for _, _dt := range data {
		totalDuration = append(totalDuration, _dt.TotalTime.Seconds())
		connectionTime = append(connectionTime, _dt.ConnectionTime.Seconds())
		totalAccess++

		if _dt.StatusCode >= 300 && _dt.StatusCode <= 399 {
			totalAccessStatusCode300++
		}

		if _dt.StatusCode >= 400 && _dt.StatusCode <= 499 {
			totalAccessStatusCode400++
		}

		if _dt.StatusCode >= 500 && _dt.StatusCode <= 599 {
			totalAccessStatusCode500++
		}

		if _dt.StatusCode >= 200 && _dt.StatusCode <= 299 {
			totalAccessStatusCode200++

			if minTime == maxTime {
				maxTime = _dt.TimesStamp
			}

			if _dt.TimesStamp.UnixMilli() < minTime.UnixMilli() {
				minTime = _dt.TimesStamp
			}

			if _dt.TimesStamp.UnixMilli() > maxTime.UnixMilli() {
				maxTime = _dt.TimesStamp
			}
		}
	}
	_mean := 0.0
	_p95 := 0.0
	_p99 := 0.0

	accessRate := totalAccess / int(maxTime.Sub(minTime).Milliseconds()/1000)

	fmt.Printf("Total Access: %d Access Rate: %d Elapsed Time: %v \n", totalAccess, accessRate, maxTime.Sub(minTime))
	fmt.Printf("Status Code 200: %d \n", totalAccessStatusCode200)
	fmt.Printf("Status Code 300: %d \n", totalAccessStatusCode300)
	fmt.Printf("Status Code 400: %d \n", totalAccessStatusCode400)
	fmt.Printf("Status Code 500: %d \n", totalAccessStatusCode500)

	_mean, _p95, _p99 = calculateAverageAndP95(connectionTime)
	fmt.Printf("Total Connection Duration  Average: %f P95: %f p99: %f\n", _mean, _p95, _p99)

	_mean, _p95, _p99 = calculateAverageAndP95(totalDuration)
	fmt.Printf("Total Access Duration: Average: %f p95: %f p99: %f\n", _mean, _p95, _p99)

	return
}

func StartClient(url string, TotalWorkers int, totalTime int) {
	wg := sync.WaitGroup{}
	startSignal := sync.WaitGroup{}
	responseChannel := make(chan StatusQuery)
	returnChannel := make(chan []StatusQuery)

	startSignal.Add(1)

	go LogStatus(responseChannel, returnChannel, &wg, &startSignal)

	for i := 0; i < TotalWorkers; i++ {
		wg.Add(1)
		go httpWorker(i, url, 10*time.Second, responseChannel, &wg, &startSignal)
	}

	startSignal.Done()

	wg.Wait()
	close(responseChannel)

	metric := <-returnChannel

	AccessMetrics(metric)

}
