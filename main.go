package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

func main() {
	sendPayload("https://google.com")
}
func sendPayload(url string) {
	tr := &http.Transport{
		MaxIdleConns:       10,
		IdleConnTimeout:    30 * time.Second,
		DisableCompression: true,
	}
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

	makeRequest(client, req)
}
func makeRequest(client *http.Client, req *http.Request) {
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return
	}

	defer resp.Body.Close()

	io.Copy(os.Stdout, resp.Body) //вывод в консоль
}
