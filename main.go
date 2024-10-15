package main

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/valyala/fasthttp"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	datahub_key := os.Getenv("DATAHUB_KEY")
	datahub_url := os.Getenv("DATAHUB_URL")
	endpoint := "/api/v1/domains/outdated"

	req := fasthttp.AcquireRequest()
	resp := fasthttp.AcquireResponse()

	req.SetRequestURI("https://" + datahub_url + endpoint)
	// add bearer authorization header
	req.Header.Set("Authorization", "Bearer "+datahub_key)

	if err := fasthttp.Do(req, resp); err != nil {
		fmt.Printf("Error: %s\n", err)
		return
	}

	fmt.Printf("Response: %s\n", resp.Body())

	fasthttp.ReleaseRequest(req)
	fasthttp.ReleaseResponse(resp)
}
