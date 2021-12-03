package main

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"sync"
	"testing"
)

// RoundTripFunc .
type RoundTripFunc func(req *http.Request) *http.Response

// RoundTrip .
func (f RoundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req), nil
}

//NewTestClient returns *http.Client with Transport replaced to avoid making real calls
func NewTestClient(fn RoundTripFunc) *http.Client {
	return &http.Client{
		Transport: fn,
	}
}

//NewReq returns *http.Request or error for testing different requests
func NewReq(url string) (*http.Request, error) {
	req, err := http.NewRequest(
		"GET", url, nil,
	)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Accept", "text/html")
	req.Header.Add("User-Agent", "MSIE/15.0")
	return req, nil
}

func NewRoundTripFunc() RoundTripFunc {
	return func(req *http.Request) *http.Response {
		return &http.Response{
			StatusCode: 200,
			Body:       ioutil.NopCloser(bytes.NewBufferString(`OK`)),
			Header:     make(http.Header),
		}
	}
}

func TestMakeRequest(t *testing.T) {
	url := "google.com"
	req, err := NewReq(url)
	if err != nil {
		panic(err)
	}
	client := NewTestClient(func(req *http.Request) *http.Response {
		// Test request parameters
		return &http.Response{
			StatusCode: 200,
			// Send response to be tested
			Body: ioutil.NopCloser(bytes.NewBufferString(`OK`)),
			// Must be set to non-nil value or it panics
			Header: make(http.Header),
		}
	})
	makeRequest(client, req)
}

func TestSendPayload(t *testing.T) {
	tr := NewRoundTripFunc()
	sendPayload(tr, "test.com")
}

func TestPayloadWorker(t *testing.T) {
	var wg sync.WaitGroup

	url := "https://google.com"
	tr := NewRoundTripFunc()

	wg.Add(1)
	go payloadWorker(&wg, tr, url)
	wg.Wait()
}
