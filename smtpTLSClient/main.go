package main

import (
	"bytes"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"mime"
	"mime/multipart"
	"net/smtp"
	"net/textproto"
	"os"
	"path/filepath"
	"strings"

	cf "github.com/thepcn3rd/goAdvsCommonFunctions"
)

// TLS Output Key File to use in Wireshark for Decryption
// - I ran into a scenario where I needed to see the TLS Output Key File and added to the config.json file the capability to capture the TLS Key to a file
// - In Wireshark...  Edit -> Preferences --> Protocols --> TLS (Then upload to Pre-Master Secret Log File)

// Supports Unencrypted port 25 with no auth
// - Change the port to 25
// - Leave blank the smtpUsername and smtpPassword

// Supports adding an attachment
// - Leave blank if no attachment and then specify the path to load attachment from...

// If you want to send an html file then create and add to the HTMLBodyFile in the config.json
// If you do not want to send HTML leave the file blank
// If you do not want to send the plain-text then leave the Body blank
// Populate both if you want to send HTML and plain-text
// Note: Yahoo displays the HTML with the image from test.html, gmail shows the Alt text

// sample.png file is present in the files on github, this is the base64 encoded image placed in test.html
// test.html is a sample html file that was sent testing the client sending HTML files

// HTB:axlle
// HTB:mailing

// Future Enhancements...
// Need to come back to this and configure the SMTP encrypted password in the config.json file
// Allow the option to use SSL

type Configuration struct {
	SMTPHost      string   `json:"smtpHost"`
	SMTPPort      string   `json:"smtpPort"`
	SMTPUsername  string   `json:"smtpUsername"`
	SMTPPassword  string   `json:"smtpPassword"`
	FromAddress   string   `json:"fromAddress"`
	ToAddress     []string `json:"toAddresses"`
	Subject       string   `json:"subject"`
	Body          string   `json:"body"`
	HTMLBodyFile  string   `json:"htmlBodyFile"`
	CaptureTLSKey bool     `json:"captureTLSKey"`
	OutputTLSFile string   `json:"outputFile"`
	Attachment    string   `json:"attachment"`
}

