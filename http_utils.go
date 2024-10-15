package main

import (
	"io"
	"net/http"
)

func NewHttpRequest(uri string) (*http.Request, error) {
	req, err := http.NewRequest("GET", uri, nil)
	if err != nil {
		return nil, err
	}

	for k, v := range RequestHeaders {
		req.Header.Add(k, v)
	}

	return req, nil
}

func ReadHttpResponseContent(resp *http.Response) []byte {
	var buff [MaxRespLen]byte

	n, _ := io.ReadFull(resp.Body, buff[:])

	return buff[:n]
}

func GetHttpResponseHeaders(resp *http.Response) map[string]string {
	var ret = map[string]string{}

	for k, v := range resp.Header {
		ret[k] = v[0]
	}

	return ret
}
