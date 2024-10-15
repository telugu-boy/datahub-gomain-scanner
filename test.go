package main

import (
	"bufio"
	"encoding/json"
	"fmt"
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

	scanreports := make([]ScanReport, len(links))
	var wg sync.WaitGroup
	for i, domain := range links {
		wg.Add(1)
		go func(wg *sync.WaitGroup, i int, domain string) {
			defer wg.Done()
			var err error
			scanreports[i], err = ScanDomain(domain)
			if err != nil {
				fmt.Println("scan error:", i, err.Error())
			}
			fmt.Println("Done scanning", i, domain)
		}(&wg, i, domain)
	}
	wg.Wait()
	fmt.Println("Done all")
	jsonreports, _ := json.Marshal(scanreports)
	os.WriteFile("domainreports.json", jsonreports, os.ModePerm)
}
