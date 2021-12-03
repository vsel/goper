package main

import (
	"fmt"
	"net/http"
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

	url := "https://google.com"

	for i := 0; i < 30; i++ {
		guard <- struct{}{}
		go func() {
			wg.Add(1)
			payloadWorker(&wg, tr, url)
			<-guard
		}()
	}

	wg.Wait()
}

func payloadWorker(wg *sync.WaitGroup, tr http.RoundTripper, url string) {
	defer wg.Done()
	sendPayload(tr, url)
}

func sendPayload(tr http.RoundTripper, url string) {

	client := &http.Client{Transport: tr}
	req, err := http.NewRequest(
		"GET", url, nil,
	)
	if err != nil {
		fmt.Println("NewRequest error")
		fmt.Println(err)
		return
	}

	req.Header.Add("Accept", "text/html")
	req.Header.Add("User-Agent", "MSIE/15.0")

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

	fmt.Println(&client)
}
