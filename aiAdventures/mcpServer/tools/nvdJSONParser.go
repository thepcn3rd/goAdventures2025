package tools

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// In a struct the first letter of the variable has to be capitalized
type nvdAPIStruct struct {
	ResultsPerPage int           `json:"resultsPerPage"`
	StartIndex     int           `json:"startIndex"`
	TotalResults   int           `json:"totalResults"`
	Format         string        `json:"format"`
	Version        string        `json:"version"`
	Timestamp      string        `json:"timestamp"`
	Vulns          []vulnsStruct `json:"vulnerabilities"`
}

type vulnsStruct struct {
	CVE cveStruct `json:"cve"`
}

type cveStruct struct {
	ID                    string         `json:"id"`
	SourceIdentifier      string         `json:"sourceIdentifier"`
	Published             string         `json:"published"`
	LastModified          string         `json:"lastModified"`
	VulnStatus            string         `json:"vulnStatus"`
	CisaExploitAdd        string         `json:"cisaExploitAdd"`
	CisaActionDue         string         `json:"cisaActionDue"`
	CisaRequiredAction    string         `json:"cisaRequiredAction"`
	CisaVulnerabilityName string         `json:"cisaVulnerabilityName"`
	Descriptions          []descStruct   `json:"descriptions"`
	Metrics               metricsStruct  `json:"metrics"`
	Weaknesses            []weakStruct   `json:"weaknesses"`
	Configurations        []configStruct `json:"configurations"`
	References            []refStruct    `json:"references"`
}

type descStruct struct {
	Lang  string `json:"lang"`
	Value string `json:"value"`
}

type metricsStruct struct {
	CvssMetricV31 []cvss31Struct `json:"cvssMetricV31"`
	CvssMetricV2  []cvss2Struct  `json:"cvssMetricV2"`
}

type cvss31Struct struct {
	Source              string           `json:"source"`
	Type                string           `json:"type"`
	CvssData            cvssData31Struct `json:"cvssData"`
	ExploitabilityScore string           `json:"exploitabilityScore"`
	ImpactScore         string           `json:"impactScore"`
}

type cvssData31Struct struct {
	Version               string `json:"version"`
	VectorString          string `json:"vectorString"`
	AttackVector          string `json:"attackVector"`
	AttackComplexity      string `json:"attackComplexity"`
	PrivilegesRequired    string `json:"privilegesRequired"`
	UserInteraction       string `json:"userInteraction"`
	Scope                 string `json:"scope"`
	ConfidentialityImpact string `json:"confidentialityImpact"`
	IntegrityImpact       string `json:"integrityImpact"`
	AvailabilityImpact    string `json:"availabilityImpact"`
	BaseScore             string `json:"baseScore"`
	BaseSeverity          string `json:"baseSeverity"`
}

type cvss2Struct struct {
	Source                  string          `json:"source"`
	Type                    string          `json:"type"`
	CvssData                cvssData2Struct `json:"cvssData"`
	BaseSeverity            string          `json:"baseSeverity"`
	ExploitabilityScore     string          `json:"exploitabilityScore"`
	ImpactScore             string          `json:"impactScore"`
	AcInsufInfo             string          `json:"acInsufInfo"`
	ObtainAllPrivilege      string          `json:"obtainAllPrivilege"`
	ObtainUserPrivilege     string          `json:"obtainUserPrivilege"`
	ObtainOtherPrivilege    string          `json:"obtainOtherPrivilege"`
	UserInteractionRequired string          `json:"userInteractionRequired"`
}

type cvssData2Struct struct {
	Version               string `json:"version"`
	VectorString          string `json:"vectorString"`
	AccessVector          string `json:"accessVector"`
	AccessComplexity      string `json:"accessComplexity"`
	Authentication        string `json:"authentication"`
	ConfidentialityImpact string `json:"confidentialityImpact"`
	IntegrityImpact       string `json:"integrityImpact"`
	AvailabilityImpact    string `json:"availabilityImpact"`
	BaseScore             string `json:"baseScore"`
}

type weakStruct struct {
	Source      string           `json:"source"`
	Type        string           `json:"type"`
	Description []weakDescStruct `json:"description"`
}

type weakDescStruct struct {
	Lang  string `json:"lang"`
	Value string `json:"value"`
}

