package main

import (
	"fmt"
	"net/http"
	"strings"
	"sync"
)

func main() {
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
