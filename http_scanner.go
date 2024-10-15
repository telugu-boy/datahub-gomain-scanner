package main

import (
	"bufio"
	"net/url"
	"strings"

	"github.com/anaskhan96/soup"

	//"github.com/jimsmart/grobotstxt"
	"github.com/valyala/fasthttp"
)

func AcquireAndSetupReq(uri string) *fasthttp.Request {
	req := fasthttp.AcquireRequest()
	req.SetRequestURI(uri)
	for k, v := range RequestHeaders {
		req.Header.Set(k, v)
	}
	return req
}

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

func ScanRobots(target url.URL) ([]RobotDirective, error) {
	req := fasthttp.AcquireRequest()
	resp := fasthttp.AcquireResponse()
	res := []RobotDirective{}

	defer fasthttp.ReleaseRequest(req)
	defer fasthttp.ReleaseResponse(resp)

	target.Path = "/robots.txt"
	req.SetRequestURI(target.String())
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

	return res, nil
}

func CheckHttpDumbRedirection(report *HttpReport, target url.URL) bool {
	/**
		* 3 cases of dumb redirection:
		* 1. redirecting always to the same URL
		* 2. redirecting always to another domain/schema, keeping path
		*    (www.example.com -> example.com or http://... -> https://...)
		* 3. redirecting always to https://some_site.tld/login.php?path=/amazing/path/to/that/resource.txt
		*
		* so i think the best way to avoid most false-negative is to check
		* if it redirects to an external link when reaching random pages.
		* (if it performs local redirection, then it should have some sort of logic behind it)
	**/

	target.Path = "/are_you_dumb_redirect"

	req := AcquireAndSetupReq(target.String())
	resp := fasthttp.AcquireResponse()

	defer fasthttp.ReleaseRequest(req)
	defer fasthttp.ReleaseResponse(resp)

	if err := fasthttp.DoTimeout(req, resp, ReqTimeout); err != nil {
		return false
	}

	var StatusCode = resp.StatusCode()

	if StatusCode < 300 || StatusCode > 399 {
		return false
	}

	var loc = ""
	resp.Header.VisitAll(func(key []byte, val []byte) {
		if strings.ToLower(string(key)) == "location" {
			loc = string(val)
		}
	})

	if loc == "" {
		return false
	}

	parsed_loc, _ := url.Parse(loc) // TODO: parse it better (idk if it would parse correctly "login/page.php", it would think "login" is the host)

	if (parsed_loc.Scheme == "" || parsed_loc.Scheme == target.Scheme) && (parsed_loc.Host == "" || parsed_loc.Host == target.Host) {
		return false
	}

	report.Tags = append(report.Tags, "dumb-redirect")
	return true
}

// return: continue_scan / follow redirect
func CheckHttpRedirection(resp *fasthttp.Response, target *url.URL, report *HttpReport) (bool, bool) {
	loc, e := report.Headers["Location"]

	continue_scan := true
	follow_redirect := false

	if !e {
		report.Tags = append(report.Tags, "invalid-redirect")
		return continue_scan, follow_redirect
	}

	new_loc, _ := url.Parse(loc) // TODO: parse it better (idk if it would parse correctly "login/page.php", it would think "login" is the host)

	if new_loc.Host == "" || new_loc.Host == target.Host {
		if new_loc.Scheme == "" || new_loc.Scheme == target.Scheme {
			report.Tags = append(report.Tags, "local-redirect")
			follow_redirect = true
			return true, true
		} else {
			report.Tags = append(report.Tags, new_loc.Scheme+"-redirect")
		}
	} else {
		report.Tags = append(report.Tags, "external-redirect")
	}

	return continue_scan, follow_redirect
}

func ScanHTTP(target *url.URL) (HttpReport, error) {
	report := HttpReport{}
	report.Headers = make(map[string]string)

	req := fasthttp.AcquireRequest()
	resp := fasthttp.AcquireResponse()

	defer fasthttp.ReleaseRequest(req)
	defer fasthttp.ReleaseResponse(resp)

	req.SetRequestURI(target.String())
	for k, v := range RequestHeaders {
		req.Header.Set(k, v)
	}

	// TODO: enable to connect to servers with invalid certificates
	if err := fasthttp.DoTimeout(req, resp, ReqTimeout); err != nil {
		return report, err
	}

	report.Path = target.Path
	report.StatusCode = resp.StatusCode()
	resp.Header.VisitAll(func(key, value []byte) {
		report.Headers[string(key)] = string(value)
	})

	if report.StatusCode >= 300 && report.StatusCode <= 399 {
		continue_scan, follow_redirect := CheckHttpRedirection(resp, target, &report)

		_ = follow_redirect

		if !continue_scan {
			report.Title, report.HtmlMeta = ParseHomepage(resp)
			return report, nil
		} else if follow_redirect {
			// TODO:
		}
	}

	report.Title, report.HtmlMeta = ParseHomepage(resp)
	//robot_directives, err := ScanRobots(*target)
	//if err != nil {
	//	return report, err
	//}
	//report.RobotTxt = robot_directives

	//fmt.Print(report.Title)

	return report, nil
}
