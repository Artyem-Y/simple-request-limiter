package main

import (
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strings"
	"sync"
	limit "test/http-handler/limiter"
	"time"
)

var RequestLimit = limit.NewRequestLimitService(10*time.Second, 100)

type HttpHandler struct {
}

type Urls struct {
	Urls string
}

func (h *HttpHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	if r.URL.Path != "/urls" {
		http.Error(w, "404 page not found", http.StatusNotFound)
		return
	}

	if r.Method != "POST" {
		http.Error(w, "Sorry, method "+r.Method+" not supported", http.StatusBadRequest)
		return
	}

	var urls Urls

	err := json.NewDecoder(r.Body).Decode(&urls)
	switch {
	case err == io.EOF:
		http.Error(w, "Body can't be empty", http.StatusBadRequest)
		return
	case err != nil:
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	urlsArr := strings.Split(urls.Urls, "\r\n")
	checkURLs(urlsArr)

	log.Printf("%.2fs elapsed", time.Since(start).Seconds())
	hashStr, err := createHashStr(urlsArr)

	if err != nil {
		http.Error(w, "Error: "+err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	resp := make(map[string]string)
	resp["message"] = "Status OK"
	resp["data"] = hashStr
	jsonResp, err := json.Marshal(resp)
	if err != nil {
		log.Fatalf("Error happened in JSON marshal. Err: %s", err)
	}

	w.Write(jsonResp)
}

func main() {
	// Define a serveMux to handle routes.
	mux := http.NewServeMux()
	httpHandler := &HttpHandler{}
	mux.Handle("/urls", Middleware(httpHandler))

	log.Println("Listening...")

	// Custom http server.
	s := &http.Server{
		Addr: ":8080",
		// Wrap the servemux with the limit middleware.
		Handler:        mux,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	if err := s.ListenAndServe(); err != nil {
		log.Fatalf("server failed to start with error %v", err.Error())
	}
}

func checkURLs(urls []string) {
	var wg sync.WaitGroup
	c := make(chan string)
	for _, url := range urls {
		wg.Add(1)
		go checkURL(url, c, &wg)
	}

	go func() {
		wg.Wait()
		close(c)
	}()

	for msg := range c {
		log.Println(msg)
	}
}

func checkURL(url string, c chan string, wg *sync.WaitGroup) {
	defer (*wg).Done()
	_, err := http.Get(url)
	if err != nil {
		log.Println("err: ", err)
		c <- url + " can not be reached"
	} else {
		c <- url + " can be reached"
	}
}

func createHashStr(urlsArr []string) (string, error) {
	var hashStr string
	var wg sync.WaitGroup

	for i, url := range urlsArr {
		// Increment the WaitGroup counter.
		wg.Add(1)
		resp, err := http.Get(url)
		if err != nil {
			return "", err
		}

		// Launch a goroutine to fetch the URL.
		go func(url string) {
			// Decrement the counter when the goroutine completes.
			defer wg.Done()
			// Fetch the URL.
			http.Get(url)
		}(url)

		// Wait for all HTTP fetches to complete.
		wg.Wait()

		if resp.StatusCode == http.StatusOK {
			bodyBytes, err := io.ReadAll(resp.Body)
			if err != nil {
				return "", err
			}

			bodyString := string(bodyBytes)
			h := sha1.New()
			h.Write([]byte(bodyString))
			sha1_hash := hex.EncodeToString(h.Sum(nil))

			if len(urlsArr) != i+1 {
				hashStr = hashStr + sha1_hash + "\r\n"
			} else {
				hashStr = hashStr + sha1_hash
			}
		}
		resp.Body.Close()
	}

	return hashStr, nil
}

func Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if RequestLimit.IsAvailable() {
			RequestLimit.Increase()
			log.Println(RequestLimit.ReqCount)
			next.ServeHTTP(w, r)
		} else {
			http.Error(w, "too many requests", http.StatusTooManyRequests)
			return
		}
	})
}
