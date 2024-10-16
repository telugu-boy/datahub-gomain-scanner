package main

import (
	"bufio"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/anaskhan96/soup"
	//"github.com/jimsmart/grobotstxt"
)

var ErrExternalRedirect = errors.New("external redirection")
var ErrTooManyRedir = errors.New("too many redirections")

func ParseHomepage(resp *http.Response) (string, []HtmlMeta) {
	doc := soup.HTMLParse(string(ReadHttpResponseContent(resp)))
	title := ""
	title_root := doc.Find("title")
	if title_root.Error == nil {
		title = title_root.Text()
	}
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

func ScanRobots(target *url.URL) []RobotDirective {
	res := []RobotDirective{}

	target.Path = "/robots.txt"
	req := NewHttpRequest(target.String())

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return []RobotDirective{}
	}
	defer resp.Body.Close()

	// loop thru each line and log directives in res
	// no library
	curr_uagent := "*"
	cnt := 0
	scanner := bufio.NewScanner(resp.Body)
	scanner.Split(bufio.ScanLines)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		line = strings.Split(line, "#")[0]
		if line == "" {
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

	return res
}

const (
	HTTP_LOCAL_REDIR  = 1
	HTTP_SCHEME_REDIR = 2
	HTTP_EXTERN_REDIR = 3
)

func GetRedirectionType(loc *url.URL, target *url.URL) int {
	if loc.Host == target.Host {
		if loc.Scheme == target.Scheme {
			return HTTP_LOCAL_REDIR
		} else {
			return HTTP_SCHEME_REDIR
		}
	} else {
		return HTTP_EXTERN_REDIR
	}
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
	req := NewHttpRequest(target.String())

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	if resp.StatusCode < 300 || resp.StatusCode > 399 {
		return false
	}

	loc, e := resp.Location()
	if e != nil {
		return false
	}

	return GetRedirectionType(loc, &target) != HTTP_LOCAL_REDIR
}

// return: continue_scan / follow redirect
func CheckHttpRedirection(resp *http.Response, target *url.URL, report *HttpReport) (bool, bool) {
	loc, e := resp.Location()

	if e != nil {
		report.Tags = append(report.Tags, "invalid-redirect")
		return true, false
	}

	redir_type := GetRedirectionType(loc, target)

	switch redir_type {
	case HTTP_LOCAL_REDIR:
		report.Tags = append(report.Tags, "local-redirect")
		return true, true
	case HTTP_SCHEME_REDIR:
		report.Tags = append(report.Tags, loc.Scheme+"-redirect")
	case HTTP_EXTERN_REDIR:
		report.Tags = append(report.Tags, "external-redirect")
	}

	return CheckHttpDumbRedirection(report, *target), false
}

func FollowLocalRedirections(resp *http.Response, target url.URL) (*http.Response, error) {
	is_defered := true
	if _, e := resp.Location(); e != nil {
		return resp, nil
	}
	for i := 0; i < MaxRedir; i++ {
		loc, _ := resp.Location()
		if loc.Scheme != target.Scheme || loc.Host != target.Host {
			if !is_defered {
				resp.Body.Close()
			}
			return nil, ErrExternalRedirect
		}

		req := NewHttpRequest(loc.String())
		if !is_defered {
			resp.Body.Close()
		}
		// query it
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return nil, err
		}
		is_defered = false
		if resp.Header.Get("Location") == "" {
			return resp, nil
		}
	}

	if !is_defered {
		resp.Body.Close()
	}
	return nil, ErrTooManyRedir
}

func ScanHTTP(target *url.URL) (HttpReport, error) {
	report := HttpReport{}
	report.Headers = make(map[string]string)

	req := NewHttpRequest(target.String())

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return report, err
	}
	defer resp.Body.Close()

	report.Path = "/"
	if resp.StatusCode >= 300 && resp.StatusCode <= 399 {
		continue_scan, follow_redirect := CheckHttpRedirection(resp, target, &report)

		_ = follow_redirect

		if !continue_scan {
			report.StatusCode = resp.StatusCode
			report.Headers = GetHttpResponseHeaders(resp)
			report.Title, report.HtmlMeta = ParseHomepage(resp)
			return report, nil
		} else if follow_redirect {
			/* if there is an error */
			report.StatusCode = resp.StatusCode
			report.Headers = GetHttpResponseHeaders(resp)

			resp, err = FollowLocalRedirections(resp, *target)
			if err != nil {
				fmt.Println("redirection error:", err.Error())
				return report, nil
			}
			defer resp.Body.Close()
		}
	}

	report.StatusCode = resp.StatusCode
	report.Headers = GetHttpResponseHeaders(resp)

	report.Title, report.HtmlMeta = ParseHomepage(resp)
	robot_directives := ScanRobots(target)
	report.RobotsTxt = robot_directives

	report.Matrix = ScanMatrix(*target)

	//fmt.Print(report.Title)

	return report, nil
}
