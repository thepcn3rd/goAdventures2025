package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"regexp"
	"time"
)

// Bundle represents the top-level STIX bundle.
type StixJSON struct {
	Type    string        `json:"type"`
	ID      string        `json:"id"`
	Objects []interface{} `json:"objects"` // Use interface{} to handle different object types
}

// AttackPattern represents an attack-pattern object in STIX.
type AttackPattern struct {
	SpecVersion       string    `json:"spec_version"`
	Type              string    `json:"type"`
	ID                string    `json:"id"`
	Created           time.Time `json:"created"`
	Modified          time.Time `json:"modified"`
	Name              string    `json:"name"`
	Description       string    `json:"description,omitempty"`
	Aliases           []string  `json:"aliases,omitempty"`
	CreatedByRef      string    `json:"created_by_ref,omitempty"`
	ObjectMarkingRefs []string  `json:"object_marking_refs,omitempty"`
}

// Indicator represents an indicator object in STIX.
type Indicator struct {
	SpecVersion       string    `json:"spec_version"`
	Type              string    `json:"type"`
	ID                string    `json:"id"`
	Created           time.Time `json:"created"`
	Modified          time.Time `json:"modified"`
	Pattern           string    `json:"pattern"`
	ValidFrom         time.Time `json:"valid_from"`
	PatternType       string    `json:"pattern_type"`
	CreatedByRef      string    `json:"created_by_ref,omitempty"`
	ObjectMarkingRefs []string  `json:"object_marking_refs,omitempty"`
}

// Malware represents a malware object in STIX.
type Malware struct {
	SpecVersion       string    `json:"spec_version"`
	Type              string    `json:"type"`
	ID                string    `json:"id"`
	Created           time.Time `json:"created"`
	Modified          time.Time `json:"modified"`
	Name              string    `json:"name"`
	MalwareTypes      []string  `json:"malware_types,omitempty"`
	IsFamily          bool      `json:"is_family,omitempty"`
	Aliases           []string  `json:"aliases,omitempty"`
	CreatedByRef      string    `json:"created_by_ref,omitempty"`
	ObjectMarkingRefs []string  `json:"object_marking_refs,omitempty"`
}

// Relationship represents a relationship object in STIX.
type Relationship struct {
	SpecVersion       string    `json:"spec_version"`
	Type              string    `json:"type"`
	ID                string    `json:"id"`
	Created           time.Time `json:"created"`
	Modified          time.Time `json:"modified"`
	RelationshipType  string    `json:"relationship_type"`
	SourceRef         string    `json:"source_ref"`
	TargetRef         string    `json:"target_ref"`
	CreatedByRef      string    `json:"created_by_ref,omitempty"`
	ObjectMarkingRefs []string  `json:"object_marking_refs,omitempty"`
}

// Report represents a report object in STIX.
type Report struct {
	SpecVersion       string    `json:"spec_version"`
	Type              string    `json:"type"`
	ID                string    `json:"id"`
	Created           time.Time `json:"created"`
	Modified          time.Time `json:"modified"`
	Name              string    `json:"name"`
	Description       string    `json:"description"`
	Published         time.Time `json:"published"`
	ObjectRefs        []string  `json:"object_refs"`
	CreatedByRef      string    `json:"created_by_ref,omitempty"`
	ObjectMarkingRefs []string  `json:"object_marking_refs,omitempty"`
}

// ThreatActor represents a threat-actor object in STIX.
type ThreatActor struct {
	SpecVersion       string    `json:"spec_version"`
	Type              string    `json:"type"`
	ID                string    `json:"id"`
	Created           time.Time `json:"created"`
	Modified          time.Time `json:"modified"`
	Name              string    `json:"name"`
	Aliases           []string  `json:"aliases,omitempty"`
	CreatedByRef      string    `json:"created_by_ref,omitempty"`
	ObjectMarkingRefs []string  `json:"object_marking_refs,omitempty"`
}

// Vulnerability represents a vulnerability object in STIX.
type Vulnerability struct {
	SpecVersion        string    `json:"spec_version"`
	Type               string    `json:"type"`
	ID                 string    `json:"id"`
	Created            time.Time `json:"created"`
	Modified           time.Time `json:"modified"`
	Name               string    `json:"name"`
	ExternalReferences []struct {
		SourceName string `json:"source_name"`
		ExternalID string `json:"external_id"`
	} `json:"external_references,omitempty"`
	CreatedByRef      string   `json:"created_by_ref,omitempty"`
	ObjectMarkingRefs []string `json:"object_marking_refs,omitempty"`
}

// MarkingDefinition represents a marking-definition object in STIX.
type MarkingDefinition struct {
	Type           string    `json:"type"`
	SpecVersion    string    `json:"spec_version"`
	ID             string    `json:"id"`
	Created        time.Time `json:"created"`
	DefinitionType string    `json:"definition_type"`
	Definition     struct {
		TLP string `json:"tlp"`
	} `json:"definition"`
	CreatedByRef      string   `json:"created_by_ref,omitempty"`
	ObjectMarkingRefs []string `json:"object_marking_refs,omitempty"`
}

// Location represents a location object in STIX.
type Location struct {
	Type               string    `json:"type"`
	SpecVersion        string    `json:"spec_version"`
	ID                 string    `json:"id"`
	Created            time.Time `json:"created"`
	Modified           time.Time `json:"modified"`
	Name               string    `json:"name"`
	Country            string    `json:"country"`
	AdministrativeArea string    `json:"administrative_area,omitempty"`
	CreatedByRef       string    `json:"created_by_ref,omitempty"`
	ObjectMarkingRefs  []string  `json:"object_marking_refs,omitempty"`
}

