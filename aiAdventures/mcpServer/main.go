package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"mcpServer/tools"
)

type Configuration struct {
	MCPServerIP   string `json:"mcpServerIP"`
	MCPServerPort int    `json:"mcpServerPort"`
	APIKey        string `json:"apiKey"`
}

func (c *Configuration) CreateConfig(f string) error {
	c.MCPServerIP = "0.0.0.0"
	c.MCPServerPort = 8080
	c.APIKey = ""
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

// responseWriter wraps http.ResponseWriter to capture the status code.
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func loggingHandler(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Create a response writer wrapper to capture status code.
		wrapped := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

		// Log request details.
		log.Printf("[REQ] %s | %s | %s %s",
			start.Format(time.RFC3339),
			r.RemoteAddr,
			r.Method,
			r.URL.Path)

		// Call the actual handler.
		handler.ServeHTTP(wrapped, r)

		// Log response details.
		duration := time.Since(start)
		log.Printf("[RES] %s | %s | %s %s | Status: %d | Duration: %v",
			time.Now().Format(time.RFC3339),
			r.RemoteAddr,
			r.Method,
			r.URL.Path,
			wrapped.statusCode,
			duration)
	})
}

// API Key Authentication Middleware
func apiKeyAuth(handler http.Handler, expectedAPIKey string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Extract API key from header
		apiKey := r.Header.Get("APIKey")

		// Validate API key
		if apiKey == "" {
			http.Error(w, "API key required", http.StatusUnauthorized)
			log.Printf("API key validation failed: no API key provided")
			return
		}

		if apiKey != expectedAPIKey {
			http.Error(w, "Invalid API key", http.StatusUnauthorized)
			log.Printf("API key validation failed: invalid API key from %s", r.RemoteAddr)
			return
		}

		// API key is valid, proceed to next handler
		handler.ServeHTTP(w, r)
	})
}

func runServer(url string, expectedAPIKey string) {
	// Create an MCP server.
	server := mcp.NewServer(&mcp.Implementation{
		Name:    "mcp-server",
		Version: "1.0.0",
	}, nil)

	server = tools.LoadTools(server)

	// Create the streamable HTTP handler.
	handler := mcp.NewStreamableHTTPHandler(func(req *http.Request) *mcp.Server {
		log.Printf("Connecting Host: %s IP: %s Method: %s URL: %s Referer: %s UserAgent: %s", req.Host, req.RemoteAddr, req.Method, req.URL.String(), req.Referer(), req.UserAgent())
		return server
	}, nil)

	handlerWithAuth := apiKeyAuth(handler, expectedAPIKey)
	handlerWithLogging := loggingHandler(handlerWithAuth)

	log.Printf("MCP server listening on %s", url)

	// Start the HTTP server with logging handler.
	if err := http.ListenAndServe(url, handlerWithLogging); err != nil {
		log.Fatalf("Server failed: %v", err)
	}

}

var config Configuration

func main() {
	ConfigPtr := flag.String("config", "config.json", "Configuration file to load for the proxy")
	flag.Parse()

	configFile := *ConfigPtr
	log.Println("Loading the following config file: " + configFile + "\n")
	if err := config.LoadConfig(configFile); err != nil {
		//fmt.Println("Could not load the configuration file, creating a new default config.json")
		config.CreateConfig(configFile)
		log.Fatalf("Modify the %s file to customize how the tool functions: %v\n", configFile, err)
	}

	if config.MCPServerIP != "" && config.MCPServerPort != 0 && config.APIKey != "" {
		log.Printf("Starting MCP Server on %s:%d with API Key: %s\n", config.MCPServerIP, config.MCPServerPort, config.APIKey)
		runServer(fmt.Sprintf("%s:%d", config.MCPServerIP, config.MCPServerPort), config.APIKey)
		return
	} else if config.MCPServerIP != "" && config.MCPServerPort != 0 && config.APIKey == "" {
		log.Printf("Starting MCP Server on %s:%d with NO API Key\n", config.MCPServerIP, config.MCPServerPort)
		runServer(fmt.Sprintf("%s:%d", config.MCPServerIP, config.MCPServerPort), "")
		return
	} else {
		log.Printf("Starting MCP Server on 127.0.0.1:8080 with NO API Key\n")
		runServer(fmt.Sprintf("%s:%d", "127.0.0.1", 8080), "")
		return
	}

}
