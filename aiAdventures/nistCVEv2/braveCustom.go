package main

import (
	"fmt"
	"nistCVEv2/brave"
	"reflect"
	"time"
)

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

func BraveGetSearchResults(nvdOutputAll []NVDOutputStruct, braveConfig brave.BraveConfiguration) ([]NVDOutputStruct, error) {

	for i, nvdEntry := range nvdOutputAll {
		//braveConfig.SearchKeyword = nvdEntry.CVEID
		braveConfig.SearchKeyword = "+" + nvdEntry.CVEID
		fmt.Printf("[C] CVE ID: %s\nSource: %s\nSeverity: %s\nStatus: %s\nAttack Complexity: %s\nURL: %s\n\nDescription: %s\n\n---\n\n",
			nvdEntry.CVEID,
			nvdEntry.Source,
			nvdEntry.Severity,
			nvdEntry.VulnStatus,
			nvdEntry.AttackComplexity,
			nvdEntry.CVEURL,
			nvdEntry.Description)

		// Pull the Brave Search Results for the CVE ID
		// If you do not quote the CVE ID it will return results that may not be relevant
		fmt.Printf("[B] Searching Brave for this CVE: %s\n", nvdEntry.CVEID)

		// Slowing down the Brave Searches due to the API restrictions for the free tier
		time.Sleep(2 * time.Second)
		braveResults, err := braveConfig.SubmitRequest()
		if err != nil {
			//log.Fatalf("Error submitting Brave request: %v\n", err)
			return nvdOutputAll, fmt.Errorf("error submitting Brave request: %v", err)
		}

		if IsStructValid(braveResults.Web) {
			if len(braveResults.Web.Results) >= 1 {
				//fmt.Printf("Results exist...\n")
				for _, braveEntry := range braveResults.Web.Results {
					fmt.Printf("[B] CVE: %s\n", nvdEntry.CVEID)
					fmt.Printf("[B] Brave Results\nTitle: %s\nURL: %s\nDescription:\n%s\n\n", braveEntry.Title, braveEntry.URL, braveEntry.Description)
					braveSearchResult := BraveSearchResults{
						Title:           braveEntry.Title,
						URL:             braveEntry.URL,
						Description:     braveEntry.Description,
						MatchesCVETopic: false, // This is unknown until it goes through AI
					}
					nvdOutputAll[i].BraveSearchResults = append(nvdOutputAll[i].BraveSearchResults, braveSearchResult)
				}
			} else {
				fmt.Printf("[B] No Brave Search results found for this CVE.\n\n")
				nvdOutputAll[i].BraveSearchResults = append(nvdOutputAll[i].BraveSearchResults, BraveSearchResults{Title: "", Description: "No results found", URL: "None", MatchesCVETopic: false})
			}
		} else {
			fmt.Printf("[B] No Brave Search results found for this CVE.\n\n")
			nvdOutputAll[i].BraveSearchResults = append(nvdOutputAll[i].BraveSearchResults, BraveSearchResults{Title: "", Description: "No results found", URL: "None", MatchesCVETopic: false})
		}
		//fmt.Printf("Brave Search returned %d results for %s\n\n", len(braveResults.Web), nvdEntry.CVEID)
		//nvdOutputAll = append(nvdOutputAll, nvdEntry)
	}

	return nvdOutputAll, nil

}