type configStruct struct {
	Nodes []configNodesStruct `json:"nodes"`
}

type configNodesStruct struct {
	Operator string                 `json:"operator"`
	Negate   string                 `json:"negate"`
	CPEMatch []configNodesCPEStruct `json:"cpeMatch"`
}

type configNodesCPEStruct struct {
	Vulnerable            string `json:"vulnerable"`
	Criteria              string `json:"criteria"`
	VersionStartIncluding string `json:"versionStartIncluding"`
	VersionEndExcluding   string `json:"versionEndExcluding"`
	MatchCriteriaId       string `json:"matchCriteriaId"`
}

type refStruct struct {
	URL    string   `json:"url"`
	Source string   `json:"source"`
	Tags   []string `json:"tags"`
}

// Function needs to be captitalized for it to be called correctly...
func NVDParser(httpResponseBody io.Reader) *nvdAPIStruct {
	var nvdAPI nvdAPIStruct

	json.NewDecoder(httpResponseBody).Decode(&nvdAPI)
	return &nvdAPI
}

type CVESearchParams struct {
	Keyword   string `json:"keyword" jsonschema:"Keyword provided to retrieve CVE information (i.e. wordpress)"`
	Timeframe string `json:"timeframe" jsonschema:"Timeframe to search for CVEs (1 Day, 7 Days, 30 Days)"`
}

func CVESearch(ctx context.Context, req *mcp.CallToolRequest, params *CVESearchParams) (*mcp.CallToolResult, any, error) {
	var httpClient http.Client

	// Calculate the date for the API query
	endDate := time.Now()
	lookBackDays, err := strconv.Atoi(params.Timeframe)
	if err != nil {
		lookBackDays = 7 // Default to 7 days if no timeframe provided
	}
	lookBackDays = lookBackDays * -1
	startDate := endDate.AddDate(0, 0, lookBackDays)
	nvdBaseURL := "https://services.nvd.nist.gov/rest/json/cves/2.0/?pubStartDate="
	nvdBaseURL += startDate.Format("2006-01-02")
	nvdBaseURL += "T00:00:00.000&pubEndDate="
	nvdBaseURL += endDate.Format("2006-01-02")
	nvdBaseURL += "T00:00:00.000"
	nvdBaseURL += "&keywordSearch="
	nvdBaseURL += params.Keyword
	log.Printf("NVD URL Queried: %s", nvdBaseURL)

	// Build the request for the Base URL above...
	httpRequest, err := http.NewRequest("GET", nvdBaseURL, nil)
	if err != nil {
		log.Printf("Unable to build http request for the NIST NVD API\n%v", err)
		return nil, "", err
	}

	// Receive response through the httpClient connection
	httpResponse, err := httpClient.Do(httpRequest)
	if err != nil {
		log.Printf("Unable to pull http response from NIST NVD API\n%v", err)
		return nil, "", err
	}

	defer httpResponse.Body.Close()

	// Verify we receive a 200 response and if not exit the program...
	if httpResponse.StatusCode >= 200 && httpResponse.StatusCode <= 299 {
		log.Println("Response Status: " + httpResponse.Status)
		return nil, "", nil
	}

	// Pass the httpResponse.Body to the Parser to Place it into a struct
	// Left the below lines for debugging
	//responseBody, err := io.ReadAll(httpResponse.Body)
	//fmt.Print(string(responseBody))
	nvdJSON := NVDParser(httpResponse.Body)

	log.Printf("Total CVEs Found: %d", nvdJSON.TotalResults)

	var output string
	output = "CVEID,Source,Severity,VulnerabilityStatus,Description,AttackComplexity,Link\n"
	for i := 0; i < nvdJSON.TotalResults; i++ {
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
		outputMessage := "\"" + cveID + "\","
		outputMessage += "\"" + cveSource + "\","
		outputMessage += "\"" + cveSeverity + "\",\"" + vulnStatus + "\","
		outputMessage += "\"" + cveDescription[:25] + "\","
		outputMessage += "\"" + attackComplexity + "\","
		outputMessage += "\"https://nvd.nist.gov/vuln/detail/" + cveID + "\"\n"
		// Place link in the adaptive card to go out and view the CVE

		output += outputMessage

	}

	log.Printf("Output from NVD Search CVE Tool:\n%s", output)

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: output},
		},
	}, nil, nil
}
