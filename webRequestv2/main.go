package main

import (
	"bufio"
	"bytes"
	"crypto/tls"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"
	"time"

	"golang.org/x/net/proxy"
)

type Config struct {
	Notes            string          `json:"notes"`
	Requests         []RequestConfig `json:"requests"`
	UserAgents       []string        `json:"user_agents"`
	ExcludedHeaders  []string        `json:"excluded_headers"`
	SOCKSProxy       ProxyConfig     `json:"socks_proxy"`
	HTTPProxy        ProxyConfig     `json:"http_proxy"`
	HTTPClientConfig *http.Client    `json:"-,omitempty"`
}

type RequestConfig struct {
	Notes                  string            `json:"notes"`
	URL                    string            `json:"url"`
	Method                 string            `json:"method"`
	Headers                map[string]string `json:"headers"`
	Cookies                map[string]string `json:"cookies"`
	HTTPRequestBody        string            `json:"request_body"`
	HTTPRequestBodyStrings map[string]string `json:"request_body_strings,omitempty"`
	HTTPResponse           *http.Response    `json:"-,omitempty"`
	HTTPResponseBody       []byte            `json:"response_body"`
	RepeatCount            int               `json:"repeat_count"`
}

type ProxyConfig struct {
	Enabled bool   `json:"enabled"`
	Host    string `json:"host"`
	Port    int    `json:"port"`
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
	c.ExcludedHeaders = []string{}
	c.SOCKSProxy.Enabled = false
	c.SOCKSProxy.Host = "127.0.0.1"
	c.SOCKSProxy.Port = 9000
	c.HTTPProxy.Enabled = false
	c.HTTPProxy.Host = "127.0.0.1"
	c.HTTPProxy.Port = 8080

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
	r.HTTPRequestBodyStrings = map[string]string{
		"key": "value",
	}
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

	c.ExcludedHeaders = []string{
		"Host",
		"User-Agent",
		"Cookie",
		"Connection",
		"Content-Length",
		"Upgrade-Insecure-Requests",
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

func (c *Config) SaveFile(f string) error {
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

func (c *Config) ConfigureHTTPClient() (*http.Client, error) {
	// Setup the configuration for a proxy if needed
	var client *http.Client

	if c.SOCKSProxy.Enabled && c.HTTPProxy.Enabled {
		HTTPURL := fmt.Sprintf("http://%s:%d", c.HTTPProxy.Host, c.HTTPProxy.Port)
		proxyURL, err := url.Parse(HTTPURL)
		if err != nil {
			return nil, err
		}

		SOCKSURL := fmt.Sprintf("%s:%d", c.SOCKSProxy.Host, c.SOCKSProxy.Port)
		dialer, err := proxy.SOCKS5("tcp", SOCKSURL, nil, proxy.Direct)
		if err != nil {
			return nil, err
		}

		// Create a custom transport
		transport := &http.Transport{
			Dial:            dialer.Dial,
			Proxy:           http.ProxyURL(proxyURL),
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}

		client = &http.Client{
			Transport: transport,
		}
	} else if c.SOCKSProxy.Enabled {
		SOCKSURL := fmt.Sprintf("%s:%d", c.SOCKSProxy.Host, c.SOCKSProxy.Port)
		dialer, err := proxy.SOCKS5("tcp", SOCKSURL, nil, proxy.Direct)
		if err != nil {
			return nil, err
		}

		// Create a custom transport
		transport := &http.Transport{
			Dial:            dialer.Dial,
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}

		client = &http.Client{
			Transport: transport,
		}
	} else if c.HTTPProxy.Enabled {
		HTTPURL := fmt.Sprintf("http://%s:%d", c.HTTPProxy.Host, c.HTTPProxy.Port)
		fmt.Printf("HTTP Proxy URL: %s\n", HTTPURL)
		proxyURL, err := url.Parse(HTTPURL)
		if err != nil {
			return nil, err
		}

		// Create a custom transport that uses the proxy
		transport := &http.Transport{
			Proxy:           http.ProxyURL(proxyURL),
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}

		// Create an HTTP client with the custom transport
		client = &http.Client{
			Transport: transport,
		}
	} else {
		client = &http.Client{}
	}

	return client, nil
}

func (c *Config) ParseRequestBody(rFile string, nFile string) error {
	requestFile, err := os.Open(rFile)
	if err != nil {
		return err
	}
	defer requestFile.Close()
	var requestNew RequestConfig
	var pathExtracted string
	scanner := bufio.NewScanner(requestFile)
	lastLine := ""
	for scanner.Scan() {
		line := scanner.Text()
		lastLine = line
		if line == "" {
			continue
		}
		//fmt.Println(line)
		// Get the request method and path
		if method, path, _, err := searchRegex(line, (`(?P<httpverb>GET|POST|OPTIONS)\s(?P<path>[^\s]+)\sHTTP`)); err != nil {
			return fmt.Errorf("error searching regex\n%v", err)
		} else if method != "" && path != "" {
			requestNew.Method = method
			pathExtracted = path
		}

		// Get the request headers and cookies
		if header, value, _, err := searchRegex(line, (`(?P<header>[a-zA-Z0-9\-]+)\:\s(?P<value>.+)`)); err != nil {
			return fmt.Errorf("error searching regex\n%v", err)
		} else if header != "" && value != "" {
			//fmt.Printf("Header: %s - Value: %s\n", header, value)
			if requestNew.Headers == nil {
				requestNew.Headers = make(map[string]string)
			}
			//fmt.Printf("Excluded Headers: %s\n", c.ExcludedHeaders)
			if notInList(header, c.ExcludedHeaders) {
				if header == "Origin" || header == "Host" {
					value = strings.ReplaceAll(value, "https://", "")
					value = strings.ReplaceAll(value, "http://", "")
					requestNew.Headers[header] = value
				} else {
					requestNew.Headers[header] = value
				}
			} else if header == "Cookie" {
				// Split the cookie header into individual cookies
				cookies := regexp.MustCompile(`;\s*`).Split(value, -1)
				for _, cookie := range cookies {
					parts := regexp.MustCompile(`=`).Split(cookie, 2)
					if requestNew.Cookies == nil {
						requestNew.Cookies = make(map[string]string)
					}
					if len(parts) == 2 {
						requestNew.Cookies[parts[0]] = parts[1]
					}
				}
			} else if header == "Host" || header == "Origin" {
				pathExtracted = "http://" + value + pathExtracted
			} else {
				fmt.Printf("[W] Header %s is excluded based on the config\n", header)
			}
		}

		// Assume that the last line is the request body if it exists scanning forward
		//nextLine := scanner.Scan()
		//if !nextLine && requestNew.Method == "POST" {
		//	requestNew.HTTPRequestBody = line
		//}

	}
	/**
	fmt.Printf("Request Method: %s\n", requestNew.Method)
	fmt.Printf("Request Path: %s\n", pathExtracted)
	for key, value := range requestNew.Headers {
		fmt.Printf("Request Header: %s - %s\n", key, value)
	}
	for key, value := range requestNew.Cookies {
		fmt.Printf("Request Cookie: %s - %s\n", key, value)
	}
	**/

	// If the body is plain text parse below, create additional conditions for JSON
	if lastLine != "" && requestNew.Method == "POST" {
		requestNew.HTTPRequestBody = lastLine
		values, err := url.ParseQuery(requestNew.HTTPRequestBody)
		if err != nil {
			return err
		}
		// Convert url.Values (which is map[string][]string) to map[string]string
		if requestNew.HTTPRequestBodyStrings == nil {
			requestNew.HTTPRequestBodyStrings = make(map[string]string)
		}
		for key, value := range values {
			if len(value) > 0 {
				requestNew.HTTPRequestBodyStrings[key] = value[0]
			}
		}
	}

	// Add the request to the config
	requestNew.URL = pathExtracted
	requestNew.Notes = "Added Request"
	requestNew.RepeatCount = 1
	requestNew.HTTPRequestBody = ""

	c.Requests = append(c.Requests, requestNew)
	if err := c.SaveFile(nFile); err != nil {
		return err
	}
	//fmt.Printf("Request Headers: %v\n", requestNew.Headers)
	//fmt.Printf("Verfify the request was added to the config file: %s\n", nFile)

	//config.Requests = append(config.Requests, request)
	return nil
}

func (r *RequestConfig) HTTPRequest(ua string, client *http.Client) error {
	var err error
	var req *http.Request
	//fmt.Println(r.Method)
	if r.Method == "POST" && r.HTTPRequestBody != "" {
		req, err = http.NewRequest(r.Method, r.URL, bytes.NewBuffer([]byte(r.HTTPRequestBody)))
		if err != nil {
			return err
		}
	} else if r.Method == "POST" && r.HTTPRequestBodyStrings != nil {
		// Convert to url.Values
		// Note that the HTTPRequestBodyStrings are sorted alphabetically and not in the order they were added or observed in Burp...
		values := url.Values{}
		for key, val := range r.HTTPRequestBodyStrings {
			values.Add(key, val)
		}

		// Encode to query string
		queryString := values.Encode()
		req, err = http.NewRequest(r.Method, r.URL, bytes.NewBuffer([]byte(queryString)))
		if err != nil {
			return err
		}
	} else {
		req, err = http.NewRequest(r.Method, r.URL, nil)
		if err != nil {
			return err
		}
		fmt.Println("Built new GET Request")
	}

	fmt.Println("Setup Custom User Agent")
	req.Header.Set("User-Agent", ua)

	for key, value := range r.Headers {
		req.Header.Set(key, value)
	}
	fmt.Println("Setup Custom Headers")

	for key, value := range r.Cookies {
		cookie := &http.Cookie{
			Name:  key,
			Value: value,
		}
		req.AddCookie(cookie)
	}
	fmt.Println("Setup Custom Cookies")

	//fmt.Println(req)
	r.HTTPResponse, err = client.Do(req)
	if err != nil {
		return fmt.Errorf("expected HTTP Response from Request\n%s", err)
	}

	fmt.Printf("Received HTTP Response\n%v\n", r.HTTPResponse)

	defer r.HTTPResponse.Body.Close()
	r.HTTPResponseBody, err = io.ReadAll(r.HTTPResponse.Body)
	if err != nil {
		return fmt.Errorf("error reading response body: %v", err)
	}

	return nil
}

// randomString returns a random string from the provided slice of strings
func randomString(strings []string) string {
	rand.New(rand.NewSource(time.Now().UnixNano()))
	return strings[rand.Intn(len(strings))]
}

// searchRegex searches for a regex pattern in the given byte slice and returns the first submatch
// Returnes up to 3 submatches if they exist
func searchRegex(data any, regex string) (string, string, string, error) {
	//fmt.Println(data)
	var dataBytes []byte
	switch v := data.(type) {
	case []byte:
		dataBytes = v
	case string:
		dataBytes = []byte(v)
	default:
		return "", "", "", fmt.Errorf("unsupported data type: %T", v)
	}
	re := regexp.MustCompile(regex)
	// Find submatches in byte slice
	//fmt.Printf("Data: %s\n", string(dataBytes))
	matches := re.FindSubmatch(dataBytes)

	var match1, match2, match3 string
	if matches != nil {
		//fmt.Println("No match found")
		if len(matches) > 1 {
			match1 = string(matches[1])
		}
		if len(matches) > 2 {
			match2 = string(matches[2])
		}
		if len(matches) > 3 {
			match3 = string(matches[3])
		}
		//fmt.Printf("Match 1: %s\n", match1)
		//fmt.Printf("Match 2: %s\n", match2)
	}
	return match1, match2, match3, nil
}

func notInList(value string, list []string) bool {
	for _, item := range list {
		if item == value {
			return false
		}
	}
	return true
}

func outputInformation(r RequestConfig) {
	fmt.Println("----------------------------")
	fmt.Printf("Request Information: %s\n", r.Notes)
	fmt.Printf("Request URL: %s\n", r.URL)
	fmt.Printf("Response Status: %s\n", r.HTTPResponse.Status)
	//if len(r.HTTPResponseBody) > 0 {
	//	fmt.Printf("Response Body: %s\n", string(r.HTTPResponseBody))
	//}
}

func main() {
	var config Config
	configPtr := flag.String("c", "config.json", "HTTP Request Information")
	requestPtr := flag.String("r", "", "A request placed in a text file to format for the config")
	newConfigPtr := flag.String("new", "", "The new config created from the request")
	flag.Parse()

	if err := config.LoadFile(*configPtr); err != nil {
		config.CreateFile("config.json")
		log.Fatalf("Error loading request config file, created config.json\n%v\n", err)
	}
	fmt.Println("Loaded config file")

	// Take a request from a text file and add it to the config
	if *requestPtr != "" && *newConfigPtr != "" {
		if err := config.ParseRequestBody(*requestPtr, *newConfigPtr); err != nil {
			log.Fatalf("Error parsing request body\n%v", err)
		} else {
			fmt.Printf("Request added to the config file: %s\n", *newConfigPtr)
			os.Exit(0)
		}
	} else if (*requestPtr != "" && *newConfigPtr == "") || (*requestPtr == "" && *newConfigPtr != "") {
		log.Fatalf("Error: Both request and new config file names must be provided\n")
	} else {

		client, err := config.ConfigureHTTPClient()
		if err != nil {
			log.Fatalf("Error Configuring HTTP Client: %v\n", err)
		}
		fmt.Println("Configured HTTP Client")
		// Set a unique user agent from the list provided in the config
		randomUA := randomString(config.UserAgents)

		request := config.Requests[0]
		request.Headers["Accept-Encoding"] = "identity"

		// Include in the variables passed to the HTTPRequest additional information as needed...
		fmt.Println("Making HTTP Request")
		if err = request.HTTPRequest(randomUA, client); err != nil {
			log.Fatalf("Error making HTTP request: %v\n", err)
		}

		outputInformation(request)

	}

}
