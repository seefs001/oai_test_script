package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

func main() {
	// Command-line flags
	workers := flag.Int("workers", 100, "number of concurrent workers")
	baseURL := flag.String("url", "https://api.openai.com", "base URL for API")
	apiKey := flag.String("key", "", "API key for authentication")
	model := flag.String("model", "gpt-4o-mini", "model name to use")
	interval := flag.Duration("interval", 0, "interval between requests (0 for no interval)")
	promptFile := flag.String("prompt", "prompt.txt", "file containing prompt text")
	timeout := flag.Duration("timeout", 0, "HTTP client timeout (0 for no timeout)")
	flag.Parse()

	if *apiKey == "" {
		fmt.Println("Error: API key is required")
		flag.Usage()
		os.Exit(1)
	}

	// Read prompt from file
	promptBytes, err := os.ReadFile(*promptFile)
	if err != nil {
		fmt.Printf("Error reading prompt file: %v\n", err)
		os.Exit(1)
	}
	prompt := strings.TrimSpace(string(promptBytes))

	var count int64 = 0

	// Map to track HTTP status codes
	statusCodes := make(map[int]int64)
	var statusMutex sync.RWMutex

	// Create a cancellable context
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Set up channel to capture interrupt (Ctrl+C) signals
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, os.Interrupt)
	go func() {
		<-sigs
		fmt.Println("\nReceived Ctrl+C, shutting down...")
		cancel()
	}()

	// Create a shared HTTP client with optional timeout
	client := &http.Client{}
	if *timeout > 0 {
		client.Timeout = *timeout
	}

	var wg sync.WaitGroup

	// Start worker goroutines
	for i := 0; i < *workers; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()

			var ticker *time.Ticker
			if *interval > 0 {
				ticker = time.NewTicker(*interval)
				defer ticker.Stop()
			}

			for {
				if ticker != nil {
					select {
					case <-ctx.Done():
						return
					case <-ticker.C:
					}
				} else {
					if ctx.Err() != nil {
						return
					}
				}

				// Prepare the POST payload for openai chat API
				type Message struct {
					Role    string `json:"role"`
					Content string `json:"content"`
				}
				type ChatPayload struct {
					Model    string    `json:"model"`
					Messages []Message `json:"messages"`
				}
				payload, _ := json.Marshal(ChatPayload{
					Model: *model,
					Messages: []Message{
						{Role: "user", Content: prompt},
					},
				})
				req, err := http.NewRequest("POST", fmt.Sprintf("%s/v1/chat/completions", *baseURL), bytes.NewBuffer(payload))
				if err != nil {
					fmt.Printf("Worker %d: error creating request: %v\n", workerID, err)
					continue
				}
				req.Header.Set("Content-Type", "application/json")
				req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", *apiKey))

				resp, err := client.Do(req)
				if err != nil {
					fmt.Printf("Worker %d: error during request: %v\n", workerID, err)
				} else {
					statusMutex.Lock()
					statusCodes[resp.StatusCode]++
					statusMutex.Unlock()
					resp.Body.Close()
				}
				atomic.AddInt64(&count, 1)
			}
		}(i)
	}

	// Periodically print the total number of requests and status codes
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			wg.Wait()
			fmt.Printf("\nTotal requests made: %d\n", atomic.LoadInt64(&count))
			fmt.Println("Status code distribution:")
			statusMutex.RLock()
			for code, count := range statusCodes {
				fmt.Printf("HTTP %d: %d requests\n", code, count)
			}
			statusMutex.RUnlock()
			return
		case <-ticker.C:
			// Print progress updates (carriage return used to overwrite the same line)
			fmt.Printf("Total requests made so far: %d\r", atomic.LoadInt64(&count))
		}
	}
}
