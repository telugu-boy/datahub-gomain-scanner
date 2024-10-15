package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"os"

	"github.com/joho/godotenv"
	"github.com/valyala/fasthttp"
)

func execute() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	datahub_key := os.Getenv("DATAHUB_KEY")
	datahub_url := os.Getenv("DATAHUB_URL")
	endpoint := "/api/v1/domains/outdated"

	req := fasthttp.AcquireRequest()
	resp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseRequest(req)
	defer fasthttp.ReleaseResponse(resp)

	req.SetRequestURI("https://" + datahub_url + endpoint)
	// add bearer authorization header
	req.Header.Set("Authorization", "Bearer "+datahub_key)

	if err := fasthttp.Do(req, resp); err != nil {
		fmt.Printf("Error: %s\n", err)
		return
	}

	var domain_resp OutdatedDomainResponse
	if err := json.Unmarshal(resp.Body(), &domain_resp); err != nil {
		fmt.Println("Invalid OutdatedDomainResponse JSON")
	}

	// test scanner.go
	url, err := url.Parse("vetexplainspets.com")
	if url.Scheme == "" {
		url.Scheme = "https"
	}
	report, err := ScanHTTP(url)
	fmt.Println(report)
}

func main() {
	test_scanner()
}