func (c *Configuration) CreateConfig() error {
	c.SMTPHost = "thepcn3rd.local"
	c.SMTPPort = "587"
	c.SMTPUsername = "thepcn3rd@thepcn3rd.local"
	c.SMTPPassword = "AVeryGoodPasswordthatisUnguessable!!"
	c.FromAddress = "thepcn3rd@thepcn3rd.local"
	c.ToAddress = []string{}
	c.ToAddress = append(c.ToAddress, "thepcn3rd@thepcn3rd.local")
	c.ToAddress = append(c.ToAddress, "thebabyn3rd@thepcn3rd.local")
	c.Subject = "Test Email"
	c.Body = "Switch to HTML format to read this email..."
	c.HTMLBodyFile = "test.html"
	c.CaptureTLSKey = false
	c.OutputTLSFile = "tlsKey.log"
	c.Attachment = ""

	jsonData, err := json.MarshalIndent(c, "", "    ")
	if err != nil {
		return err
	}

	err = os.WriteFile("config.json", jsonData, 0644)
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

func splitBase64(s string) string {
	var buf bytes.Buffer
	for len(s) > 0 {
		chunkSize := 76
		if len(s) < chunkSize {
			chunkSize = len(s)
		}
		buf.WriteString(s[:chunkSize] + "\n")
		s = s[chunkSize:]
	}
	return buf.String()
}

func readFileToBytes(filename string) ([]byte, error) {
	fileData, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer fileData.Close()

	// Get the file size
	fileInfo, err := fileData.Stat()
	if err != nil {
		log.Fatalln("Error getting file info:", err)
	}
	fileSize := fileInfo.Size()

	// Create a byte slice of the appropriate size
	fileBytes := make([]byte, fileSize)

	// Read the file content into the byte slice
	_, err = fileData.Read(fileBytes)
	if err != nil && err != io.EOF {
		log.Fatalln("Error reading file:", err)
	}

	return fileBytes, err
}

func createEmailWithAttachment(config Configuration) ([]byte, error) {
	var buf []byte
	//body := config.Body

	// Create a new multipart writer
	bodyBuf := &bytes.Buffer{}
	writer := multipart.NewWriter(bodyBuf)

	// Add the email body in plain text
	if config.Body != "" {
		part, err := writer.CreatePart(textproto.MIMEHeader{
			"Content-Type": []string{"text/plain; charset=utf-8"},
		})
		if err != nil {
			return nil, err
		}
		part.Write([]byte(config.Body))
	}

	// Add the HTML body file if present
	if config.HTMLBodyFile != "" {
		htmlBodyBytes, err := readFileToBytes(config.HTMLBodyFile)
		if err != nil {
			return nil, err
		}
		part, err := writer.CreatePart(textproto.MIMEHeader{
			"Content-Type": []string{"text/html; charset=utf-8"},
		})
		if err != nil {
			return nil, err
		}
		part.Write(htmlBodyBytes)
	}

	// Add the attachment if specified
	if config.Attachment != "" {
		fileBytes, err := readFileToBytes(config.Attachment)
		if err != nil {
			return nil, err
		}

		// You can use this filter in wireshark to see the transfer encoding...
		// _ws.col.protocol == "SMTP/IMF"
		// Encode the file content in base64
		encodedData := base64.StdEncoding.EncodeToString(fileBytes)
		encodedData = splitBase64(encodedData)

		part, err := writer.CreatePart(textproto.MIMEHeader{
			// "Content-Type: application/vnd.ms-excel"
			// You could force the "Content-Type" to bypass email filters on MIME Type for the attachment
			// "Content-Type: application/octet-stream"
			"Content-Type":              []string{mime.TypeByExtension(filepath.Ext(config.Attachment))},
			"Content-Disposition":       []string{fmt.Sprintf("attachment; filename=\"%s\"", filepath.Base(config.Attachment))},
			"Content-Transfer-Encoding": []string{"BASE64"},
		})
		if err != nil {
			return nil, err
		}
		part.Write([]byte(encodedData))
	}

	writer.Close()

	// Combine the headers and the body
	headers := fmt.Sprintf("Subject: %s\n", config.Subject)
	headers += fmt.Sprintf("To: %s\n", strings.Join(config.ToAddress, ", "))
	headers += fmt.Sprintf("From: %s\n", config.FromAddress)
	headers += fmt.Sprintf("MIME-Version: 1.0\n")
	headers += fmt.Sprintf("Content-Type: multipart/mixed; boundary=%s\n", writer.Boundary())

	buf = append([]byte(headers+"\n"), bodyBuf.Bytes()...)
	return buf, nil
}

func main() {
	var config Configuration
	ConfigPtr := flag.String("config", "config.json", "Configuration file to load for the proxy")
	flag.Parse()

	fmt.Println("Loading the following config file: " + *ConfigPtr + "\n")
	if err := config.LoadConfig(*ConfigPtr); err != nil {
		//fmt.Println("Could not load the configuration file, creating a new default config.json")
		config.CreateConfig()
		log.Fatalf("Modify the config.json file to customize how the tool functions: %v\n", err)
	}

	var keyfile *os.File
	if config.CaptureTLSKey && config.SMTPPort == "587" {
		keyfile, err := os.OpenFile(config.OutputTLSFile, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
		if err != nil {
			log.Fatalf("Failed to create key log file: %v", err)
		}
		defer keyfile.Close()
		log.Println("TLS keys will be logged to:", config.OutputTLSFile)
	}

	// Create the email with an attachment
	msg, err := createEmailWithAttachment(config)
	if err != nil {
		log.Fatalf("Error creating email: %v", err)
	}

	// Authentication
	var auth smtp.Auth
	if config.SMTPUsername != "" && config.SMTPPassword != "" {
		auth = smtp.PlainAuth("", config.SMTPUsername, config.SMTPPassword, config.SMTPHost)
	} else {
		log.Println("No authentication credentials provided, sending email without.")
	}

	// Connect to the SMTP server
	client, err := smtp.Dial(config.SMTPHost + ":" + config.SMTPPort)
	cf.CheckError("Error connecting to SMTP server:", err, true)

	// Start TLS if using port 587
	if config.SMTPPort == "587" {
		var tlsConfig *tls.Config
		if config.CaptureTLSKey {
			// TLS configuration to bypass self-signed certificate verification
			tlsConfig = &tls.Config{
				InsecureSkipVerify: true, // Bypass self-signed certificate verification
				ServerName:         config.SMTPHost,
				KeyLogWriter:       keyfile,
			}
		} else {
			tlsConfig = &tls.Config{
				InsecureSkipVerify: true, // Bypass self-signed certificate verification
				ServerName:         config.SMTPHost,
			}
		}

		if err = client.StartTLS(tlsConfig); err != nil {
			fmt.Println("Error starting TLS:", err)
			os.Exit(1)
		}
	}

	// Authenticate
	if auth != nil {
		if err = client.Auth(auth); err != nil {
			fmt.Println("Error authenticating:", err)
			os.Exit(1)
		}
	}

	// Set the sender
	if err = client.Mail(config.FromAddress); err != nil {
		fmt.Println("Error setting sender:", err)
		os.Exit(1)
	}

	for _, recipient := range config.ToAddress {
		if err = client.Rcpt(recipient); err != nil {
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

	log.Println("Email sent successfully!")

	keyfile.Close()
}
