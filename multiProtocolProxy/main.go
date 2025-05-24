package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"time"
)

type Configuration struct {
	ListeningDomain   string                   `json:"listeningDomain"`
	ListeningHTTPPort string                   `json:"listeningHTTPPort"`
	ListeningTLSPort  string                   `json:"listeningTLSPort"`
	ListeningTCPPort  string                   `json:"listeningTCPPort"`
	ProxyInformation  []ProxyInformationStruct `json:"proxyInformation"`
	BasicAuthOptions  BasicAuthOptions         `json:"basicAuthOptions"`
	TLSConfig         string                   `json:"tlsConfig"`
	TLSCert           string                   `json:"tlsCert"`
	TLSKey            string                   `json:"tlsKey"`
	LoggingOptions    LoggingOptionsStruct     `json:"loggingOptions"`
}

type ProxyInformationStruct struct {
	ProxyType      string `json:"proxyType"`
	ProxySubDomain string `json:"proxySubDomain"`
	ProxyEndpoint  string `json:"proxyEndpoint"`
	ProxyNotes     string `json:"proxyNotes"`
}

type BasicAuthOptions struct {
	Enabled  bool   `json:"enabled"`
	Username string `json:"username"`
	Password string `json:"password"`
}

type LoggingOptionsStruct struct {
	SyslogOptions   SyslogOptionsStruct   `json:"syslogOptions"`
	SaveFileOptions SaveFileOptionsStruct `json:"saveFileOptions"`
}

type SyslogOptionsStruct struct {
	SyslogEnabled    bool   `json:"syslogEnabled"`
	SyslogServer     string `json:"syslogServer"`
	SyslogOriginName string `json:"syslogOriginName"`
}

type SaveFileOptionsStruct struct {
	SaveFileEnabled  bool   `json:"saveFileEnabled"`
	SyslogServer     string `json:"saveFileBaseName"`
	SyslogOriginName string `json:"saveFileExtension"`
}

