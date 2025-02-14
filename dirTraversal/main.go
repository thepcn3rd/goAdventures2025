package main

import (
	"bufio"
	"crypto/tls"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	cf "github.com/thepcn3rd/goAdvsCommonFunctions"
)

// /home/thepcn3rd/hackthebox/SecLists/Fuzzing/LFI

// Future Enhancements
// Create the part that reads the payloadType for reading in a file does not exist (Currently only uses the payloadList)
// Currently supports the GET method but not the POST

type Configuration struct {
	DestinationURL  string           `json:"destinationURL"`
	Method          string           `json:"method"`
	UserAgent       string           `json:"userAgent"`
	DelayBetweenReq int              `json:"delayBetweenRequests"`
	TargetKey       string           `json:"targetKey"`
	KeyValuePairs   []KeyValueStruct `json:"keyValuePairs"`
	OutputDirectory string           `json:"outputDirectory"`
	PayloadType     string           `json:"payloadType"`
	PayloadFilename string           `json:"payloadFilename"`
	PayloadList     []string         `json:"payloadList"`
	BodySize        int              `json:"bodySize"`
}

type KeyValueStruct struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

func loadConfig(cPtr string) Configuration {

	var c Configuration
	fmt.Println("Loading the following config file: " + cPtr + "\n")
	// go logToSyslog(fmt.Sprintf("Loading the following config file: %s\n", *ConfigPtr))
	configFile, err := os.Open(cPtr)
	cf.CheckError("Unable to open the configuration file", err, true)
	defer configFile.Close()
	decoder := json.NewDecoder(configFile)
	// var config Configuration
	if err := decoder.Decode(&c); err != nil {
		cf.CheckError("Unable to decode the configuration file", err, true)
	}

	return c
}

func createPayloadListFromFile(filename string) []string {
	var payloadList []string
	file, err := os.Open(filename)
	cf.CheckError("Unable to open the file specified in config as payloadListFile", err, true)
	defer file.Close()

	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := scanner.Text()
		line = strings.Replace(line, "\n", "", -1)
		line = strings.Replace(line, "\r", "", -1)
		if len(line) > 0 {
			payloadList = append(payloadList, line)
		}
	}

	return payloadList
}

func displayTimeStamp() {
	currentTime := time.Now()
	formattedTime := currentTime.Format("2006-01-02 15:15:15")
	fmt.Println("Date/Time: ", formattedTime)
}

func main() {
	var config Configuration
	ConfigPtr := flag.String("config", "config.json", "Configuration file to load for the proxy")
	flag.Parse()

	config = loadConfig(*ConfigPtr)

	// Verify the URL is correctly formatted
	baseURL, err := url.Parse(config.DestinationURL)
	if err != nil {
		log.Fatalf("Failed to parse target URL in config.json: %v\n", err)
	}

	// Add a timeout to the client of 10 seconds
	client := http.Client{Timeout: 10 * time.Second}
	client.Get(config.DestinationURL)
	cf.CreateDirectory("/" + config.OutputDirectory)
	// For the design refer to the headlessScreenshot program
	if config.PayloadType == "File" {
		config.PayloadList = createPayloadListFromFile(config.PayloadFilename)
	}

	displayTimeStamp()

	for count, payload := range config.PayloadList {

		query := baseURL.Query()
		query.Set(config.TargetKey, payload)
		for _, item := range config.KeyValuePairs {
			if item.Key != "" {
				query.Set(item.Key, item.Value)
			}
		}
		// The below URL Encodes the strings and the payload
		//baseURL.RawQuery = query.Encode()
		var queryParts []string
		for key, values := range query {
			for _, value := range values {
				queryParts = append(queryParts, fmt.Sprintf("%s=%s", key, value))
			}
		}

		queryString := strings.Join(queryParts, "&")
		urlString := config.DestinationURL + "?" + queryString
		// Manually construct the query string
		//values := url.Values{}
		//values.Add(config.TargetKey, payload)

		// Skip the verification of a self-signed certificate
		http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
		request, err := http.NewRequest(config.Method, urlString, nil)
		cf.CheckError("Unable to prepare HTTP GET Request", err, true)
		request.Header.Set("User-Agent", config.UserAgent)

		client := &http.Client{}
		//response, err := http.Get(baseURL.String())
		response, err := client.Do(request)
		if err != nil {
			log.Printf("Error visiting URL with the payload %s: %v\n\n", baseURL.String(), err)
			continue
		}
		defer response.Body.Close()

		body, err := io.ReadAll(response.Body)
		if err != nil {
			log.Fatalf("Failed to read the response body: %v", err)
		}

		// Measure the size of the body
		bodySize := len(body)

		///fmt.Printf("%d: Testing %s - Status: %d - Length: %d\n", count, baseURL.String(), response.StatusCode, bodySize)
		if bodySize > config.BodySize {
			now := time.Now()
			formattedDate := now.Format("01-02-2006_15:04")
			fileName := config.OutputDirectory + "/saved_" + strconv.Itoa(count) + "_" + formattedDate + ".output"
			cf.SaveOutputFile(string(body), fileName)
		}

		// Introduce a delay to not cause a Denial of Service
		time.Sleep(time.Duration(config.DelayBetweenReq) * time.Second)
	}

	displayTimeStamp()

	// Start testing directory traversal
	//testDirectoryTraversal(targetURL)
}
