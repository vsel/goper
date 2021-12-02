package main

import (
	"fmt"
	"net/http"
	"time"
)

func main() {
	tr := &http.Transport{
		MaxIdleConns:       10,
		IdleConnTimeout:    30 * time.Second,
		DisableCompression: true,
	}
	SendPayload(tr, "https://google.com")
}

func SendPayload(tr http.RoundTripper, url string) {

	client := &http.Client{Transport: tr}
	req, err := http.NewRequest(
		"GET", url, nil,
	)
	if err != nil {
		fmt.Println(err)
		return
	}

	req.Header.Add("Accept", "text/html")
	req.Header.Add("User-Agent", "MSIE/15.0")

	MakeRequest(client, req)
	MakeRequest(client, req)
	MakeRequest(client, req)
}

func MakeRequest(client *http.Client, req *http.Request) {
	resp, err := client.Do(req)
	defer resp.Body.Close()

	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println(&client)
}
