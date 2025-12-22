package nvd

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

// Reference: https://nvd.nist.gov/developers/vulnerabilities

// In a struct the first letter of the variable has to be capitalized
type NVDAPIStruct struct {
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
	ID               string           `json:"id"`
	SourceIdentifier string           `json:"sourceIdentifier"`
	Published        string           `json:"published"`
	LastModified     string           `json:"lastModified"`
	VulnStatus       string           `json:"vulnStatus"`
	CveTags          []interface{}    `json:"cveTags"`
	Descriptions     []descStruct     `json:"descriptions"`
	Metrics          metricsStruct    `json:"metrics"`
	Weaknesses       []weakDescStruct `json:"weaknesses"`
	References       []refStruct      `json:"references"`
}

type descStruct struct {
	Lang  string `json:"lang,omitempty"`
	Value string `json:"value,omitempty"`
}

type metricsStruct struct {
	CvssMetricV40 []CvssMetricV40 `json:"cvssMetricV40,omitempty"`
	CvssMetricV31 []cvss31Struct  `json:"cvssMetricV31,omitempty"`
	//CvssMetricV2  []cvss2Struct   `json:"cvssMetricV2,omitempty"`
}

type CvssMetricV40 struct {
	Source   string      `json:"source"`
	Type     string      `json:"type"`
	CvssData CvssDataV40 `json:"cvssData"`
}

type CvssDataV40 struct {
	Version                           string  `json:"version"`
	VectorString                      string  `json:"vectorString"`
	BaseScore                         float64 `json:"baseScore"`
	BaseSeverity                      string  `json:"baseSeverity"`
	AttackVector                      string  `json:"attackVector"`
	AttackComplexity                  string  `json:"attackComplexity"`
	AttackRequirements                string  `json:"attackRequirements"`
	PrivilegesRequired                string  `json:"privilegesRequired"`
	UserInteraction                   string  `json:"userInteraction"`
	VulnConfidentialityImpact         string  `json:"vulnConfidentialityImpact"`
	VulnIntegrityImpact               string  `json:"vulnIntegrityImpact"`
	VulnAvailabilityImpact            string  `json:"vulnAvailabilityImpact"`
	SubConfidentialityImpact          string  `json:"subConfidentialityImpact"`
	SubIntegrityImpact                string  `json:"subIntegrityImpact"`
	SubAvailabilityImpact             string  `json:"subAvailabilityImpact"`
	ExploitMaturity                   string  `json:"exploitMaturity"`
	ConfidentialityRequirement        string  `json:"confidentialityRequirement"`
	IntegrityRequirement              string  `json:"integrityRequirement"`
	AvailabilityRequirement           string  `json:"availabilityRequirement"`
	ModifiedAttackVector              string  `json:"modifiedAttackVector"`
	ModifiedAttackComplexity          string  `json:"modifiedAttackComplexity"`
	ModifiedAttackRequirements        string  `json:"modifiedAttackRequirements"`
	ModifiedPrivilegesRequired        string  `json:"modifiedPrivilegesRequired"`
	ModifiedUserInteraction           string  `json:"modifiedUserInteraction"`
	ModifiedVulnConfidentialityImpact string  `json:"modifiedVulnConfidentialityImpact"`
	ModifiedVulnIntegrityImpact       string  `json:"modifiedVulnIntegrityImpact"`
	ModifiedVulnAvailabilityImpact    string  `json:"modifiedVulnAvailabilityImpact"`
	ModifiedSubConfidentialityImpact  string  `json:"modifiedSubConfidentialityImpact"`
	ModifiedSubIntegrityImpact        string  `json:"modifiedSubIntegrityImpact"`
	ModifiedSubAvailabilityImpact     string  `json:"modifiedSubAvailabilityImpact"`
	Safety                            string  `json:"safety"`
	Automatable                       string  `json:"automatable"`
	Recovery                          string  `json:"recovery"`
	ValueDensity                      string  `json:"valueDensity"`
	VulnerabilityResponseEffort       string  `json:"vulnerabilityResponseEffort"`
	ProviderUrgency                   string  `json:"providerUrgency"`
}

