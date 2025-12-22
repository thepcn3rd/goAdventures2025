package main

// This file contains functions that are used by multiple programs
// Place in the src folder under a folder called commonFunctions
// In the commonFunctions folder after creating and dropping common.go in it
// Execute "go mod init commonFunctions"
// Then the files in common functions can be referenced in import as:
//    cf "commonFunctions"

/*

References:
Generated the base of the code using ChatGPT 10/7/2023
URL: https://github.com/kgretzky/evilginx2/blob/master/core/certdb.go - Learned about the x509 template and extended it

Example JSON config file used to generate new certificates
{
        "DNSNames": [
                "blog.example.com",
                "www.example.com"
        ],
        "Org": "Example Inc",
        "OrgUnit": "",
        "CommonName": "example.com",
        "City": "Lewiston",
        "State": "ID",
        "Country": "US",
        "Email": "admin@example.com"
}

*/

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"math/big"
	"os"
	"time"
)

type certConfig struct {
	DNSNames   []string `json:"DNSNames"`
	Org        string   `json:"Org"`
	OrgUnit    string   `json:"OrgUnit"`
	CommonName string   `json:"CommonName"`
	City       string   `json:"City"`
	State      string   `json:"State"`
	Country    string   `json:"Country"`
	Email      string   `json:"Email"`
}

func CreateCertConfigFile(f string) {
	// Create a new config file
	configFile := `{
		"DNSNames": [
			"blog.example.com",
			"www.example.com"
		],
		"Org": "Example Inc",
		"OrgUnit": "",
		"CommonName": "example.com",
		"City": "Lewiston",
		"State": "ID",
		"Country": "US",
		"Email": "admin@example.com"
	}`
	SaveOutputFile(configFile, f)

	fmt.Printf("\nNew certificate configuration file created: %s\n", f)
	fmt.Println("Edit the configuration file and run the program again to create the self-signed certificate")
	os.Exit(0) // Exit the program after creating the config file

}

func CreateCerts(c string, cert string, key string) {
	//ConfigPtr := flag.String("config", "certConfig.json", "Location of the configuration file for certificate generation")
	//flag.Parse()

	currentDir, _ := os.Getwd()
	configLocation := currentDir + "/" + c

	// Read the configuration file
	configFile, err := os.ReadFile(configLocation)
	CheckError("Failed to read configuration file", err, true)

	// Parse the configuration file
	var certConfig certConfig
	err = json.Unmarshal(configFile, &certConfig)
	CheckError("Failed to parse configuration file", err, true)

	fmt.Printf("Generating certificate and private key from %s...", c)

	// Generate a new private key
	privateKey, err := ecdsa.GenerateKey(elliptic.P384(), rand.Reader)
	CheckError("Failed to generate private key", err, true)

	// Generate values for the below options in the x509 Cert
	// Evaluated how evilginx2 did it in certdb.go
	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		fmt.Printf("Failed to generate serial number: %v\n", err)
		return
	}

	// Create a certificate template with DNSNames (FQDN)
	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Organization:       []string{certConfig.Org},
			OrganizationalUnit: []string{certConfig.OrgUnit},
			CommonName:         certConfig.CommonName,
			Locality:           []string{certConfig.City},    // City
			Province:           []string{certConfig.State},   // State
			Country:            []string{certConfig.Country}, // Country,
		},
		NotBefore:             time.Now().Add(-37 * 24 * time.Hour),
		NotAfter:              time.Now().Add((365 - 37) * 24 * time.Hour), // Valid for 1 year
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		DNSNames:              []string{certConfig.DNSNames[0], certConfig.DNSNames[1]},
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		EmailAddresses:        []string{certConfig.Email},
	}

	// Create a self-signed certificate
	derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, &privateKey.PublicKey, privateKey)
	CheckError("Failed to create certificate: %v\n", err, true)

	// Save the private key to a file
	privateKeyFile, err := os.Create(currentDir + "/" + key)
	CheckError("Failed to create private key file: %v\n", err, true)
	defer privateKeyFile.Close()

	// Save the private key to a file
	privKeyBytes, _ := x509.MarshalECPrivateKey(privateKey)
	pem.Encode(privateKeyFile, &pem.Block{Type: "EC PRIVATE KEY", Bytes: privKeyBytes})

	// Save the certificate to a file
	certificateFile, err := os.Create(currentDir + "/" + cert)
	CheckError("Failed to create certificate file: %v\n", err, true)
	defer certificateFile.Close()
	pem.Encode(certificateFile, &pem.Block{Type: "CERTIFICATE", Bytes: derBytes})
	fmt.Println("Saved the certificate and key")
	fmt.Println("\nCertificate and private key generated successfully!")
}
