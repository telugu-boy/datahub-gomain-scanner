package main

import (
	"encoding/json"
	"net/http"
	"net/url"
	"slices"
)

var node_info_prefered_versions = []string{
	//lower index = preferred
	"http://nodeinfo.diaspora.software/ns/schema/2.1",
	"http://nodeinfo.diaspora.software/ns/schema/2.0",
	"http://nodeinfo.diaspora.software/ns/schema/1.1",
	"http://nodeinfo.diaspora.software/ns/schema/1.0",
}

func ScanNodeInfo(target url.URL) (any, any) {
	target.Path = "/.well-known/nodeinfo"
	req := NewHttpRequest(target.String())

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, nil
	}

	content := ReadHttpResponseContent(resp)
	resp.Body.Close()

	var list_ret any
	var nodeinf_list struct {
		Links []struct {
			Rel  string `json:"rel"`
			Href string `json:"href"`
		} `json:"links"`
	}

	if json.Unmarshal(content, &nodeinf_list) != nil {
		return nil, nil
	}

	// if it worked with nodeinf_list, then why it wouldn't work for _any_?
	json.Unmarshal(content, &list_ret)

	var href = ""
	var idx = int(-1) >> 1
	for _, link := range nodeinf_list.Links {
		lid := slices.Index(node_info_prefered_versions, link.Rel)
		if lid == -1 || lid < idx {
			continue
		}

		href = link.Href
		idx = lid
	}

	loc, e := url.Parse(href)
	if href == "" || e != nil || GetRedirectionType(*loc, target) != HTTP_LOCAL_REDIR {
		return list_ret, nil
	}

	node_req := NewHttpRequest(loc.String())

	resp, err = http.DefaultClient.Do(node_req)
	if err != nil {
		return list_ret, nil
	}

	content = ReadHttpResponseContent(resp)
	resp.Body.Close()

	var ret any

	if err := json.Unmarshal(content, &ret); err != nil {
		return list_ret, nil
	}

	return list_ret, ret
}
