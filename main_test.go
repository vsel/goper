package main

import (
	"bytes"
	"io/ioutil"
	"net/http"
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
	roundFn := func(req *http.Request) *http.Response {
		// Test request parameters
		return &http.Response{
			StatusCode: 200,
			// Send response to be tested
			Body: ioutil.NopCloser(bytes.NewBufferString(`OK`)),
			// Must be set to non-nil value or it panics
			Header: make(http.Header),
		}
	}
	tr := RoundTripFunc(roundFn)

	sendPayload(tr, "test.com")
}