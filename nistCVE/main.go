package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	cf "github.com/thepcn3rd/goAdvsCommonFunctions"
)

type Configuration struct {
	TeamsWebhookURL string          `json:"teamsWebhookURL"`
	DebugFile       string          `json:"debugFile"`
	Keywords        []keywordStruct `json:"keywords"`
}

type keywordStruct struct {
	Value       string `json:"value"`
	Description string `json:"description"`
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

func main() {
	var config Configuration
	var httpClient http.Client
	var httpRequest *http.Request
	var err error

	ConfigPtr := flag.String("config", "config.json", "Configuration file to load for the proxy")
	flag.Parse()

	config = loadConfig(*ConfigPtr)

	//nvdBaseURL := "http://services.nvd.nist.gov/rest/json/cves/2.0?cveId=CVE-2023-39440"

	// Gather the current directory that the binary executes from
	//path, err := os.Getwd()
	cf.CreateDirectory("/debug")
	//path := "/opt/nistCVE" // Had to hardcode the path for the crontab to work properly in linux, probably similar if executed on windows
	cf.CheckError("Unable to get the current working directory", err, true)

	// Calculate the date for the API query
	endDate := time.Now()
	startDate := endDate.AddDate(0, 0, -1)
	nvdBaseURL := "https://services.nvd.nist.gov/rest/json/cves/2.0/?pubStartDate="
	nvdBaseURL += startDate.Format("2006-01-02")
	nvdBaseURL += "T00:00:00.000&pubEndDate="
	nvdBaseURL += endDate.Format("2006-01-02")
	nvdBaseURL += "T00:00:00.000"
	//fmt.Println(nvdBaseURL)

	//teamsWebhookURL := "..."
	// Build the request for the Base URL above...
	httpRequest, err = http.NewRequest("GET", nvdBaseURL, nil)
	cf.CheckError("Unable to build http request for the NIST NVD API", err, true)

	// Receive response through the httpClient connection
	httpResponse, err := httpClient.Do(httpRequest)
	cf.CheckError("Unable to pull http response from NIST NVD API", err, true)

	// Verify we receive a 200 response and if not exit the program...
	if httpResponse.Status != "200 OK" {
		fmt.Println("Response Status: " + httpResponse.Status)
		os.Exit(0)
	}

	// Pass the httpResponse.Body to the Parser to Place it into a struct
	// Left the below lines for debugging
	//responseBody, err := io.ReadAll(httpResponse.Body)
	//fmt.Print(string(responseBody))
	nvdJSON := NVDParser(httpResponse.Body)

	// Save the JSON to a file for debugging
	debugPath := "debug/" + config.DebugFile
	outFile, err := os.Create(debugPath)
	cf.CheckError("Unable to create debug.json file", err, true)
	defer outFile.Close()
	w := bufio.NewWriter(outFile)
	jsonBuffer, err := json.MarshalIndent(nvdJSON, "", "")
	cf.CheckError("Unable to Marshall indent the JSON", err, true)
	_, err = w.WriteString(string(jsonBuffer))
	cf.CheckError("Unable to Write out the JSON Buffer", err, true)
	outFile.Sync()
	w.Flush()
	outFile.Close()

	// Total CVEs Returned Based on the Criteria
	totalResults := nvdJSON.TotalResults
	//fmt.Printf("Total Results Returned: %d\n", totalResults)

	// Open keywords file and store in a string slice
	var keywords []string
	for _, keyword := range config.Keywords {
		keywords = append(keywords, keyword.Value)
	}

	// Create a for loop to go through the CVEs available
	for i := 0; i < totalResults; i++ {
		// The following line is meant for debugging the following values being pulled from the JSON
		//fmt.Println(nvdJSON.Vulns[0].CVE.Metrics.CvssMetricV31[0].CvssData.BaseSeverity)
		// The CISA Vulnerability Name may not exist within 24 hours of publication...
		//vulnName := nvdJSON.Vulns[i].CVE.CisaVulnerabilityName
		cveID := nvdJSON.Vulns[i].CVE.ID
		// vulnStatus can be the following "Analyzed" or "Awaiting Analysis" or "Undergoing Analysis"
		vulnStatus := nvdJSON.Vulns[i].CVE.VulnStatus
		// Assumes the 1st description is English and displays the value
		cveDescription := nvdJSON.Vulns[i].CVE.Descriptions[0].Value
		cveSource := nvdJSON.Vulns[i].CVE.SourceIdentifier
		var attackComplexity string
		var cveSeverity string
		// Sometimes the CVSSMetrics V31 is nil and will abort at this location, verify that 1 exists in the array
		if len(nvdJSON.Vulns[i].CVE.Metrics.CvssMetricV31) > 1 {
			attackComplexity = nvdJSON.Vulns[i].CVE.Metrics.CvssMetricV31[0].CvssData.AttackComplexity
			cveSeverity = nvdJSON.Vulns[i].CVE.Metrics.CvssMetricV31[0].CvssData.BaseSeverity
		} else {
			attackComplexity = "Not Available"
			cveSeverity = "Not Available"
		}
		teamsMessage := "**" + cveID + "**\n\n"
		teamsMessage += "**Source: " + cveSource + "**\n\n\n"
		teamsMessage += "_Severity:_ " + cveSeverity + "  _Status:_ " + vulnStatus + "\n\n\n"
		teamsMessage += cveDescription + "\n\n\n"
		teamsMessage += "**Attack Complexity:** " + attackComplexity + "\n\n"
		teamsMessage += "**URL:** https://nvd.nist.gov/vuln/detail/" + cveID + "\n\n"
		// Place link in the adaptive card to go out and view the CVE

		// Before posting the CVE identify if the CVE applies to keywords specified
		// If a keyword matches the description more than once it will print the message more than once
		var duplicate bool
		duplicate = false
		for _, k := range keywords {
			if strings.Contains(strings.ToLower(cveDescription), strings.ToLower(k)) && duplicate == false {
				teamsMessage += "**Keyword Matched On:** " + strings.ToLower(k) + "\n\n\n"
				fmt.Printf("%s\n\n%d\n", teamsMessage, i)
				duplicate = true
				//SendTeamsMessage(teamsMessage, teamsWebhookURL)
				// Sleep 10 seconds between Teams messages sent due to the http.client timing out if it is too rapid
				time.Sleep(10 * time.Second)
			}
			duplicate = false
		}

	}
}
