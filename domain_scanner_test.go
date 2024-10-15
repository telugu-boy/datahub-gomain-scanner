package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"sync"
	"testing"
)

func TestScanDomain(t *testing.T) {
	// get a list of links as string from test_links.txt
	file, _ := os.Open("test_links.txt")
	defer file.Close()
	scanner := bufio.NewScanner(file)
	var links []string
	for scanner.Scan() {
		links = append(links, strings.TrimSpace(scanner.Text()))
	}

	//scanreports := make([]ScanReport, len(links))
	for i, domain := range links {
		wg.Add(1)
		go func(wg *sync.WaitGroup, i int, domain string) {
			defer wg.Done()
			var err error
			//scanreports[i], err = ScanDomain(domain)
			_, err = ScanDomain(domain)
			if err != nil {
				fmt.Println("scan error:", i, err.Error())
			}
			fmt.Println("Done scanning", i, domain)
		}(&wg, i, domain)
	}
	wg.Wait()
	fmt.Println("Done all domainscan")
	//jsonreports, _ := json.Marshal(scanreports)
	//os.WriteFile("domain_reports.json", jsonreports, os.ModePerm)
}
