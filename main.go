package main

import (
	"encoding/json"
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
	s.mux.HandleFunc("/payload", s.payload)

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
	s.log("%s %s", r.Method, r.URL.Path)
	s.mux.ServeHTTP(w, r)
}

func (s *Server) log(format string, v ...interface{}) {
	s.logger.Printf(format+"\n", v...)
}

func (s *Server) index(w http.ResponseWriter, _ *http.Request) {
	_, err := w.Write([]byte("Goper API"))
	if err != nil {
		s.logger.Printf("ListenAndServe error %s", err)
		return
	}
}

func (s *Server) payload(w http.ResponseWriter, req *http.Request) {
	decoder := json.NewDecoder(req.Body)
	var payload Payload
	err := decoder.Decode(&payload)
	if err != nil {
		panic(err)
	}
	go processPayload(
		&payload.Url,
		&payload.Headers,
		&payload.Body,
		payload.ParallelWorkersNum,
		payload.WorkDuration,
		payload.RepeatTimeout,
	)
	_, err = w.Write([]byte("Task sent to workers"))
	if err != nil {
		s.logger.Printf("ListenAndServe error %s", err)
		return
	}
	return
}

type Payload struct {
	Url                string            `json:"url"`
	Headers            map[string]string `json:"headers"`
	Body               string            `json:"body"`
	ParallelWorkersNum int               `json:"parallel_workers"`
	WorkDuration       int               `json:"duration"`
	RepeatTimeout      int               `json:"repeat_request_timeout"`
}

func processPayload(
	url *string,
	headers *map[string]string,
	body *string,
	parallelWorkersNum int,
	workDuration int,
	repeatTimeout int,
) {
	tr := &http.Transport{
		MaxIdleConns:          1,
		IdleConnTimeout:       1 * time.Millisecond,
		ResponseHeaderTimeout: 1 * time.Millisecond,
		ExpectContinueTimeout: 1 * time.Millisecond,
		DisableCompression:    true,
	}

	maxGoroutines := parallelWorkersNum
	guard := make(chan struct{}, maxGoroutines)

	var wg sync.WaitGroup

	cancelChan := make(chan struct{})

	for i := 1; i < maxGoroutines+1; i++ {
		guard <- struct{}{}
		wg.Add(1)
		go func() {
			payloadWorker(&wg, tr, *url, *headers, body, cancelChan, time.Duration(repeatTimeout))
			<-guard
		}()
		if i%maxGoroutines == 0 {
			killTimer := time.NewTimer(time.Duration(workDuration) * time.Second)
			go func() {
				<-killTimer.C
				for z := 0; z < maxGoroutines; z++ {
					cancelChan <- struct{}{}
				}
			}()
		}
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