type cvss31Struct struct {
	Source              string           `json:"source,omitempty"`
	Type                string           `json:"type,omitempty"`
	CvssData            cvssData31Struct `json:"cvssData,omitempty"`
	ExploitabilityScore float64          `json:"exploitabilityScore,omitempty"`
	ImpactScore         float64          `json:"impactScore,omitempty"`
}

type cvssData31Struct struct {
	Version      string `json:"version,omitempty"`
	VectorString string `json:"vectorString,omitempty"`
	//AttackVector          string `json:"attackVector,omitempty"`
	AttackComplexity string `json:"attackComplexity,omitempty"`
	//PrivilegesRequired    string `json:"privilegesRequired,omitempty"`
	//UserInteraction       string `json:"userInteraction,omitempty"`
	//Scope                 string `json:"scope,omitempty"`
	//ConfidentialityImpact string `json:"confidentialityImpact,omitempty"`
	//IntegrityImpact       string `json:"integrityImpact,omitempty"`
	//AvailabilityImpact    string `json:"availabilityImpact,omitempty"`
	//BaseScore             string `json:"baseScore,omitempty"`
	BaseSeverity string `json:"baseSeverity,omitempty"`
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
	URL    string `json:"url"`
	Source string `json:"source"`
	//Tags   []string `json:"tags,omitempty"`
}

type NVDSearchConfiguration struct {
	NVDURL           string       `json:"nvdURL"`
	NVDFullURL       string       `json:"nvdFullURL"`
	Timeframe        int          `json:"timeframe"` // number of days to look back
	NVDResponse      NVDAPIStruct `json:"nvdResponse,omitempty"`
	PullFromXDaysAgo int          `json:"pullFromXDaysAgo,omitempty"`
}

func (n *NVDSearchConfiguration) BuildURL() error {
	if n.NVDURL == "" {
		n.NVDURL = "https://services.nvd.nist.gov/rest/json/cves/2.0/"
	}
	// Calculate the date for the API query
	endDate := time.Now()
	endDate = endDate.AddDate(0, 0, -n.PullFromXDaysAgo)
	startDate := endDate.AddDate(0, 0, -n.Timeframe)

	values := url.Values{}
	values.Add("pubStartDate", startDate.Format("2006-01-02")+"T00:00:00.000")
	values.Add("pubEndDate", endDate.Format("2006-01-02")+"T00:00:00.000")

	n.NVDFullURL = n.NVDURL + "?" + values.Encode()

	return nil
}

func (n *NVDSearchConfiguration) SubmitRequest() (NVDAPIStruct, error) {
	err := n.BuildURL()
	if err != nil {
		return NVDAPIStruct{}, err
	}

	req, err := http.NewRequest("GET", n.NVDFullURL, nil)
	if err != nil {
		return NVDAPIStruct{}, fmt.Errorf("unable to build http request for the NIST NVD API: %v", err)
	}

	client := &http.Client{}

	// Receive response through the httpClient connection
	resp, err := client.Do(req)
	if err != nil {
		return NVDAPIStruct{}, fmt.Errorf("unable to pull http response from NIST NVD API: %v", err)
	}
	defer resp.Body.Close()

	// Verify we receive a 200 response and if not exit the program...
	if resp.Status != "200 OK" {
		return NVDAPIStruct{}, fmt.Errorf("response status is non-200: %s", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return NVDAPIStruct{}, fmt.Errorf("unable to read the response body %v", err)
	}

	//log.Printf("Response Body\n%s\n\n", string(body))

	// Parse NVD response (structure would depend on their API)

	var nvdResponse NVDAPIStruct
	if err := json.Unmarshal(body, &nvdResponse); err != nil {
		return NVDAPIStruct{}, fmt.Errorf("unable to unmarshal the NVD response")
	}
	/**
	var nvdResponse NVDAPIStruct
	if err := json.NewDecoder(resp.Body).Decode(&nvdResponse); err != nil {
		return NVDAPIStruct{}, fmt.Errorf("unable to unmarshal the NVD response")
	}
	**/
	return nvdResponse, nil
}
