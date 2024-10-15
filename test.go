package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"strings"
	"sync"
)

func test_scanner() {
	// get a list of links as string from test_links.txt
	file, _ := os.Open("test_links.txt")
	defer file.Close()
	scanner := bufio.NewScanner(file)
	var links []string
	for scanner.Scan() {
		links = append(links, strings.TrimSpace(scanner.Text()))
	}

	httpreports := make([]HttpReport, len(links))
	var wg sync.WaitGroup
	for i, domain := range links {
		target, err := url.Parse(domain)
		if err != nil {
			fmt.Println("Error parsing URL", target)
			continue
		}
		if target.Scheme == "" {
			target.Scheme = "https"
		}
		wg.Add(1)
		go func(wg *sync.WaitGroup, i int, target *url.URL) {
			defer wg.Done()
			httpreports[i], err = ScanHTTP(target)
			fmt.Println("Done scanning", i, target)
		}(&wg, i, target)
	}
	wg.Wait()
	fmt.Println("Done all")
	jsonreports, _ := json.Marshal(httpreports)
	os.WriteFile("httpreports.json", jsonreports, os.ModePerm)
}
