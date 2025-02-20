package main

import (
	"crypto/tls"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/smtp"
	"os"

	cf "github.com/thepcn3rd/goAdvsCommonFunctions"
)

// TLS Output Key File to use in Wireshark for Decryption
// - I ran into a scenario where I needed to see the TLS Output Key File and added to the config.json file the capability to capture the TLS Key to a file
// - In Wireshark...  Edit -> Preferences --> Protocols --> TLS (Then upload to Pre-Master Secret Log File)

// Future Enhancements...
// Need to come back to this and configure the SMTP encrypted password in the config.json file
// Allow the option to use SSL and plain auth (This version only supports TLS)
// Allow for reading from a .txt file or a .html file and encode for email type

type Configuration struct {
	SMTPHost              string            `json:"smtpHost"`
	SMTPPort              string            `json:"smtpPort"`
	SMTPUsername          string            `json:"smtpUsername"`
	SMTPEncryptedPassword string            `json:"smtpEncryptedPassword"`
	FromAddress           string            `json:"fromAddress"`
	ToAddress             []toAddressStruct `json:"toAddresses"`
	Subject               string            `json:"subject"`
	Body                  string            `json:"body"`
	CaptureTLSKey         bool              `json:"captureTLSKey"`
	OutputFile            string            `json:"outputFile"`
}

type toAddressStruct struct {
	Email string `json:"email"`
}

func loadConfig(cPtr string) Configuration {

	var c Configuration
	fmt.Println("Loading the following config file: " + cPtr + "\n")
	// go logToSyslog(fmt.Sprintf("Loading the following config file: %s\n", *ConfigPtr))
	configFile, err := os.Open(cPtr)
	cf.CheckError("Unable to open the configuration file", err, true)
	defer configFile.Close()
	decoder := json.NewDecoder(configFile)
	// var config Configuration
	if err := decoder.Decode(&c); err != nil {
		cf.CheckError("Unable to decode the configuration file", err, true)
	}

	return c
}

func main() {
	var config Configuration
	ConfigPtr := flag.String("config", "config.json", "Configuration file to load for the proxy")
	flag.Parse()

	config = loadConfig(*ConfigPtr)

	var keyfile *os.File
	if config.CaptureTLSKey {
		keyLogFile := config.OutputFile
		keyfile, err := os.OpenFile(keyLogFile, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
		if err != nil {
			log.Fatalf("Failed to create key log file: %v", err)
		}
		defer keyfile.Close()
		log.Println("TLS keys will be logged to:", keyLogFile)
	}

	// SMTP server configuration
	smtpHost := config.SMTPHost
	smtpPort := config.SMTPPort
	smtpUsername := config.SMTPUsername
	smtpPassword := config.SMTPEncryptedPassword

	// Sender and recipient
	from := config.FromAddress
	to := config.ToAddress

	// Email content
	subject := "Subject: " + config.Subject + "\n"
	body := config.Body + "\n"
	msg := []byte(subject + "\n" + body)

	// Authentication
	auth := smtp.PlainAuth("", smtpUsername, smtpPassword, smtpHost)

	var tlsConfig *tls.Config
	if config.CaptureTLSKey {
		// TLS configuration to bypass self-signed certificate verification
		tlsConfig = &tls.Config{
			InsecureSkipVerify: true, // Bypass self-signed certificate verification
			ServerName:         smtpHost,
			KeyLogWriter:       keyfile,
		}
	} else {
		tlsConfig = &tls.Config{
			InsecureSkipVerify: true, // Bypass self-signed certificate verification
			ServerName:         smtpHost,
		}
	}

	// Connect to the SMTP server
	client, err := smtp.Dial(smtpHost + ":" + smtpPort)
	cf.CheckError("Error connecting to SMTP server:", err, true)

	// Start TLS
	if err = client.StartTLS(tlsConfig); err != nil {
		fmt.Println("Error starting TLS:", err)
		os.Exit(1)
	}

	// Authenticate
	if err = client.Auth(auth); err != nil {
		fmt.Println("Error authenticating:", err)
		os.Exit(1)
	}

	// Set the sender
	if err = client.Mail(from); err != nil {
		fmt.Println("Error setting sender:", err)
		os.Exit(1)
	}

	for _, recipient := range to {
		if err = client.Rcpt(recipient.Email); err != nil {
			fmt.Println("Error setting recipient:", err)
			os.Exit(1)
		}
	}

	w, err := client.Data()
	cf.CheckError("Error preparing email body:", err, true)

	_, err = w.Write(msg)
	cf.CheckError("Error writing email body:", err, true)

	err = w.Close()
	cf.CheckError("Error closing email body:", err, true)

	// Quit the client
	err = client.Quit()
	cf.CheckError("Error quitting client:", err, true)

	fmt.Println("Email sent successfully!")

	keyfile.Close()
}
