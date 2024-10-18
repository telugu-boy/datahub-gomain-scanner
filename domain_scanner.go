package main

import (
	"errors"
	"fmt"
	"net/url"
	"os"
	"strings"
	"time"

	"golang.org/x/net/idna"
)

var ErrHiddenService = errors.New("cannot scan hidden service")

func IsIgnorableError(err error) bool {
	if err == nil {
		return true
	}
	return os.IsTimeout(err)
}

func ScanDomain(domain string) (ScanReport, error) {
	scanreport := ScanReport{
		Version: SCAN_REPORT_VERSION,
		Domain:  domain,
		Meta: map[string]any{
			"nameservers": []string{DNSServer},
			"started_at":  time.Now().UTC().Format(SCAN_REPORT_TIME_FORMAT),
		},
		Services:    map[string]any{},
		HttpReports: map[string]HttpReport{},
		Records:     map[string][]string{},
	}

	defer func() {
		scanreport.Meta["ended_at"] = time.Now().UTC().Format(SCAN_REPORT_TIME_FORMAT)
	}()

	domain, err := idna.ToASCII(strings.ToLower(strings.Trim(strings.TrimSpace(domain), ".")))
	if err != nil {
		return scanreport, err
	}

	if strings.HasSuffix(domain, ".onion") ||
		strings.HasSuffix(domain, ".i2p") ||
		strings.HasSuffix(domain, ".bit") {
		return scanreport, ErrHiddenService
	}

	scanreport.Records = EnumerateDnsRecords(domain)

	// no need to http scan if there is no A or AAAA records
	if len(scanreport.Records["A"]) != 0 || len(scanreport.Records["AAAA"]) != 0 {
		for _, scheme := range []string{"http", "https"} {
			var target = &url.URL{Scheme: scheme, Host: domain}
			report, err := ScanHTTP(target)

			if err != nil && !IsIgnorableError(err) {
				fmt.Println("error while scanning", domain, target.String(), scheme, ":", err.Error())
			} else if err == nil {
				scanreport.HttpReports[scheme] = report
			}
		}
	}

	return scanreport, nil
}