func (c *Configuration) CreateConfig(f string) error {
	c.ListeningDomain = "4gr8.local"
	c.ListeningHTTPPort = "8080"
	c.ListeningTLSPort = "8443"
	c.ListeningTCPPort = "9000"
	c.ProxyInformation = []ProxyInformationStruct{
		{
			ProxyType:      "https",
			ProxySubDomain: "api",
			ProxyEndpoint:  "http://localhost:8000",
			ProxyNotes:     "This is a test proxy",
		},
		{
			ProxyType:      "http",
			ProxySubDomain: "app",
			ProxyEndpoint:  "http://localhost:8000",
			ProxyNotes:     "This is a test proxy",
		},
		// Currently only the last tcp proxy in the config will work because it is a 1 to 1 mapping
		{
			ProxyType:      "tcp",
			ProxySubDomain: "tcp",
			ProxyEndpoint:  "localhost:18000",
			ProxyNotes:     "This is a test proxy",
		},
	}
	c.BasicAuthOptions = BasicAuthOptions{
		Enabled:  false,
		Username: "thepcn3rd",
		Password: "T3sting",
	}
	c.TLSConfig = "keys/tlsconfig.json"
	c.TLSCert = "keys/tls.crt"
	c.TLSKey = "keys/tls.key"
	c.LoggingOptions = LoggingOptionsStruct{
		SyslogOptions: SyslogOptionsStruct{
			SyslogEnabled:    false,
			SyslogServer:     "localhost",
			SyslogOriginName: "example.com",
		},
		SaveFileOptions: SaveFileOptionsStruct{
			SaveFileEnabled:  false,
			SyslogServer:     "localhost",
			SyslogOriginName: "example.com",
		},
	}

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

// SubdomainProxy holds the configuration for subdomain routing
type SubdomainProxy struct {
	// Map of subdomains to their target URLs
	Routes map[string]string
}

// NewSubdomainProxy creates a new SubdomainProxy instance
func NewSubdomainProxy(routes map[string]string) *SubdomainProxy {
	return &SubdomainProxy{Routes: routes}
}

// createReverseProxy creates a reverse proxy that respects subdomains
func createReverseProxy(target string) (*httputil.ReverseProxy, error) {
	targetURL, err := url.Parse(target)
	if err != nil {
		return nil, fmt.Errorf("unable to parse target URL: %w", err)
	}

	proxy := httputil.NewSingleHostReverseProxy(targetURL)

	// Modify the request to preserve the Host header (important for subdomains)
	originalDirector := proxy.Director
	proxy.Director = func(req *http.Request) {
		originalDirector(req)
		// Preserve the original host header
		req.Host = targetURL.Host
	}

	// Custom transport can be configured here if needed
	proxy.Transport = &http.Transport{
		// Your transport configuration
	}

	return proxy, nil
}

// ServeHTTP implements http.Handler interface
func (s *SubdomainProxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	//fmt.Printf("Remote Host: %s\n", r.Host)
	hostParts := strings.Split(r.Host, ".")
	if len(hostParts) < 2 {
		http.Error(w, "Invalid host", http.StatusBadRequest)
		return
	}

	// The subdomain is the first part before the main domain
	subdomain := hostParts[0]
	//fmt.Printf("Subdomain: %s\n", subdomain)
	// Look up the target for this subdomain
	target, exists := s.Routes[subdomain]
	if !exists {
		http.NotFound(w, r)
		return
	}

	// Create or reuse a reverse proxy for this subdomain
	//fmt.Printf("Target: %s\n", target)
	proxy, err := createReverseProxy(target)
	if err != nil {
		http.Error(w, "Error creating reverse proxy", http.StatusInternalServerError)
		return
	}

	// Serve the request
	proxy.ServeHTTP(w, r)
}

type TCPProxy struct {
	ListenAddr  string
	BackendAddr string
	Timeout     time.Duration
}

func (p *TCPProxy) Start() error {
	listener, err := net.Listen("tcp", p.ListenAddr)
	if err != nil {
		return err
	}
	defer listener.Close()

	//log.Printf("TCP proxy %s -> %s", p.ListenAddr, p.BackendAddr)

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-stop
		log.Println("Shutting down proxy...")
		listener.Close()
	}()

	for {
		clientConn, err := listener.Accept()
		if err != nil {
			log.Printf("Error in accepting connection sleeping for 100ms: %v", err)
			time.Sleep(100 * time.Millisecond)
			return err
		}

		go p.handleClient(clientConn)
	}
}

func (p *TCPProxy) handleClient(clientConn net.Conn) {
	defer clientConn.Close()

	// Set timeout if configured
	if p.Timeout > 0 {
		clientConn.SetDeadline(time.Now().Add(p.Timeout))
	}

	backendConn, err := net.DialTimeout("tcp", p.BackendAddr, p.Timeout)
	if err != nil {
		log.Printf("Backend connection failed: %v", err)
		return
	}
	defer backendConn.Close()

	if p.Timeout > 0 {
		backendConn.SetDeadline(time.Now().Add(p.Timeout))
	}

	var wg sync.WaitGroup
	wg.Add(6)

	go func() {
		defer wg.Done()
		_, err := io.Copy(backendConn, clientConn)
		//log.Printf("Copied %d bytes from client to backend", n)
		if err != nil {
			log.Printf("Client->Backend error: %v", err)
		}
		backendConn.(*net.TCPConn).CloseWrite()
	}()

	go func() {
		defer wg.Done()
		_, err := io.Copy(clientConn, backendConn)
		//log.Printf("Copied %d bytes from backend to client", n)
		if err != nil {
			log.Printf("Backend->Client error: %v", err)
		}
		clientConn.(*net.TCPConn).CloseWrite()
	}()

	wg.Wait()
}

