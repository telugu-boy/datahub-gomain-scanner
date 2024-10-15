package main

import (
	"bufio"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/anaskhan96/soup"
	//"github.com/jimsmart/grobotstxt"
	"github.com/valyala/fasthttp"
)

func ParseHomepage(resp *fasthttp.Response) (string, []HtmlMeta) {
	doc := soup.HTMLParse(string(resp.Body()))
	title := ""
	title_root := doc.Find("title")
	if title_root.Error == nil {
		title = title_root.Text()
	}
	//print(title)
	metas := doc.FindAll("meta")
	htmlmetas := []HtmlMeta{}
	for _, meta := range metas {
		attrs := meta.Attrs()
		if attrs["content"] != "" {
			if attrs["name"] != "" {
				htmlmetas = append(htmlmetas, HtmlMeta{
					Property: attrs["name"],
					Content:  attrs["content"],
				})
			} else if attrs["property"] != "" {
				htmlmetas = append(htmlmetas, HtmlMeta{
					Property: attrs["property"],
					Content:  attrs["content"],
				})
			}
		}
	}

	/*
		links := doc.FindAll("a")
		for _, link := range links {
			fmt.Println(link.Text(), "| Link :", link.Attrs()["href"])
		}
	*/

	return title, htmlmetas
}

func ScanRobots(target *url.URL) ([]RobotDirective, error) {
	req := fasthttp.AcquireRequest()
	resp := fasthttp.AcquireResponse()
	res := []RobotDirective{}

	req.SetRequestURI(target.Hostname() + "/robots.txt")
	for k, v := range RequestHeaders {
		req.Header.Set(k, v)
	}

	if err := fasthttp.DoTimeout(req, resp, ReqTimeout); err != nil {
		return res, err
	}

	// loop thru each line and log directives in res
	// no library
	curr_uagent := "*"
	cnt := 0
	resp_reader := resp.BodyStream()
	scanner := bufio.NewScanner(resp_reader)
	scanner.Split(bufio.ScanLines)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		if line == "" {
			continue
		}
		if strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.Split(line, ":")
		if len(parts) != 2 {
			continue
		}
		if parts[0] == "User-agent" && parts[1] != "*" {
			curr_uagent = parts[1]
			continue
		}

		res = append(res, RobotDirective{
			UserAgent: curr_uagent,
			Directive: parts[0],
			Data:      parts[1],
		})

		cnt++
		if cnt > 250 {
			break
		}
	}

	defer fasthttp.ReleaseRequest(req)
	defer fasthttp.ReleaseResponse(resp)

	return res, nil
}

func ScanHTTP(target *url.URL) (HttpReport, error) {
	report := HttpReport{}
	report.Headers = make(map[string]string)

	req := fasthttp.AcquireRequest()
	resp := fasthttp.AcquireResponse()

	//fmt.Println("Scanning", target)
	req.SetRequestURI(target.String())
	for k, v := range RequestHeaders {
		req.Header.Set(k, v)
	}

	if err := fasthttp.DoTimeout(req, resp, ReqTimeout); err != nil {
		return report, err
	}

	if resp.StatusCode() != 200 {
		return report, fmt.Errorf("Non-200 status code: %d", resp.StatusCode())
	}

	report.Path = target.Path
	report.StatusCode = resp.StatusCode()
	resp.Header.VisitAll(func(key, value []byte) {
		report.Headers[string(key)] = string(value)
	})

	report.Title, report.HtmlMeta = ParseHomepage(resp)
	robot_directives, err := ScanRobots(target)
	if err != nil {
		return report, err
	}
	report.RobotTxt = robot_directives

	fmt.Print(report.Title)

	defer fasthttp.ReleaseRequest(req)
	defer fasthttp.ReleaseResponse(resp)

	return report, nil
}

func ScanURL(target string) (ScanReport, error) {
	scanreport := ScanReport{
		Version: SCAN_REPORT_VERSION,
		Meta: map[string]any{
			"nameservers": []string{DNSServer},
			"started_at":  time.Now().Format(time.RFC3339),
		},
	}

	url, err := url.Parse(target)
	if err != nil {
		return scanreport, err
	}

	scanreport.Domain = url.Hostname()

	return scanreport, nil
}
