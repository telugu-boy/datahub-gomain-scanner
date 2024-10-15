package main

import (
	"io"
	"net/http"
)

func NewHttpRequest(uri string) *http.Request {
	req, err := http.NewRequest("GET", uri, nil)
	AssertError(err)
	for k, v := range RequestHeaders {
		req.Header.Add(k, v)
	}

	return req
}

func ReadHttpResponseContent(resp *http.Response) []byte {
	bufsiz := min(resp.ContentLength, MaxRespLen)
	if bufsiz == -1 {
		bufsiz = MaxRespLen
	}
	buff := make([]byte, bufsiz)

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
