package main

import (
	"bufio"
	"crypto/tls"
	"encoding/csv"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

type BrowserHeaders struct {
	UserAgent               string                    `json:"userAgent"`
	AcceptLanguage          string                    `json:"acceptLanguage"`
	UpgradeInsecureRequests string                    `json:"upgradeInsecureRequests"`
	AcceptEncoding          string                    `json:"acceptEncoding"`
	Connection              string                    `json:"connection"`
	AdditionalHeaders       []additionalHeadersStruct `json:"additionalHeaders"`
}

type additionalHeadersStruct struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type CertList struct {
	List []CertInformation `json:"list"`
}

type CertInformation struct {
	URL   string       `json:"url"`
	Notes string       `json:"notes"`
	Certs []CertStruct `json:"certs"`
}

type CertStruct struct {
	Subject            string   `json:"subject"`
	Issuer             string   `json:"issuer"`
	CommonName         string   `json:"commonname"`
	Country            string   `json:"country"`
	IssuerSerialNumber string   `json:"issuerserialnumber"`
	NotBefore          string   `json:"notbefore"`
	NotAfter           string   `json:"notafter"`
	DNSNames           []string `json:"dnsnames"`
	SerialNumber       string   `json:"serialnumber"`
	SignatureAlgo      string   `json:"signaturealgo"`
	PublicKeyAlgo      string   `json:"publickeyalgo"`
	Version            string   `json:"version"`
	Emails             string   `json:"emailaddresses"`
	IPAddresses        string   `json:"ipaddresses"`
}

func (c *CertList) OutputCertJSON() error {
	jsonData, err := json.MarshalIndent(c, "", "    ")
	if err != nil {
		return err
	}

	err = os.WriteFile("certInformation.json", jsonData, 0644)
	if err != nil {
		return err
	}

	return nil
}

func (c *BrowserHeaders) LoadConfig(cPtr string) error {
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

func (b *BrowserHeaders) CreateConfig() error {
	b.UserAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/124.0.6367.118 Safari/537.36"
	b.AcceptLanguage = "en-US"
	b.UpgradeInsecureRequests = "1"
	b.AcceptEncoding = "gzip, deflate, br"
	b.Connection = "keep-alive"
	var header additionalHeadersStruct
	header.Key = "Content-Type"
	header.Value = "text/plain;charset=UTF-8"
	b.AdditionalHeaders = append(b.AdditionalHeaders, header)

	jsonData, err := json.MarshalIndent(b, "", "    ")
	if err != nil {
		return err
	}

	err = os.WriteFile("browserHeaders.json", jsonData, 0644)
	if err != nil {
		return err
	}

	return nil
}

func (b *BrowserHeaders) RequestURL(url string) (*tls.ConnectionState, error) {
	// Create a custom HTTP client with a custom Transport to inspect the certificate
	client := &http.Client{
		Transport: &http.Transport{
			DialContext: (&net.Dialer{
				Timeout: 5 * time.Second, // Timeout for establishing the TCP connection
			}).DialContext,
			TLSHandshakeTimeout:   5 * time.Second,  // Timeout for TLS handshake
			ResponseHeaderTimeout: 10 * time.Second, // Timeout for receiving response headers
			ExpectContinueTimeout: 1 * time.Second,  // Timeout for Expect: 100-continue
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true, // Skip certificate verification for demonstration purposes
			},
		},
		Timeout: 30 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			// Allow up to 3 redirects
			/**
			if len(via) >= 3 {
				return fmt.Errorf("stopped after 3 redirects")
			}
			fmt.Printf("Redirecting to: %s\n", req.URL)
			return nil
			**/
			return http.ErrUseLastResponse // Do not allow any redirects...
		},
	}

	// Create a new HTTP request
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Fatalf("Failed to create request: %v", err)
	}

	// Created the ability to impersonate an actual browser to go through bot filters and WAFs to pull the Certificates
	if len(b.UserAgent) > 0 {
		req.Header.Set("User-Agent", b.UserAgent)
	}
	if len(b.AcceptLanguage) > 0 {
		req.Header.Set("Accept-Language", b.AcceptLanguage)
	}
	if len(b.UpgradeInsecureRequests) > 0 {
		req.Header.Set("Upgrade-Insecure-Requests", b.UpgradeInsecureRequests)
	}
	if len(b.AcceptEncoding) > 0 {
		req.Header.Set("Accept-Encoding", b.AcceptEncoding)
	}
	if len(b.Connection) > 0 {
		req.Header.Set("Connection", b.Connection)
	}

	// Add the additional headers from the browserHeaders.json
	for _, values := range b.AdditionalHeaders {
		if len(values.Key) > 0 {
			req.Header.Set(values.Key, values.Value)
		}
	}
	// Make the request
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Failed to make request: %v", err)
		return nil, err
	}
	defer resp.Body.Close()

	// Get the connection state which contains the certificate details
	state := resp.TLS
	if state == nil {
		log.Printf("No TLS state found in response")
	}

	return state, nil
}

func RemoveBadChars(s string) string {
	s = strings.Replace(s, "\r", "", -1)
	s = strings.Replace(s, "\n", "", -1)
	return s
}

