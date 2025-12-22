package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"

	"nistCVEv2/brave"
	"nistCVEv2/nvd"
)

type BraveSearchResults struct {
	Title           string
	URL             string
	Description     string
	MatchesCVETopic bool
}

type Configuration struct {
	BraveAPIKey    string `json:"braveAPIKey"`
	OllamaURL      string `json:"ollamaURL"`
	OllamaWaitTime int    `json:"ollamaWaitTime"`
}

func (c *Configuration) CreateConfig(f string) error {
	c.BraveAPIKey = ""
	c.OllamaURL = "http://localhost:11434/api/chat"
	c.OllamaWaitTime = 10 // HTTP Waittime for a response from ollama in minutes

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

func SaveOutputFile(message string, fileName string) {
	outFile, _ := os.Create(fileName)
	//CheckError("Unable to create txt file", err, true)
	defer outFile.Close()
	w := bufio.NewWriter(outFile)
	n, err := w.WriteString(message)
	if n < 1 {
		log.Printf("unable to save to txt file: %v", err)
		os.Exit(0)
	}
	outFile.Sync()
	w.Flush()
	outFile.Close()
}

// Load the Configuration file
var config Configuration

func AppendIfNotExists(slice []string, element string) []string {
	// Create a map for efficient lookup
	exists := false
	for _, item := range slice {
		if item == element {
			exists = true
			break
		}
	}
	// Check if element exists
	if !exists {
		return append(slice, element)
	}
	return slice
}

func main() {
	ConfigPtr := flag.String("config", "config.json", "Path to configuration file")
	flag.Parse()

	configFile := *ConfigPtr
	log.Println("Loading the following config file: " + configFile + "\n")
	if err := config.LoadConfig(configFile); err != nil {
		//fmt.Println("Could not load the configuration file, creating a new default config.json")
		config.CreateConfig(configFile)
		log.Fatalf("Modify the %s file to customize how the tool functions: %v\n", configFile, err)
	}

	// --------------------------------------- Pull CVE Information from NIST ---------
	// Pull NVD CVE Data for the specified timeframe

	NVDConfig := nvd.NVDSearchConfiguration{
		NVDURL:           "https://services.nvd.nist.gov/rest/json/cves/2.0/",
		Timeframe:        1, // Look back 1 day for CVEs
		PullFromXDaysAgo: 6, // Do you want to pull yesterdays CVE with a 0, or pull going back 3 days
	}

	nvdJSON, err := NVDConfig.SubmitRequest()
	if err != nil {
		log.Fatalf("Error submitting NVD request: %v\n", err)
	}

	fmt.Println(nvdJSON.TotalResults, "Total CVE(s) found in the specified timeframe.")

	nvdOutputAll := CreateBasicNVDStruct(nvdJSON)
	nvdOutputAll = FilterBasicNVDStruct(nvdOutputAll)
	fmt.Printf("Total CVE(s) after filtering: %d\n", len(nvdOutputAll))

	// ----------------------------- Search Brave -------------------------
	// ------------------ Specific Use Cases -----------------------
	/**
	var nvdOutputAll []NVDOutputStruct
	var nvdOutput NVDOutputStruct
	// Test case with CVE that has been in the news
	nvdOutput.CVEID = "CVE-2025-64446"
	nvdOutput.Severity = "CRITICAL"
	nvdOutput.VulnStatus = "Analyzed"
	nvdOutput.Description = "Description: A relative path traversal vulnerability in Fortinet FortiWeb 8.0.0 through 8.0.1, FortiWeb 7.6.0 through 7.6.4, FortiWeb 7.4.0 through 7.4.9, FortiWeb 7.2.0 through 7.2.11, FortiWeb 7.0.0 through 7.0.11 may allow an attacker to execute administrative commands on the system via crafted HTTP or HTTPS requests."
	nvdOutputAll = append(nvdOutputAll, nvdOutput)
	// Test case with a CVE that is not in the news at this time...
	/**
	nvdOutput.CVEID = "CVE-2024-9126"
	nvdOutput.Severity = "HIGH"
	nvdOutput.VulnStatus = "Analyzed"
	nvdOutput.Description = "Use after free in Internals in Google Chrome on iOS prior to 127.0.6533.88 allowed a remote attacker who convinced a user to engage in specific UI gestures to potentially exploit heap corruption via a series of curated UI gestures. (Chromium security severity: Medium)"
	nvdOutputAll = append(nvdOutputAll, nvdOutput)
	**/
	braveConfig := brave.BraveConfiguration{
		BraveURL: "https://api.search.brave.com/res/v1/web/search",
		//SearchKeyword: "\"" + nvdEntry.CVEID + "\"",
		SearchKeyword:   "", // Adds the search keyword in the first line of the for loop...
		BraveAPIKey:     config.BraveAPIKey,
		ResultCount:     10,   // Could set this back to 10
		Freshness:       "pm", // pd past day, pm past month
		SafeSearch:      "off",
		TextDecorations: "false",
		Summary:         "true",
		RequestHeaders: map[string]string{
			"X-Subscription-Token": "", // To be populated from config
			"Accept":               "application/json",
			"User-Agent":           "golang brave search 0.1",
		},
	}

	nvdOutputAll, err = BraveGetSearchResults(nvdOutputAll, braveConfig)
	if err != nil {
		log.Fatalf("Error getting Brave search results: %v\n", err)
	}
	// ---------------------------------- Agentic AI Verify Search Results ------------------------------
	// Ask an agent if the results of a topic and a search result match
	// Compare the results of 3 agents using different models to determine a match
	// References: https://google.github.io/adk-docs/agents/workflow-agents/loop-agents/#full-example-iterative-document-improvement
	/** Idea
		Using a loop agent at this time could not get it to work... 11/23

	**/

	nvdOutputAll, err = OllamaVerifySearchResults(nvdOutputAll)
	if err != nil {
		log.Fatalf("Error verifying search results with Ollama: %v\n", err)
	}

	for nvdCount, nvd := range nvdOutputAll {
		matchCount := 0
		for _, result := range nvd.BraveSearchResults {
			if result.MatchesCVETopic {
				matchCount++
			}
		}
		nvdOutputAll[nvdCount].BraveSearchMatches = matchCount
		// You should store this and then skip these in the Create Summaries and Output Results
		fmt.Printf("CVE: %s - Total Brave Search Results Matching CVE Topic: %d\n", nvd.CVEID, matchCount)
	}

	// ----------------------------- Agentic AI Create AI Summary of Results -------------------------
	/**
	1. Pull all of the Brave Search Results that matched the CVE Topic
	2. Create a Summary of the Brave Search Results that matched the CVE Topic (AI Agent #1)
	3. Create a 2nd Summary of the Brave Search Results that matched the CVE Topic (AI Agent #2)
	3. Select which Summary is the Best (AI Agent)
	4. Save the Summary to the Struct
	**/

	nvdOutputAll, err = OllamaCreateSummaryResults(nvdOutputAll)
	if err != nil {
		log.Fatalf("Error verifying search results with Ollama: %v\n", err)
	}

	// ---------------------------------- Output Results in Markdown ------------------------------
	for _, nvd := range nvdOutputAll {
		outputResult := ""
		outputResult += fmt.Sprintf("## %s\n", nvd.CVEID)
		outputResult += fmt.Sprintf("- **URL:** %s\n", nvd.CVEURL)
		outputResult += fmt.Sprintf("- **Severity:** %s\n", nvd.Severity)
		outputResult += fmt.Sprintf("- **Vulnerability Status:** %s\n\n", nvd.VulnStatus)
		outputResult += fmt.Sprintf("- **Attack Complexity:** %s\n\n", nvd.AttackComplexity)
		outputResult += fmt.Sprintf("### CVE Description\n%s\n\n", nvd.Description)
		outputResult += fmt.Sprintf("### AI Summary from Search Results\n%s\n\n", nvd.BraveSearchSummary)
		outputResult += "### Additional URL References\n"

		// Deduplicate References that are output
		var references []string
		for _, ref := range nvd.References {
			//outputResult += fmt.Sprintf("%s\n", ref)
			references = AppendIfNotExists(references, ref)
		}
		countBraveMatches := 0
		for _, result := range nvd.BraveSearchResults {
			if result.MatchesCVETopic {
				//fmt.Println("Brave Match Found")
				references = append(references, result.URL)
				//outputResult += fmt.Sprintf("%s\n", result.URL)
				countBraveMatches++
			}
		}
		// Output the references for the Markdown
		for _, ref := range references {
			outputResult += fmt.Sprintf("%s\n", ref)
		}

		//fmt.Print(outputResult)
		//SaveOutputFile(outputResult, nvd.CVEID+".md")

		// Increases the accuracy of the output by requiring at least 2 Brave Search Results to match the CVE Topic
		if countBraveMatches > 1 {
			fmt.Print(outputResult)
			SaveOutputFile(outputResult, "output/"+nvd.CVEID+".md")
		} else {
			fmt.Printf("[W] Not enough Brave Search Results matched the CVE Topic for %s, skipping markdown output file.\n", nvd.CVEID)
		}

	}

}
