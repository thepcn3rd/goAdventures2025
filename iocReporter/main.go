package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"regexp"
	"strings"
)

// Inspired by SANS Holiday Hack Challenge 2025
// Build a program to extract emails, URLs, domains, email addresses - Remove trusted items
// Defang the http to hxxp, IP Addresses [.], Emails [@], URLs [://]
// Add capability for IPv6 addresses
// Trusted items that should not be included
// Read multiple files

type Configuration struct {
	ReportFilename   string   `json:"reportFilename"`   // Filename of the report to be generated
	InputFilenames   []string `json:"inputFilenames"`   // List of files to read from
	TrustedDomains   []string `json:"trustedDomains"`   // List of trusted domains to be used in the graph
	IgnoredDomains   []string `json:"ignoredDomains"`   // List of ignored domains to be used in the graph
	TrustedNetworks  []string `json:"trustedNetworks"`  // List of trusted networks to be used in the graph
	InternalNetworks []string `json:"internalNetworks"` // List of internal networks to be used in the graph
	IgnoredNetworks  []string `json:"ignoredNetworks"`  // List of networks to ignore in the graph
}

func (c *Configuration) CreateConfig(f string) error {
	c.ReportFilename = "report.txt"
	c.InputFilenames = []string{"email.txt", "info.txt"}
	c.TrustedDomains = []string{"dosisneighborhood.corp"}
	c.IgnoredDomains = []string{"google.com", "myvendor.com"}
	c.TrustedNetworks = []string{"145.6.0.0/16", "165.6.0.0/16"}
	c.InternalNetworks = []string{"10.0.0.0/8", "172.16.0.0/12", "192.168.0.0/16"}
	c.IgnoredNetworks = []string{"169.254.0.0/16"}
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

// IPExtractor struct for more organized approach
type IOCExtractor struct {
	ipv4Pattern   *regexp.Regexp
	ipv6Pattern   *regexp.Regexp
	domainPattern *regexp.Regexp
	emailPattern  *regexp.Regexp
	urlPattern    *regexp.Regexp
}

// Helper function to remove duplicate items from a string slice
func removeDuplicates(items []string) []string {
	seen := make(map[string]bool)
	result := []string{}

	for _, ip := range items {
		if !seen[ip] {
			seen[ip] = true
			result = append(result, ip)
		}
	}

	return result
}

// NewIOCExtractor creates a new IOC extractor with compiled regex patterns
func NewIOCExtractor() (*IOCExtractor, error) {
	// IPv4 pattern
	ipv4Pattern := `\b(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.` +
		`(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.` +
		`(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.` +
		`(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\b`

	// IPv6 pattern
	ipv6Pattern := `([[:xdigit:]]{1,4}(?::[[:xdigit:]]{1,4}){7}|::|:(?::[[:xdigit:]]{1,4}){1,6}|[[:xdigit:]]{1,4}:(?::[[:xdigit:]]{1,4}){1,5}|(?:[[:xdigit:]]{1,4}:){2}(?::[[:xdigit:]]{1,4}){1,4}|(?:[[:xdigit:]]{1,4}:){3}(?::[[:xdigit:]]{1,4}){1,3}|(?:[[:xdigit:]]{1,4}:){4}(?::[[:xdigit:]]{1,4}){1,2}|(?:[[:xdigit:]]{1,4}:){5}:[[:xdigit:]]{1,4}|(?:[[:xdigit:]]{1,4}:){1,6}:)`

	// Domain pattern
	domainPattern := `\b(?:[a-zA-Z0-9-]{1,63}\.)+[a-zA-Z]{2,6}\b`

	// Email pattern
	emailPattern := `[a-z0-9_\.-]+\@[\da-z\.-]+\.[a-z\.]{2,6}`

	// URL pattern
	urlPattern := `https?://[^\s/$.?#].[^\s]*`

	ipv4Re, err := regexp.Compile(ipv4Pattern)
	if err != nil {
		return nil, fmt.Errorf("failed to compile IPv4 regex: %v", err)
	}

	ipv6Re, err := regexp.Compile(ipv6Pattern)
	if err != nil {
		return nil, fmt.Errorf("failed to compile IPv6 regex: %v", err)
	}

	domainRe, err := regexp.Compile(domainPattern)
	if err != nil {
		return nil, fmt.Errorf("failed to compile domain regex: %v", err)
	}

	emailRe, err := regexp.Compile(emailPattern)
	if err != nil {
		return nil, fmt.Errorf("failed to compile email regex: %v", err)
	}

	urlRe, err := regexp.Compile(urlPattern)
	if err != nil {
		return nil, fmt.Errorf("failed to compile URL regex: %v", err)
	}

	return &IOCExtractor{
		ipv4Pattern:   ipv4Re,
		ipv6Pattern:   ipv6Re,
		domainPattern: domainRe,
		emailPattern:  emailRe,
		urlPattern:    urlRe,
	}, nil
}

// ExtractAll extracts both IPv4 and IPv6 addresses
func (e *IOCExtractor) ExtractIPs(content string) ([]string, []string) {
	ipv4Addresses := e.ipv4Pattern.FindAllString(content, -1)
	ipv6Addresses := e.ipv6Pattern.FindAllString(content, -1)
	return removeDuplicates(ipv4Addresses), removeDuplicates(ipv6Addresses)
}

func (e *IOCExtractor) ExtractDomains(content string) []string {
	domains := e.domainPattern.FindAllString(content, -1)
	return removeDuplicates(domains)
}

func (e *IOCExtractor) ExtractEmails(content string) []string {
	emails := e.emailPattern.FindAllString(content, -1)
	return removeDuplicates(emails)
}

func (e *IOCExtractor) ExtractURLs(content string) []string {
	urls := e.urlPattern.FindAllString(content, -1)
	return removeDuplicates(urls)
}

// ReadFileToString reads a file and returns its content as a string
func ReadFileToString(filename string) (string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return "", err
	}
	defer file.Close()

	var content string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		content += scanner.Text() + "\n"
	}

	if err := scanner.Err(); err != nil {
		return "", err
	}

	return content, nil
}

