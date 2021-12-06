package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

func main() {
	logger := log.New(os.Stdout, "", 0)

	hs := setup(logger)

	logger.Printf("Listening on http://0.0.0.0%s\n", hs.Addr)

	err := hs.ListenAndServe()
	if err != nil {
		logger.Printf("ListenAndServe error %s", err)
	}
}

func setup(logger *log.Logger) *http.Server {
	return &http.Server{
		Addr:         getAddr(),
		Handler:      newServer(logWith(logger)),
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}
}

func getAddr() string {
	if port := os.Getenv("PORT"); port != "" {
		return ":" + port
	}

	return ":8383"
}

func newServer(options ...Option) *Server {
	s := &Server{logger: log.New(ioutil.Discard, "", 0)}

	for _, o := range options {
		o(s)
	}

	s.mux = http.NewServeMux()
	s.mux.HandleFunc("/", s.index)

	return s
}

type Option func(*Server)

func logWith(logger *log.Logger) Option {
	return func(s *Server) {
		s.logger = logger
	}
}

type Server struct {
	mux    *http.ServeMux
	logger *log.Logger
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	s.log("%s %s", r.Method, r.URL.Path)

	s.mux.ServeHTTP(w, r)
}

func (s *Server) log(format string, v ...interface{}) {
	s.logger.Printf(format+"\n", v...)
}

func (s *Server) index(w http.ResponseWriter, _ *http.Request) {
	_, err := w.Write([]byte("Hello, world!"))
	if err != nil {
		s.logger.Printf("ListenAndServe error %s", err)
		return
	}
}

func payloadWorker(
	wg *sync.WaitGroup,
	tr http.RoundTripper,
	url string,
	headers map[string]string,
	body *string,
	cancelChan chan struct{},
	repeatTimeout time.Duration,
) {
	defer wg.Done()
	limitTicker := time.NewTicker(repeatTimeout * time.Millisecond)
	defer limitTicker.Stop()

payloadLOOP:
	for {
		select {
		case <-cancelChan:
			fmt.Println("killed")
			break payloadLOOP
		case <-limitTicker.C:
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
		fmt.Println("NewRequest error", err)
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
			fmt.Println("Body close error", err)
			return
		}
	}(resp)

	if err != nil {
		fmt.Println("makeRequest error", err)
		return
	}
}
