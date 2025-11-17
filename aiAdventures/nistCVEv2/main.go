package main

import (
	"bufio"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"reflect"
	"strings"
	"time"

	"google.golang.org/adk/agent"
	"google.golang.org/adk/agent/llmagent"
	"google.golang.org/adk/agent/workflowagents/sequentialagent"
	"google.golang.org/adk/model"
	"google.golang.org/adk/runner"
	"google.golang.org/adk/session"
	"google.golang.org/genai"
)

type NVDOutputStruct struct {
	CVEID              string
	VulnStatus         string
	Description        string
	Source             string
	AttackComplexity   string
	Severity           string
	CVEURL             string
	KeywordMatched     []string
	BraveSearchResults []BraveSearchResults
}

type BraveSearchResults struct {
	Title           string
	URL             string
	Description     string
	MatchesCVETopic bool
}

type Configuration struct {
	TeamsWebhookURL string          `json:"teamsWebhookURL"`
	BraveAPIKey     string          `json:"braveAPIKey"`
	OllamaURL       string          `json:"ollamaURL"`
	OllamaWaitTime  int             `json:"ollamaWaitTime"`
	DebugFile       string          `json:"debugFile"`
	Keywords        []keywordStruct `json:"keywords"`
}

type keywordStruct struct {
	Value       string `json:"value"`
	Description string `json:"description"`
}

