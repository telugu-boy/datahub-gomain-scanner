package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

func ScanMatrixWellknown(target url.URL, meta *MatrixMeta) url.URL {
	target.Path = "/.well-known/matrix/client"
	req := NewHttpRequest(target.String())
	target.Path = "/"
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return target
	}
	defer resp.Body.Close()

	content := ReadHttpResponseContent(resp)

	var tmp any
	if json.Unmarshal(content, &tmp) != nil {
		return target
	}

	meta.WellknownClient = tmp

	if v, e := tmp.(map[string]any); e {
		if vv, e := v["m.homeserver"].(map[string]any); e {
			if vvv, e := vv["base_url"]; e {
				if u, e := url.Parse(fmt.Sprint(vvv)); e == nil {
					target = *u
				}
			}
		}
	}

	return target
}

func ScanMatrix(target url.URL) *MatrixMeta {
	var ret = MatrixMeta{
		ClientVersions: []string{},
		ClientFeatures: []string{}}

	target = ScanMatrixWellknown(target, &ret)

	target.Path += "/_matrix/client/versions"
	req := NewHttpRequest(target.String())
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil
	}
	defer resp.Body.Close()
	content := ReadHttpResponseContent(resp)

	var versions struct {
		Verssions        []string        `json:"versions"`
		UnstableFeatures map[string]bool `json:"unstable_features"`
	}

	if err := json.Unmarshal(content, &versions); err != nil {
		return nil
	}

	if versions.Verssions != nil {
		ret.ClientVersions = versions.Verssions
	}

	for k, v := range versions.UnstableFeatures {
		if v {
			ret.ClientFeatures = append(ret.ClientFeatures, k)
		}
	}

	return &ret
}