type IOCInformationStruct struct {
	IPv4Addresses  []string
	IPv6Addresses  []string
	Domains        []string
	URLs           []string
	EmailAddresses []string
}

// ProcessFile processes a file and extracts IP addresses
func (i *IOCInformationStruct) ProcessFile(filename string) error {
	// Read file content
	content, err := ReadFileToString(filename)
	if err != nil {
		return fmt.Errorf("error reading file: %v", err)
	}

	extractor, err := NewIOCExtractor()
	if err != nil {
		return fmt.Errorf("error creating IOC extractor: %v", err)
	}

	// Extract IP Addresses
	i.IPv4Addresses, i.IPv6Addresses = extractor.ExtractIPs(content)
	// Extract Domains
	i.Domains = extractor.ExtractDomains(content)
	// Extract URLs
	i.URLs = extractor.ExtractURLs(content)
	// Extract Emails
	i.EmailAddresses = extractor.ExtractEmails(content)

	return nil
}

func (i *IOCInformationStruct) CreateReport(filename string) error {
	// Create a new file for the report
	reportFile, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("error creating report file: %v", err)
	}
	defer reportFile.Close()

	reportFile.WriteString("IOC Report\n")
	reportFile.WriteString("* Verify the IOCs due to false positives\n")
	reportFile.WriteString("--------------------------------------------------------\n")

	// Write IPv4 addresses to report file
	for _, ip := range i.IPv4Addresses {
		reportFile.WriteString("IPv4: " + ip + "\n")
	}

	// Write IPv6 addresses to report file
	for _, ip := range i.IPv6Addresses {
		reportFile.WriteString("IPv6: " + ip + "\n")
	}

	// Write Domains to report file
	for _, domain := range i.Domains {
		reportFile.WriteString("Domain: " + domain + "\n")
	}

	// Write URLs to report file
	for _, url := range i.URLs {
		reportFile.WriteString("URL: " + url + "\n")
	}

	// Write Emails to report file
	for _, email := range i.EmailAddresses {
		reportFile.WriteString("Email: " + email + "\n")
	}

	reportFile.Close()

	return nil
}

