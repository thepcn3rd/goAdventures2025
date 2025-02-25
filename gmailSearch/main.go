package main

import (
	"context"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/gmail/v1"
	"google.golang.org/api/option"

	cf "github.com/thepcn3rd/goAdvsCommonFunctions"
)

type Configuration struct {
	ListeningPort    string `json:"listeningPort"`
	TLSEnabled       bool   `json:"tlsEnabled"`
	SSLConfig        string `json:"sslConfig"`
	SSLCert          string `json:"sslCert"`
	SSLKey           string `json:"sslKey"`
	CredentialFile   string `json:"credentialFile"`
	TokenFile        string `json:"tokenFile"`
	ArchiveDirectory string `json:"archiveDirectory"`
	Query            string `json:"query"`
}

var authCode string // Code Received by the HTTP Server with the GET Request sent by OAuth2
var config Configuration

// Retrieve a token, saves the token, then returns the generated client.
func getClient(configOAuth *oauth2.Config) *http.Client {
	// The file token.json stores the user's access and refresh tokens, and is
	// created automatically when the authorization flow completes for the first
	// time.
	tokenFile := config.TokenFile
	token, err := tokenFromFile(tokenFile)
	if err != nil {
		token = getTokenFromWeb(configOAuth)
		saveToken(tokenFile, token)
	}
	return configOAuth.Client(context.Background(), token)
}

// Build a simple website to receive the token
func handler(w http.ResponseWriter, r *http.Request) {
	//fmt.Printf("URL: %s\n", r.URL.String())
	//fmt.Printf("Method: %s\n", r.Method)
	//fmt.Printf("Headers: %v\n", r.Header)
	if len(r.URL.Query().Get("code")) > 1 {
		fmt.Printf("Authentication Code: %v\n", r.URL.Query().Get("code"))
		authCode = r.URL.Query().Get("code")
	}
	//r.URL.Query().Get("code")

	fmt.Fprintf(w, "Go!")
}

// Request a token from the web, then returns the retrieved token.
func getTokenFromWeb(configOAuth *oauth2.Config) *oauth2.Token {
	// Generate or verify the keyFile and certFiles exist for the TLS connection
	cf.CreateDirectory("/keys")
	// Does the certConfig.json  file exist in the keys folder
	configFileExists := cf.FileExists("/" + config.SSLConfig)
	//fmt.Println(configFileExists)
	if !configFileExists {
		cf.CreateCertConfigFile()
		log.Println("WARNING: Created keys/certConfig.json, modify the values to create the self-signed cert to be utilized")
		os.Exit(0)
	}

	// Does the server.crt and server.key files exist in the keys folder
	crtFileExists := cf.FileExists("/" + config.SSLCert)
	//keyFileExists := cf.FileExists("/" + keyFile)
	if !crtFileExists {
		cf.CreateCerts()
		//crtFileExists := cf.FileExists("/" + certFile)
		keyFileExists := cf.FileExists("/" + config.SSLKey)
		if !keyFileExists {
			fmt.Println("Failed to create server.crt and server.key files")
			os.Exit(0)
		}
	}

	// I do not like that the authURL has GET parameters related to the authentication...
	authURL := configOAuth.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Printf("\nGo to the following link in your browser then the authorization code will be sent: \n\n%v\n\n", authURL)

	var server *http.Server
	if config.TLSEnabled {
		// Create an HTTP server
		tlsConfig := &tls.Config{
			MinVersion: tls.VersionTLS12,
		}

		server = &http.Server{
			Addr:      config.ListeningPort,
			Handler:   http.HandlerFunc(handler),
			TLSConfig: tlsConfig,
		}
	} else {
		server = &http.Server{
			Addr:    config.ListeningPort,
			Handler: http.HandlerFunc(handler),
		}
	}

	go func() {
		fmt.Println("Starting HTTP Server on :8080 to catch the OAuth2 Code. The redirect URI in the credentials.json file needs to have http(s)://localhost:8080...")
		if err := server.ListenAndServeTLS(config.SSLCert, config.SSLKey); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	// Wait until code is received by GET request...
	for {
		if authCode != "" {
			break
		}
	}

	// Create a context with a timeout for graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Gracefully shutdown the server
	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Forced to shutdown: %v", err)
	}

	fmt.Printf("HTTP Server stopped gracefully\n\n")

	// Automated the reception of the code through running a server on port 8080
	token, err := configOAuth.Exchange(context.TODO(), authCode)
	if err != nil {
		log.Fatalf("Unable to retrieve token from web: %v", err)
	}
	return token
}

// Retrieves a token from a local file.
func tokenFromFile(file string) (*oauth2.Token, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	token := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(token)
	return token, err
}

