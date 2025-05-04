package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"time"
)

type Config struct {
	Notes      string          `json:"notes"`
	Requests   []RequestConfig `json:"requests"`
	UserAgents []string        `json:"user_agents"`
}

type RequestConfig struct {
	Notes           string            `json:"notes"`
	URL             string            `json:"url"`
	Method          string            `json:"method"`
	Headers         map[string]string `json:"headers"`
	Cookies         map[string]string `json:"cookies"`
	HTTPRequestBody string            `json:"request_body"`
	HTTPResponse    *http.Response    `json:"-,omitempty"`
	RepeatCount     int               `json:"repeat_count"`
}

func (c *Config) LoadFile(f string) error {
	configFile, err := os.Open(f)
	if err != nil {
		return err
	}
	defer configFile.Close()
	decoder := json.NewDecoder(configFile)
	if err := decoder.Decode(&c); err != nil {
		return err
	}

	return nil
}

func (c *Config) CreateFile(f string) error {
	c.Notes = "Test config file"
	c.Requests = []RequestConfig{}
	c.UserAgents = []string{}

	// Example of creating a request
	var r RequestConfig
	r.Notes = "Request 1"
	r.URL = "http://localhost:8080/api/endpoint"
	r.Method = "POST"
	r.Headers = map[string]string{
		"Content-Type": "application/json",
		"Accept":       "application/json",
		"User-Agent":   "MyCustomUserAgent/1.0", // If the user agent is specified it will not use a random one from the config
	}
	r.Cookies = map[string]string{
		"session_id": "1234567890",
		"auth_token": "abcdef123456",
	}
	r.HTTPRequestBody = (`{"key": "value"}`)
	r.RepeatCount = 1
	c.Requests = append(c.Requests, r)

	// Example of creating a request
	var r2 RequestConfig
	r2.Notes = "Request 2"
	r2.URL = "http://localhost:8080/api/endpoint2"
	r2.Method = "POST"
	r2.Headers = map[string]string{
		"Content-Type": "application/json",
		"Accept":       "application/json",
	}
	r2.Cookies = map[string]string{
		"session_id": "1234567890",
		"auth_token": "abcdef123456",
	}
	r2.HTTPRequestBody = (`{"key": "value"}`)
	r2.RepeatCount = 1
	c.Requests = append(c.Requests, r2)

	// Example list of user agents
	c.UserAgents = []string{
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/136.0.0.0 Safari/537.36",
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/136.0.0.0 Safari/537.36",
		"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/136.0.0.0 Safari/537.36",
		"Mozilla/5.0 (iPhone; CPU iPhone OS 17_7 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) CriOS/136.0.7103.56 Mobile/15E148 Safari/604.1",
		"Mozilla/5.0 (iPad; CPU OS 17_7 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) CriOS/136.0.7103.56 Mobile/15E148 Safari/604.1",
	}

	jsonData, err := json.MarshalIndent(c, "", "    ")
	if err != nil {
		return err
	}

	err = os.WriteFile(f, jsonData, 0644)
	if err != nil {
		return err
	}

	return nil
}

func (r *RequestConfig) HTTPRequest(ua string) error {
	var err error
	var req *http.Request
	if r.Method == "POST" {
		req, err = http.NewRequest(r.Method, r.URL, bytes.NewBuffer([]byte(r.HTTPRequestBody)))
		if err != nil {
			return err
		}
	} else {
		req, err = http.NewRequest(r.Method, r.URL, nil)
		if err != nil {
			return err
		}
	}

	req.Header.Set("User-Agent", ua)

	for key, value := range r.Headers {
		req.Header.Set(key, value)
	}

	for key, value := range r.Cookies {
		cookie := &http.Cookie{
			Name:  key,
			Value: value,
		}
		req.AddCookie(cookie)
	}

	client := &http.Client{}
	r.HTTPResponse, err = client.Do(req)
	if err != nil {
		return err
	}

	return nil
}

// randomString returns a random string from the provided slice of strings
func randomString(strings []string) string {
	rand.New(rand.NewSource(time.Now().UnixNano()))
	return strings[rand.Intn(len(strings))]
}

func main() {
	var config Config
	configPtr := flag.String("c", "config.json", "HTTP Request Information")
	flag.Parse()

	if err := config.LoadFile(*configPtr); err != nil {
		config.CreateFile("config.json")
		log.Fatalf("Error loading request config file, created config.json\n%v\n", err)
	}

	for _, request := range config.Requests {
		for i := range request.RepeatCount {
			// Set a unique user agent from the list provided in the config
			randomUA := randomString(config.UserAgents)
			// Include in the variables passed to the HTTPRequest additional information as needed...
			if err := request.HTTPRequest(randomUA); err != nil {
				log.Fatalf("Error making HTTP request: %v\n", err)
			}

			// Output the information about the request and response conducted
			fmt.Println("----------------------------")
			fmt.Printf("Request Information: %s\n", request.Notes)
			fmt.Printf("Request URL: %s\n", request.URL)
			fmt.Printf("Response Status: %s\n", request.HTTPResponse.Status)
			fmt.Printf("Repeat Count: %d\n", i+1)
		}
	}

}