func (i *IOCInformationStruct) CreateReportDefanged(filename string) error {
	// Create a new file for the report
	reportFile, err := os.Create("defanged_" + filename)
	if err != nil {
		return fmt.Errorf("error creating report file: %v", err)
	}
	defer reportFile.Close()

	reportFile.WriteString("IOC Report\n")
	reportFile.WriteString("* Verify the IOCs due to false positives\n")
	reportFile.WriteString("--------------------------------------------------------\n")

	// Write IPv4 addresses to report file
	for _, ip := range i.IPv4Addresses {
		// Defang the IP address
		ip = strings.ReplaceAll(ip, ".", "[.]")
		reportFile.WriteString("IPv4: " + ip + "\n")
	}

	// Write IPv6 addresses to report file
	for _, ip := range i.IPv6Addresses {
		// Defang the IP address
		ip = strings.ReplaceAll(ip, ":", "[:]")
		reportFile.WriteString("IPv6: " + ip + "\n")
	}

	// Write Domains to report file
	for _, domain := range i.Domains {
		// Defang the domain
		domain = strings.ReplaceAll(domain, ".", "[.]")
		domain = strings.ReplaceAll(domain, "@", "[@]")
		reportFile.WriteString("Domain: " + domain + "\n")
	}

	// Write URLs to report file
	for _, url := range i.URLs {
		url = strings.ReplaceAll(url, ".", "[.]")
		url = strings.ReplaceAll(url, "@", "[@]")
		url = strings.ReplaceAll(url, "://", "[://]")
		url = strings.ReplaceAll(url, "http", "hxxp")
		url = strings.ReplaceAll(url, "https", "hxxps")
		reportFile.WriteString("URL: " + url + "\n")
	}

	// Write Emails to report file
	for _, email := range i.EmailAddresses {
		email = strings.ReplaceAll(email, ".", "[.]")
		email = strings.ReplaceAll(email, "@", "[@]")
		reportFile.WriteString("Email: " + email + "\n")
	}

	reportFile.Close()

	return nil
}

func IPInSubnet(ipStr, cidrStr string) bool {
	// Parse the IP address
	ip := net.ParseIP(ipStr)
	if ip == nil {
		return false
	}

	// Parse the CIDR notation
	_, subnet, err := net.ParseCIDR(cidrStr)
	if err != nil {
		return false
	}

	// Check if the IP is within the subnet
	return subnet.Contains(ip)
}

func (iocs *IOCInformationStruct) RemoveTrustedDomains(trustedDomains []string) {
	var newDomains []string
	// Remove trusted domains
	for _, domain := range iocs.Domains {
		for _, trustedDomain := range trustedDomains {
			if !strings.Contains(domain, trustedDomain) {
				newDomains = append(newDomains, domain)
				break
			}
		}
	}
	var newURLs []string
	for _, url := range iocs.URLs {
		for _, trustedDomain := range trustedDomains {
			if !strings.Contains(url, trustedDomain) {
				newURLs = append(newURLs, url)
				break
			}
		}
	}

	var newEmails []string
	for _, email := range iocs.EmailAddresses {
		for _, trustedDomain := range trustedDomains {
			if !strings.Contains(email, trustedDomain) {
				newEmails = append(newEmails, email)
				break
			}
		}
	}

	iocs.Domains = newDomains
	iocs.URLs = newURLs
	iocs.EmailAddresses = newEmails
}

