package main

import (
	"crypto/tls"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	cf "github.com/thepcn3rd/goAdvsCommonFunctions"
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

func main() {
	urlPtr := flag.String("u", "http://127.0.0.1", "URL to test with the method")
	browserPtr := flag.String("b", "browserHeaders.json", "Use the browser headers file to impersonate a connection")
	flag.Parse()

	// Load browserHeaders.json file
	var config BrowserHeaders
	log.Println("Loading the following browser headers file: " + *browserPtr + "\n")
	configFile, err := os.Open(*browserPtr)
	cf.CheckError("Unable to open the configuration file", err, true)
	defer configFile.Close()
	decoder := json.NewDecoder(configFile)

	if err := decoder.Decode(&config); err != nil {
		cf.CheckError("Unable to decode the configuration file", err, true)
	}

	url := *urlPtr

	// Create a custom HTTP client with a custom Transport to inspect the certificate
	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true, // Skip certificate verification for demonstration purposes
			},
		},
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			// Allow up to 10 redirects
			if len(via) >= 10 {
				return fmt.Errorf("stopped after 10 redirects")
			}
			fmt.Printf("Redirecting to: %s\n", req.URL)
			return nil
		},
	}

	// Create a new HTTP request
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Fatalf("Failed to create request: %v", err)
	}

	// Created the ability to impersonate an actual browser to go through bot filters and WAFs to pull the Certificates
	if len(config.UserAgent) > 0 {
		req.Header.Set("User-Agent", config.UserAgent)
	}
	if len(config.AcceptLanguage) > 0 {
		req.Header.Set("Accept-Language", config.AcceptLanguage)
	}
	if len(config.UpgradeInsecureRequests) > 0 {
		req.Header.Set("Upgrade-Insecure-Requests", config.UpgradeInsecureRequests)
	}
	if len(config.AcceptEncoding) > 0 {
		req.Header.Set("Accept-Encoding", config.AcceptEncoding)
	}
	if len(config.Connection) > 0 {
		req.Header.Set("Connection", config.Connection)
	}

	// Add the additional headers from the browserHeaders.json
	for _, values := range config.AdditionalHeaders {
		if len(values.Key) > 0 {
			req.Header.Set(values.Key, values.Value)
		}
	}
	// Make the request
	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf("Failed to make request: %v", err)
	}
	defer resp.Body.Close()

	// Get the connection state which contains the certificate details
	state := resp.TLS
	if state == nil {
		log.Fatalf("No TLS state found in response")
	}

	// Print the certificate details
	for i, cert := range state.PeerCertificates {
		fmt.Println("----------------------------------------")
		fmt.Printf("Certificate #%d:\n", i+1)
		fmt.Printf("\tSubject: %s\n", cert.Subject)

		fmt.Printf("\tIssuer: %s\n", cert.Issuer)
		fmt.Printf("\t\tCommon Name: %v\n", cert.Issuer.CommonName)
		fmt.Printf("\t\tCountry: %v\n", cert.Issuer.Country)
		fmt.Printf("\t\tIssuer Serial Number: %s\n", cert.Issuer.SerialNumber)
		fmt.Printf("\tNot Before: %s\n", cert.NotBefore.Format(time.RFC3339))
		fmt.Printf("\tNot After: %s\n", cert.NotAfter.Format(time.RFC3339))
		fmt.Printf("\tDNS Names\n")
		for _, name := range cert.DNSNames {
			fmt.Printf("\t\t%s\n", name)
		}
		fmt.Printf("\tSerial Number: %s\n", cert.SerialNumber)
		fmt.Printf("\tSignature Algorithm: %s\n", cert.SignatureAlgorithm)
		fmt.Printf("\tPublic Key Algorithm: %s\n", cert.PublicKeyAlgorithm)
		fmt.Printf("\tVersion: %d\n", cert.Version)
		fmt.Printf("\tEmails: %s\n", cert.EmailAddresses)
		fmt.Printf("\tIP Addresses: %v\n", cert.IPAddresses)

	}
}