func main() {
	ConfigPtr := flag.String("config", "config.json", "Configuration file to load for the proxy")
	flag.Parse()

	// Load the Configuration file
	var config Configuration
	log.Println("Loading the following config file: " + *ConfigPtr + "\n")
	if err := config.LoadConfig(*ConfigPtr); err != nil {
		//fmt.Println("Could not load the configuration file, creating a new default config.json")
		config.CreateConfig(*ConfigPtr)
		log.Fatalf("Modify the %s file to customize how the tool functions: %v\n", *ConfigPtr, err)
	}

	// Verify the TLS Certificate and Key files exist for the https server
	// Create the location of the keys folder
	dirPathTLS := filepath.Dir(config.TLSConfig)
	err := os.MkdirAll(dirPathTLS, os.ModePerm)
	if err != nil {
		log.Fatalf("Failed to create directories for TLS: %v", err)
	}
	// Does the tlsConfig.json  file exist in the keys folder
	TLSConfigFileExists := FileExists("/" + config.TLSConfig)
	//fmt.Println(configFileExists)
	if !TLSConfigFileExists {
		CreateCertConfigFile(config.TLSConfig)
		log.Fatalf("Created %s, modify the values to create the self-signed cert utilized", config.TLSConfig)
	}

	// Does the server.crt and server.key files exist in the keys folder
	crtFileExists := FileExists("/" + config.TLSCert)
	keyFileExists := FileExists("/" + config.TLSKey)
	if !crtFileExists || !keyFileExists {
		CreateCerts(config.TLSConfig, config.TLSCert, config.TLSKey)
		crtFileExists := FileExists("/" + config.TLSCert)
		keyFileExists := FileExists("/" + config.TLSKey)
		if !crtFileExists || !keyFileExists {
			fmt.Printf("Failed to create %s and %s files\n", config.TLSCert, config.TLSKey)
			os.Exit(0)
		}
	}

	tcpRoutes := TCPProxy{}
	httpRoutes := make(map[string]string)
	httpsRoutes := make(map[string]string)
	for _, proxyRoute := range config.ProxyInformation {
		// iterate through the proxy information and add the routes to the map
		if proxyRoute.ProxyType == "https" {
			httpsRoutes[proxyRoute.ProxySubDomain] = proxyRoute.ProxyEndpoint
		} else if proxyRoute.ProxyType == "http" {
			httpRoutes[proxyRoute.ProxySubDomain] = proxyRoute.ProxyEndpoint
		} else if proxyRoute.ProxyType == "tcp" {
			tcpRoutes.ListenAddr = proxyRoute.ProxySubDomain + "." + config.ListeningDomain + ":" + config.ListeningTCPPort
			tcpRoutes.BackendAddr = proxyRoute.ProxyEndpoint
			tcpRoutes.Timeout = 600 * time.Second // Timeout after 10 minutes
		} else {
			log.Printf("Unknown proxy type in config.json: %s\n", proxyRoute.ProxyType)
		}
	}

	//fmt.Println("HTTPS Routes:")
	//for subdomain, endpoint := range httpsRoutes {
	//	fmt.Printf("Subdomain: %s - Endpoint: %s\n", subdomain, endpoint)
	//}
	if len(httpsRoutes) > 0 {
		go func() {
			httpsProxy := NewSubdomainProxy(httpsRoutes)

			// Start the server
			listenAddressTLS := fmt.Sprintf("%s:%s", config.ListeningDomain, config.ListeningTLSPort)
			fmt.Printf("Starting HTTPS Proxy on %s\n", listenAddressTLS)
			err = http.ListenAndServeTLS(listenAddressTLS, config.TLSCert, config.TLSKey, httpsProxy)
			if err != nil {
				log.Fatalf("Failed to start server: %v", err)
			}

		}()
	}

	if len(httpRoutes) > 0 {
		go func() {
			httpProxy := NewSubdomainProxy(httpRoutes)

			// Start the server
			listenAddressHTTP := fmt.Sprintf("%s:%s", config.ListeningDomain, config.ListeningHTTPPort)
			fmt.Printf("Starting HTTP Proxy on %s\n", listenAddressHTTP)
			err = http.ListenAndServe(listenAddressHTTP, httpProxy)
			if err != nil {
				log.Fatalf("Failed to start server: %v", err)
			}
		}()
	}

	if len(tcpRoutes.BackendAddr) > 0 {
		go func() {
			fmt.Printf("Starting TCP Proxy on %s\n", tcpRoutes.ListenAddr)
			err := tcpRoutes.Start()
			if err != nil {
				log.Fatalf("Failed to start TCP proxy: %v", err)
			}
		}()
	}

	time.Sleep(2 * time.Second) // Wait for servers to start
	_ = PressEnterKeytoContinue("Press Enter to Stop the Proxies...")

}
