package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
)

/**
References:
https://google.github.io/adk-docs/agents/multi-agents/#workflow-agents-as-orchestrators
https://www.kaggle.com/kaggle5daysofai/code

https://google.github.io/adk-docs/agents/workflow-agents/sequential-agents/


**/

type Configuration struct {
	APIKey string `json:"apikey"`
}

type InformationStruct struct {
	DemographicInfo         string
	URL1                    string
	URLImage1               string
	URL2                    string
	URLImage2               string
	URL3                    string
	URLImage3               string
	URL4                    string
	URLImage4               string
	CategoryURL             string
	HTMLEmail               string
	VerificationInformation string
	FrenchInformation       string
	GermanInformation       string
	FinalResponse           string
}

func (c *Configuration) CreateConfig(f string) error {
	c.APIKey = "NOTHING"
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

func CreateInformation(w http.ResponseWriter, r *http.Request, apiKey string, i *InformationStruct) error {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	r.ParseMultipartForm(10 << 20)
	// Reads the variable password ...  The contained password has to match for the upload to be successful
	// Due to injection we are going to compare the hashes of the passwords...
	apiKeySHA256 := CalcSHA256Hash(apiKey)
	apiKeyInputSHA256 := CalcSHA256Hash(r.FormValue("txtApikey"))
	if apiKeyInputSHA256 != apiKeySHA256 {
		fmt.Fprintf(w, "Failed to Validate API Key\n")
		fmt.Fprintf(w, "<a href='/'>Home\n")
		return fmt.Errorf("failed to validate API Key: %v")
	}

	// Create the prompt for the createEmail AI
	i.DemographicInfo = r.FormValue("txtDemographic")
	i.URL1 = r.FormValue("txtURL1")
	i.URLImage1 = r.FormValue("txtURLImage1")
	i.URL2 = r.FormValue("txtURL2")
	i.URLImage2 = r.FormValue("txtURLImage2")
	i.URL3 = r.FormValue("txtURL3")
	i.URLImage3 = r.FormValue("txtURLImage3")
	i.URL4 = r.FormValue("txtURL4")
	i.URLImage4 = r.FormValue("txtURLImage4")
	i.CategoryURL = r.FormValue("txtCategory")

	return nil

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

	/**
	agentInformation, err := createEmail(config)
	if err != nil {
		log.Fatalf("Unable to use the agents to create an HTML email and verify")
	}
	fmt.Printf("Created Email:\n%s\n\n", strings.ReplaceAll(agentInformation.HTMLEmail, "```", ""))
	fmt.Printf("Critique:\n%s\n\n", agentInformation.VerificationInformation)

	// Save the Created Email and the Critique
	now := time.Now()
	timestamp := now.Format("2006-01-02_15:04:05")
	SaveOutputFile(agentInformation.HTMLEmail, "output/ad_"+timestamp+".html")
	SaveOutputFile(agentInformation.VerificationInformation, "output/ad_critique_"+timestamp+".md")
	**/

	CreateDirectory("/static")
	CreateIndexHTML("/static/index.html")
	CreateDirectory("/static/output")
	//CreateIndexHTML("/static/output/index.html")

	validAPIKey := "hack"
	infoStruct := InformationStruct{}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		//http.Handle("/", http.FileServer(http.Dir("./static")))
		http.FileServer(http.Dir("./static")).ServeHTTP(w, r)
	})

	http.HandleFunc("/input.html", ProvideInformationHTML)

	http.HandleFunc("/input2.html", ProvideInformationHTML)

	http.HandleFunc("/input3.html", ProvideInformationHTML)

	http.HandleFunc("/create", func(w http.ResponseWriter, r *http.Request) {
		err := CreateInformation(w, r, validAPIKey, &infoStruct)
		if err != nil {
			log.Fatalf("Unable to create the information: %v\n", err)
		}
		if err == nil {
			infoStruct, errCreateEmail := CreateEmail(config, infoStruct)
			if errCreateEmail != nil {
				log.Fatalf("Unable to use the agents to create an HTML email and verify")
			}

			now := time.Now()
			timestamp := now.Format("2006-01-02_15:04:05")
			SaveOutputFile(infoStruct.HTMLEmail, "static/output/ad_"+timestamp+".html")
			SaveOutputFile(infoStruct.FrenchInformation, "static/output/ad_french_"+timestamp+".html")
			SaveOutputFile(infoStruct.GermanInformation, "static/output/ad_german_"+timestamp+".html")
			SaveOutputFile(infoStruct.VerificationInformation, "static/output/ad_critique_"+timestamp+".md")

			fmt.Fprint(w, headerHTML())
			fmt.Fprint(w, "<a href='/input.html'>Link to Create Another Email</a><br /><br />")
			fmt.Fprintf(w, "<a href='output/ad_%s.html' target='_blank'>View Created HTML Email</a><br /><br />", timestamp)
			fmt.Fprintf(w, "<a href='output/ad_french_%s.html' target='_blank'>View Created HTML Email (French)</a><br /><br />", timestamp)
			fmt.Fprintf(w, "<a href='output/ad_german_%s.html' target='_blank'>View Created HTML Email (German)</a><br /><br />", timestamp)
			fmt.Fprintf(w, "<a href='output/ad_critique_%s.md' target='_blank'>View Critique of Created HTML Email</a><br /><br />", timestamp)
			fmt.Fprint(w, tailHTML())

			//fmt.Printf("Created Email:\n%s\n\n", strings.ReplaceAll(infoStruct.HTMLEmail, "```", ""))
			//fmt.Printf("Critique:\n%s\n\n", infoStruct.VerificationInformation)

		}
	})

	listeningPort := "127.0.0.1:9500"
	fmt.Printf("Started the webserver with no encryption on port: %s\n", listeningPort)
	log.Fatal(http.ListenAndServe(listeningPort, nil))

}
