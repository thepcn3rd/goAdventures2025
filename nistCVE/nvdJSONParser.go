package main

import (
	"encoding/json"
	"io"
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
	//nvdJSON := `{"resultsPerPage":1,"startIndex":0,"totalResults":1,"format":"NVD_CVE","version":"2.0","timestamp":"2023-08-02T15:54:26.923","vulnerabilities":"test"}`
	//nvdJSON := `{"resultsPerPage":1,"startIndex":0,"totalResults":1,"format":"NVD_CVE","version":"2.0","timestamp":"2023-08-02T15:54:26.923","vulnerabilities":[{"cve":{"id": "test"}}]}`
	//nvdJSON := `{"resultsPerPage":1,"startIndex":0,"totalResults":1,"format":"NVD_CVE","version":"2.0","timestamp":"2023-08-02T15:54:26.923","vulnerabilities":[{"cve":{"id":"CVE-2019-18935","sourceIdentifier":"cve@mitre.org","published":"2019-12-11T13:15:11.767","lastModified":"2023-03-15T18:15:10.143","vulnStatus":"Modified","cisaExploitAdd":"2021-11-03","cisaActionDue":"2022-05-03","cisaRequiredAction":"Apply updates per vendor instructions.","cisaVulnerabilityName":"Progress Telerik UI for ASP.NET AJAX Deserialization of Untrusted Data Vulnerability","descriptions":[{"lang":"en","value":"Progress Telerik UI for ASP.NET AJAX through 2019.3.1023 contains a .NET deserialization vulnerability in the RadAsyncUpload function. This is exploitable when the encryption keys are known due to the presence of CVE-2017-11317 or CVE-2017-11357, or other means. Exploitation can result in remote code execution. (As of 2020.1.114, a default setting prevents the exploit. In 2019.3.1023, but not earlier versions, a non-default setting can prevent exploitation.)"},{"lang":"es","value":"Progress Telerik UI para ASP.NET AJAX hasta 2019.3.1023 contiene una vulnerabilidad de deserialización de .NET en la función RadAsyncUpload. Esto es explotable cuando las claves de cifrado se conocen debido a la presencia de CVE-2017-11317 o CVE-2017-11357, u otros medios. La explotación puede resultar en la ejecución remota de código. (A partir de 2020.1.114, una configuración predeterminada evita la explotación. En 2019.3.1023, pero no en versiones anteriores, una configuración no predeterminada puede evitar la explotación)."}],"metrics":{"cvssMetricV31":[{"source":"nvd@nist.gov","type":"Primary","cvssData":{"version":"3.1","vectorString":"CVSS:3.1\/AV:N\/AC:L\/PR:N\/UI:N\/S:U\/C:H\/I:H\/A:H","attackVector":"NETWORK","attackComplexity":"LOW","privilegesRequired":"NONE","userInteraction":"NONE","scope":"UNCHANGED","confidentialityImpact":"HIGH","integrityImpact":"HIGH","availabilityImpact":"HIGH","baseScore":9.8,"baseSeverity":"CRITICAL"},"exploitabilityScore":3.9,"impactScore":5.9}],"cvssMetricV2":[{"source":"nvd@nist.gov","type":"Primary","cvssData":{"version":"2.0","vectorString":"AV:N\/AC:L\/Au:N\/C:P\/I:P\/A:P","accessVector":"NETWORK","accessComplexity":"LOW","authentication":"NONE","confidentialityImpact":"PARTIAL","integrityImpact":"PARTIAL","availabilityImpact":"PARTIAL","baseScore":7.5},"baseSeverity":"HIGH","exploitabilityScore":10.0,"impactScore":6.4,"acInsufInfo":false,"obtainAllPrivilege":false,"obtainUserPrivilege":false,"obtainOtherPrivilege":false,"userInteractionRequired":false}]},"weaknesses":[{"source":"nvd@nist.gov","type":"Primary","description":[{"lang":"en","value":"CWE-502"}]}],"configurations":[{"nodes":[{"operator":"OR","negate":false,"cpeMatch":[{"vulnerable":true,"criteria":"cpe:2.3:a:telerik:ui_for_asp.net_ajax:*:*:*:*:*:*:*:*","versionStartIncluding":"2011.1.315","versionEndExcluding":"2019.3.1023","matchCriteriaId":"EA6EF7DC-FEA2-4141-954C-C4A32CA4BEBD"}]}]}],"references":[{"url":"http:\/\/packetstormsecurity.com\/files\/155720\/Telerik-UI-Remote-Code-Execution.html","source":"cve@mitre.org","tags":["Third Party Advisory"]},{"url":"http:\/\/packetstormsecurity.com\/files\/159653\/Telerik-UI-ASP.NET-AJAX-RadAsyncUpload-Deserialization.html","source":"cve@mitre.org"},{"url":"https:\/\/codewhitesec.blogspot.com\/2019\/02\/telerik-revisited.html","source":"cve@mitre.org","tags":["Not Applicable"]},{"url":"https:\/\/github.com\/bao7uo\/RAU_crypto","source":"cve@mitre.org","tags":["Exploit","Third Party Advisory"]},{"url":"https:\/\/github.com\/noperator\/CVE-2019-18935","source":"cve@mitre.org","tags":["Third Party Advisory"]},{"url":"https:\/\/know.bishopfox.com\/research\/cve-2019-18935-remote-code-execution-in-telerik-ui","source":"cve@mitre.org","tags":["Exploit","Third Party Advisory"]},{"url":"https:\/\/www.bleepingcomputer.com\/news\/security\/us-federal-agency-hacked-using-old-telerik-bug-to-steal-data\/","source":"cve@mitre.org"},{"url":"https:\/\/www.telerik.com\/support\/kb\/aspnet-ajax\/details\/allows-javascriptserializer-deserialization","source":"cve@mitre.org","tags":["Patch","Vendor Advisory"]},{"url":"https:\/\/www.telerik.com\/support\/whats-new\/aspnet-ajax\/release-history\/ui-for-asp-net-ajax-r1-2020-(version-2020-1-114)","source":"cve@mitre.org"},{"url":"https:\/\/www.telerik.com\/support\/whats-new\/release-history","source":"cve@mitre.org","tags":["Release Notes","Vendor Advisory"]}]}}]}`
	// Decode the JSON into the structs that are built...
	var nvdAPI nvdAPIStruct

	//err := json.NewDecoder(httpResponse.Body).Decode(&nvdAPI)
	//json.Unmarshal([]byte(nvdJSON), &nvdAPI)
	//json.Unmarshal([]byte(jsonTXT), &nvdAPI)
	json.NewDecoder(httpResponseBody).Decode(&nvdAPI)
	//checkError("Decoding JSON Failed", err)
	//fmt.Println(nvdAPI)
	//fmt.Println(nvdAPI.Vulns[0])
	//fmt.Println(nvdAPI.Vulns[0].CVE)
	//fmt.Println(nvdAPI.Vulns[0].CVE.ID)
	//fmt.Println(nvdAPI.Vulns[0].CVE.References)
	//fmt.Println(nvdAPI.Vulns[0].CVE.References[3].Tags[0])
	return &nvdAPI
}
