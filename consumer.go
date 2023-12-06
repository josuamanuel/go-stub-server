package main

import (
	"log"
	"net/http"
	"sync"
	"sync/atomic"
	"time"
	"github.com/go-resty/resty/v2"
)

func main() {
	// Set the target URL
	url := "http://localhost:8080/card-list/"

	// Set the desired TPS and duration
	tps := 500
	duration := 1 * time.Minute

	// Create a new Resty client
	client := resty.New()

	// Create counters for successful and failed transactions
	var successCount uint64
	var failureCount uint64

	// Create a timer to track the duration
	timer := time.NewTimer(duration)

	// Create a wait group to synchronize the goroutines
	var wg sync.WaitGroup

	// Calculate the interval between requests based on the desired TPS
	interval := time.Duration(int(time.Second) / tps)

	// Start a goroutine to log the metrics
	go func() {
		defer wg.Done()

		// Initialize variables for metrics
		var totalTime time.Duration
		var count int

		// Start a loop to collect metrics
		for {
			// Collect metrics every second
			time.Sleep(time.Second)

			// Get the current counts
			successful := atomic.LoadUint64(&successCount)
			failed := atomic.LoadUint64(&failureCount)

			// Calculate the total count
			total := successful + failed

			// Calculate the total time
			totalTime = time.Duration(total) * time.Second

			// Update the count variable
			count = int(total)

			// Check if the timer has expired
			select {
			case <-timer.C:
				// When the timer expires, calculate and log the metrics
				avgResponseTime := totalTime / time.Duration(count)
				log.Printf("Metrics:\n")
				log.Printf("  Successful transactions: %d\n", successful)
				log.Printf("  Failed transactions: %d\n", failed)
				log.Printf("  Average response time: %v\n", avgResponseTime)
				return
			default:
				// Continue collecting metrics
			}
		}
	}()

	// Start a loop to generate traffic
	for {
		// Send a request in a goroutine
		wg.Add(1)
		go func() {
			defer wg.Done()

			// Send the request and measure the response time
			startTime := time.Now()
			resp, err := client.R().Get(url)
			responseTime := time.Since(startTime)

			// Check if the request was successful and update the counters
			if err == nil && resp.StatusCode() == http.StatusOK {
				atomic.AddUint64(&successCount, 1)
			} else {
				atomic.AddUint64(&failureCount, 1)
			}

			// Log the response time
			log.Printf("Response time: %v\n", responseTime)
		}()

		// Sleep for the interval before sending the next request
		time.Sleep(interval)

		// Check if the timer has expired
		select {
		case <-timer.C:
			// When the timer expires, stop generating traffic and wait for the metrics to be logged
			wg.Wait()
			return
		default:
			// Continue generating traffic
		}
	}
}