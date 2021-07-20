package main

import (
	"log"
	"os"
	"strconv"

	"example.com/nasa-apod-fetcher/api"
	"example.com/nasa-apod-fetcher/nasa"
)

const (
	DefaultAPIKey      = "DEMO_KEY"
	DefaultConcurrency = 5
	DefaultPort        = 8080
)

func getConfig() (apiKey string, concurrency, port int) {
	if apiKey = os.Getenv("API_KEY"); apiKey == "" {
		log.Println("Env API_KEY invalid or not set, using default value")
		apiKey = DefaultAPIKey
	}
	concurrencyRaw := os.Getenv("CONCURRENT_REQUESTS")
	if conc, err := strconv.Atoi(concurrencyRaw); err == nil && conc != 0 {
		concurrency = conc
	} else {
		log.Println("Env CONCURRENT_REQUESTS invalid or not set, using default value")
		concurrency = DefaultConcurrency
	}

	portRaw := os.Getenv("PORT")
	if p, err := strconv.Atoi(portRaw); err == nil && p > 1024 && p < 65535 { // valid ports range, skip ports reserved for system
		port = p
	} else {
		log.Println("Env PORT invalid or not set, using default value")
		port = DefaultPort
	}
	return
}

func main() {
	apiKey, concurrency, port := getConfig()
	apod := nasa.NewAPOD(apiKey, concurrency, nil)
	api := api.NewAPI(apod)

	api.Listen(port)
}
