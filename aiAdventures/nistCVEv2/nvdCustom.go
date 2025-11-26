package main

import (
	"fmt"
	"log"
	"nistCVEv2/nvd"
)

type NVDOutputStruct struct {
	CVEID            string
	VulnStatus       string
	Description      string
	Source           string
	AttackComplexity string
	Severity         string
	CVEURL           string
	//KeywordMatched     []string
	BraveSearchResults []BraveSearchResults
	BraveSearchSummary string
	BraveSearchMatches int
	References         []string // URLs provided as references in the CVE struct
}

func CreateBasicNVDStruct(nvdJSON nvd.NVDAPIStruct) []NVDOutputStruct {
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
		fmt.Printf("[C] Processing CVE: %s\n", nvdOutput.CVEID)
		//log.Println(nvdOutput.CVEID)
		// vulnStatus can be the following "Analyzed" or "Awaiting Analysis" or "Undergoing Analysis" or "Received"
		nvdOutput.VulnStatus = nvdJSON.Vulns[i].CVE.VulnStatus
		// Assumes the 1st description is English and displays the value
		nvdOutput.Description = nvdJSON.Vulns[i].CVE.Descriptions[0].Value
		//log.Println(nvdOutput.Description)
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
		for _, ref := range nvdJSON.Vulns[i].CVE.References {
			nvdOutput.References = append(nvdOutput.References, ref.URL)
		}
		nvdOutputAll = append(nvdOutputAll, nvdOutput)
	}

	return nvdOutputAll
}

func FilterBasicNVDStruct(nvdOutput []NVDOutputStruct) []NVDOutputStruct {
	var nvdFilteredOutput []NVDOutputStruct
	//nvdJSON := NVDConfig.NVDResponse
	log.Println("Filtering the CVEs")
	log.Println("- Filtering on Vulnerability Status of Analyzed")
	log.Println("- Filtering out Severity of MEDIUM, Low, and Not Available")
	for _, nvd := range nvdOutput {
		if nvd.VulnStatus == "Analyzed" && nvd.Severity != "LOW" && nvd.Severity != "Not Available" && nvd.Severity != "MEDIUM" {
			nvdFilteredOutput = append(nvdFilteredOutput, nvd)
		}
	}

	return nvdFilteredOutput
}