func (iocs *IOCInformationStruct) RemoveTrustedNetworks(trustedNetworks []string) {
	var newIPv4Addresses, newIPv6Addresses []string
	// Remove trusted IPv4 addresses
	for _, ip := range iocs.IPv4Addresses {
		ipExists := false
		for _, trustedNetwork := range trustedNetworks {
			if IPInSubnet(ip, trustedNetwork) {
				ipExists = true
				break
			}
		}
		if !ipExists {
			newIPv4Addresses = append(newIPv4Addresses, ip)
		}
	}

	// Remove trusted IPv6 addresses
	for _, ip := range iocs.IPv6Addresses {
		ipExists := false
		for _, trustedNetwork := range trustedNetworks {
			if IPInSubnet(ip, trustedNetwork) {
				ipExists = true
				break
			}
		}
		if !ipExists {
			newIPv6Addresses = append(newIPv6Addresses, ip)
		}
	}

	iocs.IPv4Addresses = newIPv4Addresses
	iocs.IPv6Addresses = newIPv6Addresses

}

func main() {
	var iocs, iocsAll IOCInformationStruct
	ConfigPtr := flag.String("config", "config.json", "Configuration file to load for the proxy")

	// Load the Configuration file
	var config Configuration
	configFile := *ConfigPtr
	log.Println("Loading the following config file: " + configFile + "\n")
	if err := config.LoadConfig(configFile); err != nil {
		//fmt.Println("Could not load the configuration file, creating a new default config.json")
		config.CreateConfig(configFile)
		log.Fatalf("Modify the %s file to customize how the tool functions: %v\n", configFile, err)
	}

	//filenames := []string{"email.txt", "info.txt"}

	// Process files that may contain IOCs
	for _, filename := range config.InputFilenames {
		err := iocs.ProcessFile(filename)
		if err != nil {
			fmt.Printf("Error processing %s: %v\n", filename, err)
		}

		iocsAll.IPv4Addresses = append(iocsAll.IPv4Addresses, iocs.IPv4Addresses...)
		iocsAll.IPv6Addresses = append(iocsAll.IPv6Addresses, iocs.IPv6Addresses...)
		iocsAll.Domains = append(iocsAll.Domains, iocs.Domains...)
		iocsAll.URLs = append(iocsAll.URLs, iocs.URLs...)
		iocsAll.EmailAddresses = append(iocsAll.EmailAddresses, iocs.EmailAddresses...)
	}

	// Remove Duplicates
	iocsAll.IPv4Addresses = removeDuplicates(iocsAll.IPv4Addresses)
	iocsAll.IPv6Addresses = removeDuplicates(iocsAll.IPv6Addresses)
	iocsAll.Domains = removeDuplicates(iocsAll.Domains)
	iocsAll.URLs = removeDuplicates(iocsAll.URLs)
	iocsAll.EmailAddresses = removeDuplicates(iocsAll.EmailAddresses)

	// Remove trusted items from the IOCs
	iocsAll.RemoveTrustedNetworks(config.TrustedNetworks)
	iocsAll.RemoveTrustedNetworks(config.InternalNetworks)
	iocsAll.RemoveTrustedNetworks(config.IgnoredNetworks)
	iocsAll.RemoveTrustedDomains(config.TrustedDomains)
	iocsAll.RemoveTrustedDomains(config.IgnoredDomains)

	// Create a report file of the IOCs found
	err := iocsAll.CreateReport(config.ReportFilename)
	if err != nil {
		fmt.Printf("Error creating report: %v\n", err)
	}

	// Create a report file of the IOCs found
	err = iocsAll.CreateReportDefanged(config.ReportFilename)
	if err != nil {
		fmt.Printf("Error creating report: %v\n", err)
	}

	fmt.Printf("Reports created successfully: %s - Defanged: %s\n", config.ReportFilename, "defanged_"+config.ReportFilename)

}