func (s *StixJSON) LoadFile(sPtr string) error {
	stixFile, err := os.Open(sPtr)
	if err != nil {
		return err
	}
	defer stixFile.Close()
	decoder := json.NewDecoder(stixFile)
	if err := decoder.Decode(&s); err != nil {
		return err
	}

	return nil
}

func (s *StixJSON) SaveNewFile(sPtr string) error {
	jsonData, err := json.MarshalIndent(s, "", "    ")
	if err != nil {
		return err
	}

	err = os.WriteFile(sPtr, jsonData, 0644)
	if err != nil {
		return err
	}

	return nil
}

type IndicatorStruct struct {
	EmailAddr  []string `json:"emailAddr"`
	MD5Hash    []string `json:"md5"`
	SHA1Hash   []string `json:"sha1"`
	SHA256Hash []string `json:"sha256"`
	SSDEEP     []string `json:"ssdeep"`
	FileName   []string `json:"filename"`
	URL        []string `json:"url"`
}

type Configuration struct {
	Webhook string `json:"teamsWebhook"`
}

func (c *Configuration) CreateConfig(cPtr string) error {
	c.Webhook = ""

	jsonData, err := json.MarshalIndent(c, "", "    ")
	if err != nil {
		return err
	}

	err = os.WriteFile(cPtr, jsonData, 0644)
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

func parseIndicator(pattern string) {
	emailRegex := regexp.MustCompile(`email-addr:value\s*=\s*'([^']+)'`)
	urlRegex := regexp.MustCompile(`url:value\s*=\s*'([^']+)'`)
	patternMD5Regex := regexp.MustCompile(`file:hashes\.MD5\s*=\s*'([^']+)'`)
	patternSSDeepRegex := regexp.MustCompile(`file:hashes\.SSDEEP\s*=\s*'([^']+)'`)
	patternSHA1Regex := regexp.MustCompile(`file:hashes\.\'SHA-1\'\s*=\s*'([^']+)'`)
	patternSHA256Regex := regexp.MustCompile(`file:hashes\.\'SHA-256\'\s*=\s*'([^']+)'`)
	patternNameRegex := regexp.MustCompile(`file:name\s*=\s*'([^']+)'`)

	// strings.ToLower() to convert the Upper-case letters to lower
	emailMatches := emailRegex.FindAllStringSubmatch(pattern, -1)
	for _, match := range emailMatches {
		indicators.EmailAddr = append(indicators.EmailAddr, match[1])
	}

	urlMatches := urlRegex.FindAllStringSubmatch(pattern, -1)
	for _, match := range urlMatches {
		indicators.URL = append(indicators.URL, match[1])
	}

	patternMD5Matches := patternMD5Regex.FindAllStringSubmatch(pattern, -1)
	for _, match := range patternMD5Matches {
		indicators.MD5Hash = append(indicators.MD5Hash, match[1])
	}

	patternSHA1Matches := patternSHA1Regex.FindAllStringSubmatch(pattern, -1)
	for _, match := range patternSHA1Matches {
		indicators.SHA1Hash = append(indicators.SHA1Hash, match[1])
	}

	patternSHA256Matches := patternSHA256Regex.FindAllStringSubmatch(pattern, -1)
	for _, match := range patternSHA256Matches {
		indicators.SHA256Hash = append(indicators.SHA256Hash, match[1])
	}

	patternSSDeepMatches := patternSSDeepRegex.FindAllStringSubmatch(pattern, -1)
	for _, match := range patternSSDeepMatches {
		indicators.SSDEEP = append(indicators.SSDEEP, match[1])
	}

	patternNameMatches := patternNameRegex.FindAllStringSubmatch(pattern, -1)
	for _, match := range patternNameMatches {
		indicators.FileName = append(indicators.FileName, match[1])
	}
}

var indicators IndicatorStruct

func main() {
	stixFilePtr := flag.String("f", "default.json", "Specify the file that you would like to load")
	teamsConfigPtr := flag.String("t", "teams.json", "Specify the teams webhook to call")
	flag.Parse()

	var config Configuration
	var stix StixJSON
	err := stix.LoadFile(*stixFilePtr)
	if err != nil {
		log.Fatalf("Unable to load the stix JSON file: %v\n", err)
	}

	err = stix.SaveNewFile("debug.json")
	if err != nil {
		log.Fatalf("Unable to save the debug.json file: %v\n", err)
	}

	if err := config.LoadConfig(*teamsConfigPtr); err != nil {
		fmt.Println("Could not load the teams configuration file, creating a new default teams.json")
		config.CreateConfig(*teamsConfigPtr)
		log.Fatalf("Modify the teams.json file to use a webhook for the output: %v\n", err)
	}

	for _, obj := range stix.Objects {
		switch v := obj.(type) {
		case map[string]interface{}:
			switch v["type"] {
			case "indicator":
				var indicator Indicator
				jsonData, _ := json.Marshal(v)
				json.Unmarshal(jsonData, &indicator)
				//fmt.Println("Indicator:", indicator.Pattern)
				parseIndicator(indicator.Pattern)
			}
		}
	}
	fmt.Println()
	fmt.Println(indicators.URL)

	if len(config.Webhook) > 1 {
		SendTeamsMessage(indicators.URL[0], config.Webhook)
	}

}