func InputFromFile(f string) []string {
	outputStrings := []string{}
	file, err := os.Open(f)
	if err != nil {
		log.Printf("[E] Failed to open file: %s", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		line = RemoveBadChars(line)
		outputStrings = append(outputStrings, line)
	}
	return outputStrings
}

func main() {
	urlPtr := flag.String("u", "http://127.0.0.1", "URL to test with the method")
	urlListPtr := flag.String("f", "", "URL List")
	browserPtr := flag.String("b", "browserHeaders.json", "Use the browser headers file to impersonate a connection")
	flag.Parse()

	// Load browserHeaders.json file
	var config BrowserHeaders
	if err := config.LoadConfig(*browserPtr); err != nil {
		config.CreateConfig()
		log.Fatalln("Unable to Load the Browser Headers JSON file")
	}

	var inputURLs []string
	if *urlPtr != "http://127.0.0.1" {
		inputURLs = append(inputURLs, *urlPtr)
	} else if len(*urlListPtr) > 0 {
		inputURLs = InputFromFile(*urlListPtr)
	} else {
		flag.Usage()
		log.Fatalln("Unable to load a URL or URL List")
	}

	var certList CertList
	for _, url := range inputURLs {
		fmt.Printf("Analyzing URL: %s\n", url)
		var certInfo CertInformation
		state, err := config.RequestURL(url)
		if err != nil {
			log.Println("Unable to Request URL", url)
			// Record in the struct in a notes field the error
			certInfo.URL = url
			certInfo.Notes = err.Error()
			certList.List = append(certList.List, certInfo)
			continue
		}
		certInfo.URL = url

		// Print the certificate details
		//var certs []certStruct
		for i, cert := range state.PeerCertificates {
			var c CertStruct
			fmt.Println("----------------------------------------")
			fmt.Printf("Certificate #%d:\n", i+1)
			fmt.Printf("\tSubject: %s\n", cert.Subject)
			c.Subject = fmt.Sprintf("%s", cert.Subject)
			fmt.Printf("\tIssuer: %s\n", cert.Issuer)
			c.Issuer = cert.Issuer.String()
			fmt.Printf("\t\tCommon Name: %v\n", cert.Issuer.CommonName)
			c.CommonName = cert.Issuer.CommonName
			fmt.Printf("\t\tCountry: %v\n", cert.Issuer.Country)
			c.Country = fmt.Sprintf("%v", cert.Issuer.Country)
			fmt.Printf("\t\tIssuer Serial Number: %s\n", cert.Issuer.SerialNumber)
			c.IssuerSerialNumber = cert.Issuer.SerialNumber
			fmt.Printf("\tNot Before: %s\n", cert.NotBefore.Format(time.RFC3339))
			c.NotBefore = cert.NotBefore.Format(time.RFC3339)
			fmt.Printf("\tNot After: %s\n", cert.NotAfter.Format(time.RFC3339))
			c.NotAfter = cert.NotAfter.Format(time.RFC3339)
			fmt.Printf("\tDNS Names\n")
			for _, name := range cert.DNSNames {
				c.DNSNames = append(c.DNSNames, name)
				fmt.Printf("\t\t%s\n", name)
			}
			fmt.Printf("\tSerial Number: %s\n", cert.SerialNumber)
			c.SerialNumber = cert.SerialNumber.String()
			fmt.Printf("\tSignature Algorithm: %s\n", cert.SignatureAlgorithm)
			c.SignatureAlgo = cert.SignatureAlgorithm.String()
			fmt.Printf("\tPublic Key Algorithm: %s\n", cert.PublicKeyAlgorithm)
			c.PublicKeyAlgo = cert.PublicKeyAlgorithm.String()
			fmt.Printf("\tVersion: %d\n", cert.Version)
			c.Version = strconv.Itoa(cert.Version)
			fmt.Printf("\tEmails: %s\n", cert.EmailAddresses)
			c.Emails = fmt.Sprintf("%v", cert.EmailAddresses)
			fmt.Printf("\tIP Addresses: %v\n", cert.IPAddresses)
			c.IPAddresses = fmt.Sprintf("%v", cert.IPAddresses)

			certInfo.Certs = append(certInfo.Certs, c)

		}
		certList.List = append(certList.List, certInfo)
	}

	if err := certList.OutputCertJSON(); err != nil {
		log.Println("Unable to save the certificate information to JSON")
	}

	// Create CSV File
	csvFile, err := os.Create("certInformation.csv")
	if err != nil {
		fmt.Printf("Error creating CSV file: %v\n", err)
		return
	}
	defer csvFile.Close()

	writer := csv.NewWriter(csvFile)
	defer writer.Flush()

	// Write csv header
	header := []string{"URL", "Notes", "Subject", "Issuer", "CommonName", "Country", "IssuerSerialNumber", "NotBefore", "NotAfter", "DNSNames", "SerialNumber", "SignatureAlgo", "PublicKeyAlgo", "Version", "Emails", "IPAddresses"}
	err = writer.Write(header)
	if err != nil {
		fmt.Printf("Error writing CSV header: %v\n", err)
		return
	}

	for _, listCSV := range certList.List {
		for _, certInfoCSV := range listCSV.Certs {
			DNSNamesList := ""
			for _, dnsName := range certInfoCSV.DNSNames {
				DNSNamesList = DNSNamesList + dnsName + ";"
			}
			csvLine := []string{listCSV.URL,
				listCSV.Notes,
				certInfoCSV.Subject,
				certInfoCSV.Issuer,
				certInfoCSV.CommonName,
				certInfoCSV.Country,
				certInfoCSV.IssuerSerialNumber,
				certInfoCSV.NotBefore,
				certInfoCSV.NotAfter,
				DNSNamesList,
				certInfoCSV.SerialNumber,
				certInfoCSV.SignatureAlgo,
				certInfoCSV.PublicKeyAlgo,
				certInfoCSV.Version,
				certInfoCSV.Emails,
				certInfoCSV.IPAddresses,
			}
			err = writer.Write(csvLine)
			if err != nil {
				fmt.Printf("Error writing CSV record: %v\n", err)
				return
			}
		}

	}
}
