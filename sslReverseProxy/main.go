package main

import (
	"crypto/tls"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"log/syslog"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"time"

	cf "github.com/thepcn3rd/goAdvsCommonFunctions"
)

type Configuration struct {
	ListeningPort       string         `json:"listeningPort"`
	ProxiedHTTPEndpoint string         `json:"proxiedHTTPEndpoint"`
	SSLConfig           string         `json:"sslConfig"`
	SSLCert             string         `json:"sslCert"`
	SSLKey              string         `json:"sslKey"`
	SyslogOptions       SyslogConfig   `json:"syslogOptions"`
	SaveFileOptions     SaveFileConfig `json:"saveFileOptions"`
}

type SyslogConfig struct {
	SyslogEnabled    string `json:"syslogEnabled"`
	SyslogServer     string `json:"syslogServer"`
	SyslogOriginName string `json:"syslogOriginName"`
}

type SaveFileConfig struct {
	SaveFileEnabled   string `json:"saveFileEnabled"`
	SaveFileBaseName  string `json:"saveFileBaseName"`
	SaveFileExtension string `json:"saveFileExtension"`
}

var SyslogOriginName string
var SyslogServer string
var SaveFileBaseName string
var SaveFileExtension string

func createReverseProxy(target string) (*httputil.ReverseProxy, error) {
	targetURL, err := url.Parse(target)
	cf.CheckError("Unable to create the reverse proxy connection", err, true)
	return httputil.NewSingleHostReverseProxy(targetURL), nil

	/**
	If SSL30 needs to be supported the below works until it is deprecated from the module

	transport := &http.Transport {
		TLSClientConfig: &tls.Config {
			MinVersion: tls.VersionSSL30,
			InsecureSkipVerify: true,
		},
	}
	proxy := httputil.NewSingleHostReverseProxy(targetURL)
	proxy.Transport = transport
	return proxy, nil

	**/
}

func logToSyslog(message string) {
	logger, err := syslog.Dial("udp", SyslogServer, syslog.LOG_INFO|syslog.LOG_DAEMON, SyslogOriginName)
	if err != nil {
		log.Printf("Failed to connect to remote syslog server: %v", err)
		return
	}
	defer logger.Close()

	logger.Info(message)
}

func logToFile(message string) {
	currentDate := time.Now().Format("2006-01-02")
	fileName := fmt.Sprintf("logs/%s-%s%s", SaveFileBaseName, currentDate, SaveFileExtension)
	file, err := os.OpenFile(fileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Printf("Failed to open log file: %v", err)
		return
	}
	defer file.Close()

	logger := log.New(file, "", log.LstdFlags)
	logger.Println(message)

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
	var config Configuration
	if err := decoder.Decode(&config); err != nil {
		cf.CheckError("Unable to decode the configuration file", err, true)
	}

	// Configuration
	listenAddr := config.ListeningPort        // Port where the proxy will listen
	proxiedAddr := config.ProxiedHTTPEndpoint // Proxied address
	certConfig := config.SSLConfig            // Path to the SSL Configuration
	certFile := config.SSLCert                // Path to the SSL certificate
	keyFile := config.SSLKey                  // Path to the SSL private key

	if config.SyslogOptions.SyslogEnabled == "True" {
		SyslogOriginName = config.SyslogOptions.SyslogOriginName
		SyslogServer = config.SyslogOptions.SyslogServer
	}

	if config.SaveFileOptions.SaveFileEnabled == "True" {
		SaveFileBaseName = config.SaveFileOptions.SaveFileBaseName
		SaveFileExtension = config.SaveFileOptions.SaveFileExtension
	}

	// Setup the SSL Key files
	cf.CreateDirectory("/keys")
	cf.CreateDirectory("/logs")

	// Does the certConfig.json  file exist in the keys folder
	configFileExists := cf.FileExists("/" + certConfig)
	//fmt.Println(configFileExists)
	if !configFileExists {
		cf.CreateCertConfigFile()
		log.Println("WARNING: Created keys/certConfig.json, modify the values to create the self-signed cert to be utilized")
		if config.SyslogOptions.SyslogEnabled == "True" {
			go logToSyslog("WARNING: Created keys/certConfig.json, modify the values to create the self-signed cert to be utilized")
		}
		if config.SaveFileOptions.SaveFileEnabled == "True" {
			go logToFile("WARNING: Created keys/certConfig.json, modify the values to create the self-signed cert to be utilized")
		}
		os.Exit(0)
	}

	// Does the server.crt and server.key files exist in the keys folder
	crtFileExists := cf.FileExists("/" + certFile)
	//keyFileExists := cf.FileExists("/" + keyFile)
	if !crtFileExists {
		cf.CreateCerts()
		//crtFileExists := cf.FileExists("/" + certFile)
		keyFileExists := cf.FileExists("/" + keyFile)
		if !keyFileExists {
			fmt.Println("Failed to create server.crt and server.key files")
			if config.SyslogOptions.SyslogEnabled == "True" {
				go logToSyslog("WARNING: Failed to create server.crt and server.key files for a self-signed certificate")
			}
			if config.SaveFileOptions.SaveFileEnabled == "True" {
				go logToFile("WARNING: Failed to create server.crt and server.key files for a self-signed certificate")
			}
			os.Exit(0)
		}
	}

	// Create reverse proxy
	reverseProxy, err := createReverseProxy(proxiedAddr)
	if err != nil {
		if config.SyslogOptions.SyslogEnabled == "True" {
			go logToSyslog(fmt.Sprintf("Error initializing reverse proxy: %v", err))
		}
		if config.SaveFileOptions.SaveFileEnabled == "True" {
			go logToFile(fmt.Sprintf("Error initializing reverse proxy: %v", err))
		}
		log.Fatalf("Error initializing reverse proxy: %v", err)

	}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		//log.Printf("Proxying request: %s %s", r.Method, r.URL.String())
		proxyLog := fmt.Sprintf("HOST: %s REMOTEADDR: %s METHOD: %s URL: %s REFERER: %s USERAGENT: %s", r.Host, r.RemoteAddr, r.Method, r.URL.String(), r.Referer(), r.UserAgent())
		if config.SyslogOptions.SyslogEnabled == "True" {
			go logToSyslog(proxyLog)
		}
		if config.SaveFileOptions.SaveFileEnabled == "True" {
			go logToFile(proxyLog)
		}
		log.Printf("%s\n", proxyLog)
		reverseProxy.ServeHTTP(w, r)
	})

	// Configure TLS
	tlsConfig := &tls.Config{
		MinVersion: tls.VersionTLS12, // Use modern TLS settings
	}

	httpServer := &http.Server{
		Addr:      listenAddr,
		Handler:   handler,
		TLSConfig: tlsConfig,
	}

	// Start HTTPS server
	log.Printf("Starting reverse SSL proxy on %s, forwarding to %s", listenAddr, proxiedAddr)
	if err := httpServer.ListenAndServeTLS(certFile, keyFile); err != nil {
		log.Fatalf("Failed to start HTTPS Listening Server for the Proxy: %v", err)
	}
}
