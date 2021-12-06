package main

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"sync"
	"testing"
	"time"
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

func NewBadRespRoundTripFunc() RoundTripFunc {
	return func(req *http.Request) *http.Response {
		return nil
	}
}

func TestMakeRequest(t *testing.T) {
	url := "google.com"
	req, err := NewReq(url)
	if err != nil {
		panic(err)
	}
	client := NewTestClient(NewRoundTripFunc())
	makeRequest(client, req)
}

func TestBadResponseMakeRequest(t *testing.T) {
	url := "google.com"
	req, err := NewReq(url)
	if err != nil {
		panic(err)
	}
	client := NewTestClient(NewBadRespRoundTripFunc())
	makeRequest(client, req)
}

func TestSendPayload(t *testing.T) {
	tr := NewRoundTripFunc()
	headers := map[string]string{
		"Accept":     "text/html",
		"User-Agent": "MSIE/15.0",
	}
	body := "test"
	url := "google.com"
	sendPayload(tr, url, headers, &body)
}

func TestPayloadWorker(t *testing.T) {
	var wg sync.WaitGroup

	url := "https://google.com"
	tr := NewRoundTripFunc()
	headers := map[string]string{
		"Accept":     "text/html",
		"User-Agent": "MSIE/15.0",
	}
	body := "test"

	cancelChan := make(chan struct{})
	const repeatTimeout time.Duration = 1

	wg.Add(1)
	go payloadWorker(&wg, tr, url, headers, &body, cancelChan, repeatTimeout)
	time.Sleep(1 * time.Millisecond)
	cancelChan <- struct{}{}
	wg.Wait()
}

func TestMainFunc(t *testing.T) {
	tr := NewRoundTripFunc() // Mocked Transport

	maxGoroutines := 10
	guard := make(chan struct{}, maxGoroutines)

	var wg sync.WaitGroup

	url := "http://127.0.0.1:8080"

	headers := map[string]string{
		"Accept":     "text/html",
		"User-Agent": "MSIE/15.0",
	}

	body := "test"

	cancelChan := make(chan struct{})
	const repeatTimeout time.Duration = 1

	for i := 1; i < 11; i++ {
		guard <- struct{}{}
		wg.Add(1)
		go func() {
			payloadWorker(&wg, tr, url, headers, &body, cancelChan, repeatTimeout)
			<-guard
		}()
		if i%10 == 0 {
			killTimer := time.NewTimer(1 * time.Millisecond)
			go func() {
				<-killTimer.C
				for z := 0; z < 10; z++ {
					cancelChan <- struct{}{}
				}
			}()
		}
	}

	wg.Wait()
}

func TestRealWorldServer(t *testing.T) {
	//t.Skip()

	// Transport problem if server await sleep 5 second it waits. IT'S BAD!!!!
	tr := &http.Transport{
		MaxIdleConns:          1,
		IdleConnTimeout:       1 * time.Millisecond,
		ResponseHeaderTimeout: 1 * time.Millisecond,
		ExpectContinueTimeout: 1 * time.Millisecond,
		DisableCompression:    true,
	} // Real Transport

	maxGoroutines := 10
	guard := make(chan struct{}, maxGoroutines)

	var wg sync.WaitGroup

	url := "http://127.0.0.1:8080"

	headers := map[string]string{
		"Accept":     "text/html",
		"User-Agent": "MSIE/15.0",
	}

	body := "test"

	cancelChan := make(chan struct{})
	const repeatTimeout time.Duration = 1000

	for i := 1; i < (maxGoroutines*2)+1; i++ {
		guard <- struct{}{}
		wg.Add(1)
		go func() {
			payloadWorker(&wg, tr, url, headers, &body, cancelChan, repeatTimeout)
			<-guard
		}()
		if i%maxGoroutines == 0 {
			killTimer := time.NewTimer(3 * time.Second)
			go func() {
				<-killTimer.C
				for z := 0; z < maxGoroutines; z++ {
					cancelChan <- struct{}{}
				}
			}()
		}
	}

	wg.Wait()
}
