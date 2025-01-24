package main

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"regexp"
	"time"

	elasticsearch "github.com/elastic/go-elasticsearch/v8"
	cf "github.com/thepcn3rd/goAdvsCommonFunctions"
)

type ConfigStruct struct {
	ElasticURL      string                `json:"elasticURL"`
	Username        string                `json:"username"`
	Password        string                `json:"password"`
	ElasticSettings ElasticSettingsStruct `json:"elasticSettings"`
}

type ElasticSettingsStruct struct {
	Index            string `json:"index"`
	Keywords         string `json:"keywords"`
	LookbackTimeDays int    `json:"lookbackTimeDays"`
	PageSize         int    `json:"pageSize"`
	MaxPages         int    `json:"maxPages"`
	RegexRealMessage string `json:"regexRealMessage"`
}

func main() {
	green := "\033[32m"
	//yellow := "\033[33m"
	reset := "\033[0m"

	ConfigPtr := flag.String("config", "config.json", "Configuration file to load for the proxy")
	flag.Parse()

	// Load config.json file
	var config ConfigStruct
	fmt.Printf("\n%sLoading the following config file: %s%s\n", green, *ConfigPtr, reset)
	//go logToSyslog(fmt.Sprintf("Loading the following config file: %s\n", *ConfigPtr))
	configFile, err := os.Open(*ConfigPtr)
	cf.CheckError("Unable to open the configuration file", err, true)
	defer configFile.Close()
	decoder := json.NewDecoder(configFile)
	if err := decoder.Decode(&config); err != nil {
		cf.CheckError("Unable to decode the configuration file", err, true)
	}

	cfg := elasticsearch.Config{
		Addresses: []string{
			config.ElasticURL,
		},
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
		Username: config.Username,
		Password: config.Password,
	}

	es, err := elasticsearch.NewClient(cfg)
	if err != nil {
		log.Fatalf("Error creating the client: %s", err)
	}

	// Define a simple search query
	/** Searches for the message golang dns
	query := map[string]interface{}{
		"query": map[string]interface{}{
			"match": map[string]interface{}{
				"message": "golang dns",
			},
		},
	}
	**/

	keyword := config.ElasticSettings.Keywords
	startTime := time.Now().AddDate(0, 0, config.ElasticSettings.LookbackTimeDays).Format(time.RFC3339) // 7 days ago
	endTime := time.Now().Format(time.RFC3339)                                                          // Current time

	// Construct the query
	query := map[string]interface{}{
		"query": map[string]interface{}{
			"bool": map[string]interface{}{
				"must": []map[string]interface{}{
					{
						"match": map[string]interface{}{
							"message": keyword,
						},
					},
					{
						"range": map[string]interface{}{
							"@timestamp": map[string]interface{}{
								"gte":    startTime,
								"lte":    endTime,
								"format": "strict_date_optional_time",
							},
						},
					},
				},
			},
		},
	}

	// Perform the search request
	pageSize := config.ElasticSettings.PageSize
	page := 0
	maxPages := config.ElasticSettings.MaxPages

	for {

		// Break if max pages is met
		if page >= maxPages {
			break
		}

		// Convert the query to JSON
		var buf bytes.Buffer
		if err := json.NewEncoder(&buf).Encode(query); err != nil {
			log.Fatalf("Error encoding query: %s", err)
		}
		res, err := es.Search(
			es.Search.WithContext(context.Background()),
			es.Search.WithIndex(config.ElasticSettings.Index),
			es.Search.WithBody(&buf),
			es.Search.WithFrom(page*pageSize),
			es.Search.WithSize(pageSize),
			es.Search.WithTrackTotalHits(true),
			es.Search.WithPretty(),
		)
		if err != nil {
			log.Fatalf("Error getting response: %s", err)
		}
		defer res.Body.Close()

		// Check if the response is OK
		if res.IsError() {
			log.Fatalf("Error in response: %s", res.String())
		}

		//respBody, _ := io.ReadAll(res.Body)
		//fmt.Println(string(respBody))

		// Parse and print the response
		//var result map[string]interface{}
		var result ResponseStruct
		if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
			log.Fatalf("Error parsing the response body: %s", err)
		}

		if len(result.Hits.Hits) == 0 {
			// Break out of the for loop when no hits exist...
			break
		}

		// Display hits
		re := regexp.MustCompile(`for\s([^\s]+)\sto`)
		//re := regexp.MustCompile(config.ElasticSettings.RegexRealMessage)
		for _, hits := range result.Hits.Hits {
			//fmt.Printf("%s\n", hits.Source.RealMessage)
			match := re.FindStringSubmatch(hits.Source.RealMessage)
			if match != nil {
				fmt.Println(match[1])
			} else {
				fmt.Println("No match found...")
			}
		}

		page++
	}
}