func (c *Configuration) CreateConfig(f string) error {
	c.TeamsWebhookURL = ""
	c.BraveAPIKey = ""
	c.OllamaURL = "http://localhost:11434/api/chat"
	c.OllamaWaitTime = 10 // HTTP Waittime for a response from ollama in minutes
	c.DebugFile = "debug.log"
	c.Keywords = []keywordStruct{
		{Value: "Cisco", Description: "Cisco"},
		{Value: "Linux", Description: "Linux"},
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

func IsStructValid(s interface{}) bool {
	if s == nil {
		return false
	}

	val := reflect.ValueOf(s)
	if val.Kind() == reflect.Ptr {
		if val.IsNil() {
			return false
		}
		val = val.Elem()
	}

	// Check if it's a zero-valued struct
	return !val.IsZero()
}

func ollamaNewModel(ctx context.Context, modelName string, c Configuration) (model.LLM, error) {
	ctx = context.Background()
	return &ollamaModel{
		name:       modelName,
		ollamaURL:  c.OllamaURL,
		waitTime:   c.OllamaWaitTime,
		SequenceID: 0,
	}, nil

}

func agenticAgentCall(nvd NVDOutputStruct, c Configuration, i int) (string, error) {
	ctx := context.Background()
	// Send the results to Ollama for Verification that the results relate to the CVE
	llamaModel, err := ollamaNewModel(ctx, "llama3.2:3b", c)
	if err != nil {
		log.Printf("Ollama model creation failed: %v", err)
		os.Exit(0)
	}
	/**
	instructionString := "You are a research assistant and verifying that the information about the topic is similar to a search that was returned.  After evaluating the topic and the search that was returned, provide a response of 'Yes it matches' or 'No'.  "
	instructionString += "Verify the CVE number that exists in the topic is the same as in the search results.\n\n"
	instructionString += "Topic:\n"
	instructionString += nvdOutputAll[0].CVEID + "\n"
	instructionString += nvdOutputAll[0].Description + "\n\n"
	instructionString += "Search Result:\n"
	instructionString += nvdOutputAll[0].BraveSearchResults[0].Title + "\n"
	instructionString += nvdOutputAll[0].BraveSearchResults[0].Description + "\n"

	instructionString := "You are a research assistant and verifying that the information about the topic is similar to a search that was returned. "
	instructionString += "After evaluating the topic and the search that was returned, provide a response of 'Yes' or 'No'. "
	instructionString += "Verify the CVE number that exists in the topic is the same as in the search results.\n\n"
	instructionString += "Topic:\n"
	instructionString += "CVE-2025-13232" + "\n"
	instructionString += "This vulnerability is scary to experience in your network" + "\n\n"
	instructionString += "Search Result:\n"
	instructionString += nvdOutputAll[0].BraveSearchResults[0].Title + "\n"
	instructionString += nvdOutputAll[0].BraveSearchResults[0].Description + "\n"
	**/
	instructionString := "You are a research assistant and verifying that the information about the topic is similar to a search that was returned. "
	instructionString += "After evaluating the topic and the search that was returned, provide a response of 'Yes!' or 'No!'. "
	instructionString += "Verify the CVE number that exists in the topic is the same as in the search results.\n\n"
	instructionString += "Topic:\n"
	instructionString += nvd.CVEID + "\n"
	instructionString += nvd.Description + "\n\n"
	instructionString += "Search Result:\n"
	instructionString += nvd.BraveSearchResults[i].Title + "\n"
	instructionString += nvd.BraveSearchResults[i].Description + "\n"

	fmt.Printf("\n%s\n", instructionString)

	// Verification Agent
	verificationAgent, err := llmagent.New(llmagent.Config{
		Name:        "VerificationAgent",
		Model:       llamaModel,
		Instruction: instructionString,
		OutputKey:   "outputResult",
	})
	if err != nil {
		log.Fatalf("Failed to create the outline agent: %v", err)
	}

	// Sequential Agent Example
	rootAgent, err := sequentialagent.New(sequentialagent.Config{
		AgentConfig: agent.Config{
			Name:        "",
			Description: "Executes a sequence of agents to generate a blog post based on a given topic. ",
			SubAgents:   []agent.Agent{verificationAgent},
		},
	})
	if err != nil {
		log.Fatalf("Failed to create the writer agent: %v", err)
	}

	userTopic := nvd.CVEID

	// Modified to use session and a runner instead of using the command line launcher
	sessionService := session.InMemoryService()
	initialState := map[string]any{
		"topic": userTopic,
	}

	sessionInstance, err := sessionService.Create(ctx, &session.CreateRequest{
		AppName: "VerificationAgent",
		UserID:  "thepcn3rd",
		State:   initialState,
	})
	if err != nil {
		log.Fatalf("Failed to create session: %v", err)
	}

	r, err := runner.New(runner.Config{
		AppName:        "VerificationAgent",
		Agent:          rootAgent,
		SessionService: sessionService,
	})
	if err != nil {
		log.Fatalf("Failed to create runner: %v", err)
	}

	input := genai.NewContentFromText("Verify information about this CVE is accurate "+userTopic, genai.RoleUser)
	events := r.Run(ctx, "thepcn3rd", sessionInstance.Session.ID(), input, agent.RunConfig{
		StreamingMode: agent.StreamingModeSSE,
	})

	var finalResponse string
	previousAgentAuthor := ""
	eventSectionResponse := ""
	for event, err := range events {
		if err != nil {
			log.Fatalf("An error occurred during agent execution: %v", err)
		}

		// Print each event as it arrives.
		if event.Author != previousAgentAuthor {
			//fmt.Printf("\nEvent Section Response:\n%s\n", eventSectionResponse)
			//finalResponse += "\n\n--\n\n"
			eventSectionResponse = ""
			fmt.Println("\n----- Agent Response -----")
			fmt.Printf("Agent Name: %s\n", event.Author)
			//fmt.Printf("Event Author: %s\n", event.Author)
			//fmt.Printf("Event ID: %s\n", event.ID)
			//fmt.Printf("Event Branch: %s\n", event.Branch)
			//fmt.Printf("Event Invocation ID: %s\n", event.InvocationID)
			fmt.Printf("Event Timestamp: %s\n", event.Timestamp)
			//fmt.Printf("Event Content Role: %s\n", event.Content.Role)
			previousAgentAuthor = event.Author
		}

		for _, part := range event.Content.Parts {
			finalResponse += part.Text
			eventSectionResponse += part.Text
		}

	}

	//fmt.Println("\n--- Agent Interaction Result ---")
	fmt.Printf("\nAgent Final Response:\n%s\n\n", finalResponse)

	//finalSession, err := sessionService.Get(ctx, &session.GetRequest{
	_, err = sessionService.Get(ctx, &session.GetRequest{
		UserID:    "thepcn3rd",
		AppName:   "VerificationAgent",
		SessionID: sessionInstance.Session.ID(),
	})
	if err != nil {
		log.Fatalf("Failed to retrieve final session: %v", err)
	}

	//fmt.Println("Final Session State:", finalSession.Session.State())

	// Send the results to Ollama for Summarization
	// Verify the summary relates to the CVE and the description provided by NVD
	return finalResponse, nil
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

	// Pull NVD CVE Data for the specified timeframe
	var NVDConfig NVDSearchConfiguration
	NVDConfig.NVDURL = "https://services.nvd.nist.gov/rest/json/cves/2.0/"
	NVDConfig.Timeframe = 7 // Look back 1 days for new CVEs

	nvdJSON, err := NVDConfig.SubmitRequest()
	if err != nil {
		log.Fatalf("Error submitting NVD request: %v\n", err)
	}

	fmt.Println(nvdJSON.TotalResults, "Total CVE(s) found in the specified timeframe.")

	// Loop through the ressults and check for the keywords
	var keywords []string
	for _, keyword := range config.Keywords {
		keywords = append(keywords, keyword.Value)
	}

	// Create a for loop to go through the CVEs available
	var nvdOutputAll []NVDOutputStruct
	//nvdJSON := NVDConfig.NVDResponse
	log.Println("Found the following CVEs")
	for i := 0; i < nvdJSON.TotalResults; i++ {
		var nvdOutput NVDOutputStruct
		// The following line is meant for debugging the following values being pulled from the JSON
		//fmt.Println(nvdJSON.Vulns[0].CVE.Metrics.CvssMetricV31[0].CvssData.BaseSeverity)
		// The CISA Vulnerability Name may not exist within 24 hours of publication...
		//vulnName := nvdJSON.Vulns[i].CVE.CisaVulnerabilityName
		nvdOutput.CVEID = nvdJSON.Vulns[i].CVE.ID
		log.Println(nvdOutput.CVEID)
		// vulnStatus can be the following "Analyzed" or "Awaiting Analysis" or "Undergoing Analysis"
		nvdOutput.VulnStatus = nvdJSON.Vulns[i].CVE.VulnStatus
		// Assumes the 1st description is English and displays the value
		nvdOutput.Description = nvdJSON.Vulns[i].CVE.Descriptions[0].Value
		log.Println(nvdOutput.Description)
		nvdOutput.Source = nvdJSON.Vulns[i].CVE.SourceIdentifier
		// Sometimes the CVSSMetrics V31 is nil and will abort at this location, verify that 1 exists in the array
		if len(nvdJSON.Vulns[i].CVE.Metrics.CvssMetricV31) >= 1 {
			nvdOutput.AttackComplexity = nvdJSON.Vulns[i].CVE.Metrics.CvssMetricV31[0].CvssData.AttackComplexity
			nvdOutput.Severity = nvdJSON.Vulns[i].CVE.Metrics.CvssMetricV31[0].CvssData.BaseSeverity
		} else if len(nvdJSON.Vulns[i].CVE.Metrics.CvssMetricV40) >= 1 {
			nvdOutput.AttackComplexity = nvdJSON.Vulns[i].CVE.Metrics.CvssMetricV40[0].CvssData.AttackComplexity
			nvdOutput.Severity = nvdJSON.Vulns[i].CVE.Metrics.CvssMetricV40[0].CvssData.BaseSeverity
		} else {
			nvdOutput.AttackComplexity = "Not Available"
			nvdOutput.Severity = "Not Available"
		}
		nvdOutput.CVEURL = "https://nvd.nist.gov/vuln/detail/" + nvdOutput.CVEID
		// Before posting the CVE identify if the CVE applies to keywords specified
		// If a keyword matches the description more than once it will print the message more than once
		var duplicate bool
		var keywordExists bool
		duplicate = false
		keywordExists = false
		for _, k := range keywords {
			if strings.Contains(strings.ToLower(nvdOutput.Description), strings.ToLower(k)) && !duplicate {
				nvdOutput.KeywordMatched = append(nvdOutput.KeywordMatched, strings.ToLower(k))
				duplicate = true
				keywordExists = true
			}
			duplicate = false
		}
		if keywordExists {
			nvdOutputAll = append(nvdOutputAll, nvdOutput)
		}
	}
	// Save matches to a new struct for the CVE matches
	fmt.Println(len(nvdOutputAll), "CVE(s) matched the specified keywords.")

	braveConfig := BraveConfiguration{
		BraveURL: "https://api.search.brave.com/res/v1/web/search",
		//SearchKeyword: "\"" + nvdEntry.CVEID + "\"",
		SearchKeyword:   "", // Adds the search keyword in the first line of the for loop...
		BraveAPIKey:     config.BraveAPIKey,
		ResultCount:     20,   // Could set this back to 10
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

	for i, nvdEntry := range nvdOutputAll {
		braveConfig.SearchKeyword = nvdEntry.CVEID
		fmt.Printf("CVE ID: %s\nSource: %s\nSeverity: %s\nStatus: %s\nAttack Complexity: %s\nURL: %s\nKeywords Matched: %v\n\nDescription: %s\n\n---\n\n",
			nvdEntry.CVEID,
			nvdEntry.Source,
			nvdEntry.Severity,
			nvdEntry.VulnStatus,
			nvdEntry.AttackComplexity,
			nvdEntry.CVEURL,
			nvdEntry.KeywordMatched,
			nvdEntry.Description)

		// Pull the Brave Search Results for the CVE ID
		// If you do not quote the CVE ID it will return results that may not be relevant
		fmt.Printf("Searching Brave for this CVE: %s\n", nvdEntry.CVEID)

		// Slowing down the Brave Searches due to the API restrictions for the free tier
		time.Sleep(2 * time.Second)
		braveResults, err := braveConfig.SubmitRequest()
		if err != nil {
			log.Fatalf("Error submitting Brave request: %v\n", err)
		}

		if IsStructValid(braveResults.Web) {
			if len(braveResults.Web.Results) >= 1 {
				fmt.Printf("Results exist...\n")
				for _, braveEntry := range braveResults.Web.Results {
					fmt.Printf("Brave Results\nTitle: %s\nURL: %s\nDescription:\n%s\n\n", braveEntry.Title, braveEntry.URL, braveEntry.Description)
					var braveSearchResult BraveSearchResults
					braveSearchResult.Title = braveEntry.Title
					braveSearchResult.URL = braveEntry.URL
					braveSearchResult.Description = braveEntry.Description
					braveSearchResult.MatchesCVETopic = false
					nvdOutputAll[i].BraveSearchResults = append(nvdOutputAll[i].BraveSearchResults, braveSearchResult)
				}
			} else {
				fmt.Printf("No Brave Search results found for this CVE.\n\n")
				nvdOutputAll[i].BraveSearchResults = append(nvdOutputAll[i].BraveSearchResults, BraveSearchResults{Title: "", Description: "No results found", URL: "None"})
			}
		} else {
			fmt.Printf("No Brave Search results found for this CVE.\n\n")
			nvdOutputAll[i].BraveSearchResults = append(nvdOutputAll[i].BraveSearchResults, BraveSearchResults{Title: "", Description: "No results found", URL: "None"})
		}
		//fmt.Printf("Brave Search returned %d results for %s\n\n", len(braveResults.Web), nvdEntry.CVEID)
		//nvdOutputAll = append(nvdOutputAll, nvdEntry)
	}

	// Ask an agent if the results of a topic and a search result match
	for nvdCount, nvd := range nvdOutputAll {
		for iteration := range nvd.BraveSearchResults {
			finalResponse, err := agenticAgentCall(nvd, config, iteration)
			if err != nil {
				log.Printf("unable to determine if the topic matches the search results%v\n", err)
			}

			fmt.Printf("---\nIteration: %d\n", iteration)
			fmt.Println(finalResponse)

			if strings.Contains(strings.ToLower(finalResponse), "Yes!") {
				fmt.Println("Search Result Matches CVE Topic")
				nvdOutputAll[nvdCount].BraveSearchResults[iteration].MatchesCVETopic = true
			}
		}
	}

}