// Saves a token to a file path.
func saveToken(path string, token *oauth2.Token) {
	fmt.Printf("Saving credential file to: %s\n\n", path)
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		log.Fatalf("Unable to cache oauth token: %v", err)
	}
	defer f.Close()
	json.NewEncoder(f).Encode(token)
}

func getHeader(headers []*gmail.MessagePartHeader, name string) string {
	for _, h := range headers {
		// List the headers returned...
		//fmt.Printf("Header Name: %s\n", h.Name)
		if h.Name == name {
			return h.Value
		}
	}
	return ""
}

func main() {
	ConfigPtr := flag.String("config", "config.json", "Configuration file to load for the proxy")
	flag.Parse()

	// Load config.json file
	log.Println("Loading the following config file: " + *ConfigPtr + "\n")
	//go logToSyslog(fmt.Sprintf("Loading the following config file: %s\n", *ConfigPtr))
	configFile, err := os.Open(*ConfigPtr)
	cf.CheckError("Unable to open the configuration file", err, true)
	defer configFile.Close()
	decoder := json.NewDecoder(configFile)
	//var config Configuration
	if err := decoder.Decode(&config); err != nil {
		cf.CheckError("Unable to decode the configuration file", err, true)
	}

	ctx := context.Background()
	b, err := os.ReadFile(config.CredentialFile)
	if err != nil {
		log.Fatalf("Unable to read client secret file: %v", err)
	}

	// If modifying these scopes, delete your previously saved token.json.
	// Below scope is for gmail.readonly
	configGoogle, err := google.ConfigFromJSON(b, gmail.GmailReadonlyScope)
	// Below scope is for gmail.modify
	//config, err := google.ConfigFromJSON(b, gmail.GmailModifyScope)
	if err != nil {
		log.Fatalf("Unable to parse client secret file to config: %v", err)
	}
	client := getClient(configGoogle)

	srv, err := gmail.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		log.Fatalf("Unable to retrieve Gmail client: %v", err)
	}

	user := "me"
	query := config.Query
	r, err := srv.Users.Messages.List(user).Q(query).Do()
	if err != nil {
		log.Fatalf("Unable to retrieve messages: %v", err)
	}

	if len(r.Messages) == 0 {
		fmt.Println("No messages found.")
		return
	}

	fmt.Println("Messages:")
	for _, m := range r.Messages {
		msg, err := srv.Users.Messages.Get(user, m.Id).Do()
		if err != nil {
			log.Printf("Unable to retrieve message %s: %v", m.Id, err)
			continue
		}
		fmt.Printf("\nMessage ID: %s\nFrom: %s\nSubject: %s\nDate: %s\n\n", m.Id, getHeader(msg.Payload.Headers, "From"), getHeader(msg.Payload.Headers, "Subject"), getHeader(msg.Payload.Headers, "Date"))
		fmt.Printf("Move message to trash (y/n): ")
		var responseTrash string
		if _, err := fmt.Scan(&responseTrash); err != nil {
			log.Fatalf("Unable to read response for moving to trash: %v", err)
		}
		// To move the messages to trash you need the gmail.modify permissions setup
		if responseTrash == "y" || responseTrash == "Y" {
			_, err := srv.Users.Messages.Trash(user, msg.Id).Do()
			if err != nil {
				log.Printf("Unable to move message %s to trash.\nError: %v\n", m.Id, err)
				continue
			}
		}

		// Archive messages to a file (Add this function...)
		fmt.Printf("\n\nSave message to an archive file (y/n): ")
		var responseArchive string
		if _, err := fmt.Scan(&responseArchive); err != nil {
			log.Fatalf("Unable to read response for archiving the message: %v", err)
		}
		if responseArchive == "y" || responseArchive == "Y" {
			msgRaw, err := srv.Users.Messages.Get(user, m.Id).Format("raw").Do()
			if err != nil {
				log.Fatalf("Unable to retrieve message: %v", err)
			}

			// Decode the raw message from base64
			decodedMessage, err := base64.URLEncoding.DecodeString(msgRaw.Raw)
			if err != nil {
				log.Fatalf("Unable to decode message: %v", err)
			}

			outputMessage := fmt.Sprintf("Message ID: %s\nFrom: %s\nSubject: %s\nDate: %s\n\n", m.Id, getHeader(msg.Payload.Headers, "From"), getHeader(msg.Payload.Headers, "Subject"), getHeader(msg.Payload.Headers, "Date"))
			// The raw message is not working...
			outputMessage += fmt.Sprintf("Raw Message:\n\n%s\n", string(decodedMessage))
			cf.CreateDirectory("/" + config.ArchiveDirectory)
			cf.SaveOutputFile(outputMessage, config.ArchiveDirectory+"/msg_"+m.Id+".txt")
		}
	}
}
