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
	BasicAuthOptions    AuthConfig     `json:"basicAuthOptions"`
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

type AuthConfig struct {
	Enabled  bool   `json:"enabled"`
	Username string `json:"username"`
	Password string `json:"password"`
}

func (c *Configuration) CreateConfig() error {
	c.ListeningPort = "0.0.0.0:8443"
	c.ProxiedHTTPEndpoint = "http://127.0.0.1:8080"
	c.SSLConfig = "keys/certConfig.json"
	c.SSLCert = "keys/server.crt"
	c.SSLKey = "keys/server.key"
	c.SyslogOptions.SyslogEnabled = "True"
	c.SyslogOptions.SyslogServer = "10.10.10.10:514"
	c.SyslogOptions.SyslogOriginName = "proxy-server"
	c.SaveFileOptions.SaveFileEnabled = "True"
	c.SaveFileOptions.SaveFileBaseName = "proxy"
	c.SaveFileOptions.SaveFileExtension = ".log"
	c.BasicAuthOptions.Enabled = true
	c.BasicAuthOptions.Username = "thepcn3rd"
	c.BasicAuthOptions.Password = "T3sting"

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

func authConnection(next http.HandlerFunc, c Configuration) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if c.BasicAuthOptions.Enabled {
			username, password, ok := r.BasicAuth()
			sha256InputPassword := cf.CalcSHA256Hash(password)
			sha256StoredPassword := cf.CalcSHA256Hash(c.BasicAuthOptions.Password)
			if !ok || username != c.BasicAuthOptions.Username || sha256InputPassword != sha256StoredPassword {
				w.Header().Set("WWW-Authenticate", `Basic realm="Restricted"`)
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}
		}
		next(w, r)
	}
}

func main() {
	var config Configuration
	ConfigPtr := flag.String("config", "config.json", "Configuration file to load for the proxy")
	flag.Parse()

	log.Println("Loading the following config file: " + *ConfigPtr + "\n")
	if err := config.LoadConfig(*ConfigPtr); err != nil {
		//fmt.Println("Could not load the configuration file, creating a new default config.json")
		config.CreateConfig()
		log.Fatalf("Modify the config.json file to customize how the tool functions: %v\n", err)
	}

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
	configFileExists := cf.FileExists("/" + config.SSLConfig)
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
	crtFileExists := cf.FileExists("/" + config.SSLCert)
	//keyFileExists := cf.FileExists("/" + keyFile)
	if !crtFileExists {
		cf.CreateCerts()
		//crtFileExists := cf.FileExists("/" + certFile)
		keyFileExists := cf.FileExists("/" + config.SSLKey)
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
	reverseProxy, err := createReverseProxy(config.ProxiedHTTPEndpoint)
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

	if config.BasicAuthOptions.Enabled {
		handler = authConnection(handler, config)
	}

	// Configure TLS
	tlsConfig := &tls.Config{
		MinVersion: tls.VersionTLS12, // Use modern TLS settings
	}

	httpServer := &http.Server{
		Addr:      config.ListeningPort,
		Handler:   handler,
		TLSConfig: tlsConfig,
	}

	// Start HTTPS server
	log.Printf("Starting reverse SSL proxy on %s, forwarding to %s", config.ListeningPort, config.ProxiedHTTPEndpoint)
	if err := httpServer.ListenAndServeTLS(config.SSLCert, config.SSLKey); err != nil {
		log.Fatalf("Failed to start HTTPS Listening Server for the Proxy: %v", err)
	}
}
