package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

func main() {
	tr := &http.Transport{
		MaxIdleConns:       10,
		IdleConnTimeout:    30 * time.Second,
		DisableCompression: true,
	}
	client := &http.Client{Transport: tr}
	req, err := http.NewRequest(
		"GET", "https://google.com", nil,
	)
	if err != nil {
		fmt.Println(err)
		return
	}

	req.Header.Add("Accept", "text/html")
	req.Header.Add("User-Agent", "MSIE/15.0")
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return
	}

	defer resp.Body.Close()

	io.Copy(os.Stdout, resp.Body) //вывод в консоль
}