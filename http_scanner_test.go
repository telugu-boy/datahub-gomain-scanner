package main

import (
	"bufio"
	"fmt"
	"net/url"
	"os"
	"strings"
	"sync"
	"testing"
)

func testScanHTTP(t *testing.T) {
	file, _ := os.Open("test_links.txt")
	defer file.Close()
	scanner := bufio.NewScanner(file)
	var domains []string
	for scanner.Scan() {
		domains = append(domains, strings.TrimSpace(scanner.Text()))
	}

	//httpreports := make([]HttpReport, len(domains))
	for i, domain := range domains {
		wg.Add(1)
		go func(wg *sync.WaitGroup, i int, domain string) {
			defer wg.Done()
			target := &url.URL{Scheme: "https", Host: domain}
			var err error
			//httpreports[i], err = ScanHTTP(target)
			_, err = ScanHTTP(target)
			if err != nil {
				fmt.Println("scan error:", i, err.Error())
			}
			fmt.Println("Done scanning", i, target)
		}(&wg, i, domain)
	}
	wg.Wait()
	fmt.Println("Done all httpscan")
	//jsonreports, _ := json.Marshal(httpreports)
	//os.WriteFile("http_reports.json", jsonreports, os.ModePerm)
}
