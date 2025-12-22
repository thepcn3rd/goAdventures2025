package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
)

/**
References:
freshness
Filters search results, when they were discovered, by date range. The following time deltas are supported.

    Day - pd - Discovered in last 24 hours.
    Week - pw - Discovered in last 7 Days.
    Month - pm - Discovered in last 31 Days.
    Year - py - Discovered in last 365 Days.

A timeframe is also supported by specifying the date range in the format YYYY-MM-DDtoYYYY-MM-DD.
**/

type Configuration struct {
	BraveURL             string              `json:"braveURL"`
	SearchKeyword        string              `json:"searchKeyword,omitempty"`
	BraveAPIKey          string              `json:"braveAPIKey"`
	BraveAPIKeyEncrypted string              `json:"braveEncryptedAPIKey"`
	BraveSettings        BraveSettingsStruct `json:"braveSettings"`
	BraveHeaders         map[string]string   `json:"braveHeaders"`
}

type BraveSettingsStruct struct {
	Freshness       string `json:"freshness"`       // pd past day, pw past week, pm past month, py past year
	ResultCount     int    `json:"resultCount"`     // number of results to return
	SafeSearch      string `json:"safeSearch"`      // off, moderate, strict
	TextDecorations string `json:"textDecorations"` // true, false
	Summary         string `json:"summary"`         // true, false
}

func (c *Configuration) CreateConfig(f string) error {
	c.BraveURL = "https://api.search.brave.com/res/v1/web/search"
	c.SearchKeyword = "Cybersecurity trends 2025"
	c.BraveAPIKey = "MYAPIKEY1234567890"
	c.BraveSettings = BraveSettingsStruct{
		Freshness:       "pw",
		ResultCount:     10,
		SafeSearch:      "off",
		TextDecorations: "false",
		Summary:         "true",
	}
	c.BraveHeaders = map[string]string{
		"Accept":               "application/json",
		"X-Subscription-Token": "",
		"User-Agent":           "golang brave search 0.1",
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

func (c *Configuration) SaveConfig(f string) error {
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

func CheckError(reasonString string, err error, exitBool bool) {
	if err != nil && exitBool {
		fmt.Printf("%s\n%v\n", reasonString, err)
		//fmt.Printf("%s\n\n", err)
		os.Exit(0)
	} else if err != nil && !exitBool {
		fmt.Printf("%s\n%v\n", reasonString, err)
		//fmt.Printf("%s\n", err)
		return
	}
}

func SaveOutputFile(message string, fileName string) {
	outFile, _ := os.Create(fileName)
	//CheckError("Unable to create txt file", err, true)
	defer outFile.Close()
	w := bufio.NewWriter(outFile)
	n, err := w.WriteString(message)
	if n < 1 {
		CheckError("Unable to save to txt file", err, true)
	}
	outFile.Sync()
	w.Flush()
	outFile.Close()
}

func main() {
	ConfigPtr := flag.String("config", "config.json", "Path to configuration file")
	flag.Parse()

	// Load the Configuration file
	var config Configuration
	configFile := *ConfigPtr
	log.Println("Loading the following config file: " + configFile + "\n")
	if err := config.LoadConfig(configFile); err != nil {
		//fmt.Println("Could not load the configuration file, creating a new default config.json")
		config.CreateConfig(configFile)
		log.Fatalf("Modify the %s file to customize how the tool functions: %v\n", configFile, err)
	}

	var braveConfig BraveConfiguration
	braveConfig.BraveURL = config.BraveURL
	braveConfig.SearchKeyword = config.SearchKeyword
	braveConfig.BraveAPIKey = config.BraveAPIKey
	braveConfig.ResultCount = config.BraveSettings.ResultCount
	braveConfig.Freshness = config.BraveSettings.Freshness
	braveConfig.SafeSearch = config.BraveSettings.SafeSearch
	braveConfig.TextDecorations = config.BraveSettings.TextDecorations
	braveConfig.Summary = config.BraveSettings.Summary
	braveConfig.RequestHeaders = config.BraveHeaders

	// Perform Brave Search
	log.Println("Performing Brave Search for the Following: " + config.SearchKeyword + "\n")
	results, err := braveConfig.SubmitRequest()
	if err != nil {
		log.Fatalf("Error performing Brave Search: %v\n", err)
	}

	// Display results
	//fmt.Println(results)
	SaveOutputFile(fmt.Sprintf("%+v", results), "results.json")

}
