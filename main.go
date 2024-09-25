package main

import (
	"fmt"
	"sync"
	"time"

	"github.com/go-resty/resty/v2"
)

const (
	numRequests = 1000000000                   // Number of requests to send
	concurrency = 10                           // Number of concurrent requests
	apiURL      = "http://localhost:8080/ping" // Replace with your API URL
)

type Result struct {
	duration time.Duration
	err      error
}

func main() {
	client := resty.New()

	var wg sync.WaitGroup
	semaphore := make(chan struct{}, concurrency)
	results := make(chan Result, numRequests)

	start := time.Now()

	for i := 0; i < numRequests; i++ {
		wg.Add(1)
		semaphore <- struct{}{}

		go func() {
			defer wg.Done()
			defer func() { <-semaphore }()

			start := time.Now()
			resp, err := client.R().Get(apiURL)
			duration := time.Since(start)

			if err != nil {
				results <- Result{duration: duration, err: err}
				return
			}

			resp.Body()
			results <- Result{duration: duration, err: nil}
		}()
	}

	wg.Wait()
	close(results)

	totalDuration := time.Since(start)
	fmt.Printf("Total time for %d requests: %s\n", numRequests, totalDuration)

	var totalResponseTime time.Duration
	var count, errorCount int
	var minTime, maxTime time.Duration

	for result := range results {
		if result.err != nil {
			errorCount++
			continue
		}

		if minTime == 0 || result.duration < minTime {
			minTime = result.duration
		}
		if result.duration > maxTime {
			maxTime = result.duration
		}

		totalResponseTime += result.duration
		count++
	}

	averageTime := totalResponseTime / time.Duration(count)
	fmt.Printf("Total requests: %d\n", numRequests)
	fmt.Printf("Successful requests: %d\n", count)
	fmt.Printf("Failed requests: %d\n", errorCount)
	fmt.Printf("Total duration: %s\n", totalDuration)
	fmt.Printf("Average response time: %s\n", averageTime)
	fmt.Printf("Minimum response time: %s\n", minTime)
	fmt.Printf("Maximum response time: %s\n", maxTime)
}
