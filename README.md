# OpenAI API Load Testing Tool

A concurrent load testing tool for OpenAI's Chat API written in Go.

## Features

- Configurable number of concurrent workers
- Customizable request intervals
- Support for different OpenAI models
- Real-time status code monitoring
- Graceful shutdown with Ctrl+C
- Configurable HTTP timeouts
- Custom prompt support

## Usage

```bash
go run main.go [flags]
```

### Required Flags

- `-key string`: Your OpenAI API key (required)

### Optional Flags

- `-workers int`: Number of concurrent workers (default: 100)
- `-url string`: Base URL for API (default: "https://api.openai.com")
- `-model string`: Model name to use (default: "gpt-4o-mini")
- `-interval duration`: Interval between requests (default: 0, meaning no interval)
- `-prompt string`: Path to file containing prompt text (default: "prompt.txt")
- `-timeout duration`: HTTP client timeout (default: 0, meaning no timeout)

### Example Commands

Basic usage:
```bash
go run main.go -key "your-api-key" -url "https://api.openai.com"
```

Advanced usage:
```bash
go run main.go -key "your-api-key" -workers 50 -interval 100ms -model "gpt-3.5-turbo" -timeout 30s
```

## Output

The tool provides real-time updates showing:
- Total number of requests made
- Distribution of HTTP status codes
- Error messages for failed requests

Press Ctrl+C to gracefully shut down the tool and see the final statistics.

## Requirements

- Go 1.x or higher
- Valid OpenAI API key
- Internet connection

## Note

Ensure you have proper rate limits and API usage quotas before running intensive load tests.
