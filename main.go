package main

import (
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"
)

func main() {
	tr := &http.Transport{
		MaxIdleConns:       10,
		IdleConnTimeout:    30 * time.Second,
		DisableCompression: true,
	}

	maxGoroutines := 10
	guard := make(chan struct{}, maxGoroutines)

	var wg sync.WaitGroup

	url := "http://127.0.0.1:8080"

	headers := map[string]string{
		"Accept":     "text/html",
		"User-Agent": "MSIE/15.0",
	}

	body := "test"

	finishChan := make(chan struct{})

	for i := 1; i < 31; i++ {
		guard <- struct{}{}
		wg.Add(1)
		go func() {
			payloadWorker(&wg, tr, url, headers, &body, finishChan)
			<-guard
		}()
		if i%10 == 0 {
			killTimer := time.NewTimer(3 * time.Second)
			go func() {
				<-killTimer.C
				for z := 0; z < 10; z++ {
					finishChan <- struct{}{}
				}
			}()
		}
	}

	wg.Wait()
}

func payloadWorker(
	wg *sync.WaitGroup,
	tr http.RoundTripper,
	url string,
	headers map[string]string,
	body *string,
	finishChan chan struct{},
) {
	defer wg.Done()

payloadLOOP:
	for {
		select {
		case <-finishChan:
			fmt.Println("killed")
			break payloadLOOP
		default:
			sendPayload(tr, url, headers, body)
		}
	}

}

func sendPayload(
	tr http.RoundTripper,
	url string,
	headers map[string]string,
	body *string,
) {
	bodyReader := strings.NewReader(*body)
	client := &http.Client{Transport: tr}
	req, err := http.NewRequest(
		"POST", url, bodyReader,
	)
	if err != nil {
		fmt.Println("NewRequest error")
		fmt.Println(err)
		return
	}
	for headerKey, headerValue := range headers {
		req.Header.Add(headerKey, headerValue)
	}

	makeRequest(client, req)
}

func makeRequest(client *http.Client, req *http.Request) {
	resp, err := client.Do(req)
	defer func(resp *http.Response) {
		if resp == nil {
			fmt.Println("no response")
			return
		}
		err := resp.Body.Close()
		if err != nil {
			fmt.Println("Body close error")
			fmt.Println(err)
			return
		}
	}(resp)

	if err != nil {
		fmt.Println("makeRequest error")
		fmt.Println(err)
		return
	}
}
