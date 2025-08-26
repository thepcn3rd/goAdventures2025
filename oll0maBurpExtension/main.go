package main

import (
	"bytes"
	"context"
	"crypto/sha256"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
	"unicode"

	"github.com/gorilla/mux"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Wishlist of features to add
/**
1. Create a configuration file to store the settings (Completed)
2. Organize the functions better (Completed)
3. Create a round robin system to send requests to multiple Ollama instances
4. Sanitize inputs to prevent injection attacks on the API
5. Pull the settings for the extension from the server instead of hardcoding them in the extension
6. If request or response sent from the extension is smaller than 30 characters it crashes the queue server (Fixed)
7. Query the Ollama models available and populate a dropdown in the extension
8. Query the files available for system prompts and populate the text area
**/

type Configuration struct {
	QueueServerURL string `json:"queueServerURL"`
	OllamaURL      string `json:"ollamaURL"`
	APIKey         string `json:"apiKey"`
	TLSConfig      string `json:"tlsConfig"`
	TLSCert        string `json:"tlsCert"`
	TLSKey         string `json:"tlsKey"`
}

// Model represents the structure of a model from Ollama API
type Model struct {
	Name       string `json:"name"`
	ModifiedAt string `json:"modified_at"`
	Size       int64  `json:"size"`
	Digest     string `json:"digest"`
}

// TagsResponse represents the response structure from /api/tags
type TagsResponse struct {
	Models []Model `json:"models"`
}

// Request represents the incoming POST request
type Request struct {
	Model         string    `json:"model"`
	Request       string    `json:"request"`
	Response      string    `json:"response"`
	SystemPrompt  string    `json:"systemPrompt"`
	Timestamp     time.Time `json:"timestamp,omitempty"`
	ID            string    `json:"id,omitempty"`
	APIKey        string    `json:"apiKey,omitempty"`
	RequestNumber string    `json:"requestNumber,omitempty"`
}

type FileRequest struct {
	APIKey string `json:"apiKey,omitempty"`
}

type FileDownloadRequest struct {
	APIKey   string `json:"apiKey,omitempty"`
	FileName string `json:"fileName,omitempty"`
}

type OllamaRequest struct {
	Stream   bool             `json:"stream"`
	Messages []OllamaMessages `json:"messages"`
	Model    string           `json:"model"`
}

type OllamaMessages struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// Job represents a queued job
type Job struct {
	Request Request
	Result  chan string
}

// QueueServer manages the HTTP server and job queue
type QueueServer struct {
	router        *mux.Router
	server        *http.Server
	jobQueue      chan Job
	logger        *zap.Logger
	wg            sync.WaitGroup
	ctx           context.Context
	cancel        context.CancelFunc
	jobCounter    int
	mu            sync.Mutex
	apiKey        string
	requestNumber string
}

func (c *Configuration) CreateConfig(f string) error {
	c.QueueServerURL = "https://localhost:8443"
	c.APIKey = ""
	c.OllamaURL = "http://localhost:11434"
	c.TLSConfig = "keys/tlsConfig.json"
	c.TLSCert = "keys/tls.crt"
	c.TLSKey = "keys/tls.key"
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

// NewQueueServer creates a new queue server instance
func NewQueueServer(addr string, queueSize int, requestNumber string) (*QueueServer, error) {
	// Create context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())

	// Initialize logger
	logger, err := initLogger()
	if err != nil {
		cancel()
		return nil, fmt.Errorf("failed to initialize logger: %w", err)
	}

	qs := &QueueServer{
		router:        mux.NewRouter(),
		jobQueue:      make(chan Job, queueSize),
		logger:        logger,
		ctx:           ctx,
		cancel:        cancel,
		requestNumber: requestNumber,
	}

	// Configure routes
	qs.setupRoutes()

	// Configure HTTPS server
	qs.server = &http.Server{
		Addr:    addr,
		Handler: qs.router,
		TLSConfig: &tls.Config{
			MinVersion: tls.VersionTLS12,
		},
	}

	return qs, nil
}

// initLogger sets up structured logging with both console and file output
func initLogger() (*zap.Logger, error) {
	// Create logs directory if it doesn't exist
	if err := os.MkdirAll("logs", 0755); err != nil {
		return nil, err
	}

	// File encoder config
	fileEncoderConfig := zap.NewProductionEncoderConfig()
	fileEncoderConfig.TimeKey = "timestamp"
	fileEncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	// Console encoder config
	consoleEncoderConfig := zap.NewDevelopmentEncoderConfig()
	consoleEncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder

	// File writer
	timestamp := time.Now().Format("2006-01-02")
	fileName := fmt.Sprintf("logs/app_%s.log", timestamp)
	logFile, err := os.OpenFile(fileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, err
	}

	// Create cores
	fileCore := zapcore.NewCore(
		zapcore.NewJSONEncoder(fileEncoderConfig),
		zapcore.AddSync(logFile),
		zap.InfoLevel,
	)

	consoleCore := zapcore.NewCore(
		zapcore.NewConsoleEncoder(consoleEncoderConfig),
		zapcore.AddSync(os.Stdout),
		zap.InfoLevel,
	)

	// Combine cores
	core := zapcore.NewTee(fileCore, consoleCore)

	return zap.New(core), nil
}

// setupRoutes configures the HTTP routes
func (qs *QueueServer) setupRoutes() {
	qs.router.HandleFunc("/api/submit", qs.handlePostJob).Methods("POST")
	qs.router.HandleFunc("/api/health", qs.handleHealthCheck).Methods("GET")
	qs.router.HandleFunc("/api/stats", qs.handleStats).Methods("GET")
	qs.router.HandleFunc("/api/files", qs.handleFileList).Methods("POST")
	qs.router.HandleFunc("/api/file", qs.handleFileDownload).Methods("POST")
	qs.router.HandleFunc("/api/loadconfig", qs.handleLoadConfig).Methods("POST")
}

func (qs *QueueServer) handleLoadConfig(w http.ResponseWriter, r *http.Request) {
	var req FileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		qs.logger.Error("Failed to decode request", zap.Error(err))
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if req.APIKey != config.APIKey {
		qs.logger.Warn("Unauthorized access attempt (Invalid API Key)")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Fetch model names from Ollama server
	modelNames, err := GetModelNames()
	if err != nil {
		qs.logger.Error("Failed to get model names", zap.Error(err))
		http.Error(w, "Failed to get model names", http.StatusInternalServerError)
		return
	}

	fmt.Println(modelNames)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"ollamaURL": config.OllamaURL,
		"models":    modelNames,
	})
	qs.logger.Info("Configuration Retrieved")

}

// GetOllamaModels fetches the list of available models from Ollama server
func GetOllamaModels() ([]Model, error) {
	//url := fmt.Sprintf("http://%s:%d/api/tags", host, port)
	url := config.OllamaURL + "/api/tags"
	// Create HTTP client with timeout
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	// Create GET request
	resp, err := client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Ollama server: %w", err)
	}
	defer resp.Body.Close()

	// Check status code
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("server returned non-200 status: %s", resp.Status)
	}

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	//fmt.Println("Response from Ollama /api/tags endpoint:")
	//fmt.Println(string(body))

	// Parse JSON response
	var tagsResponse TagsResponse
	err = json.Unmarshal(body, &tagsResponse)
	if err != nil {
		return nil, fmt.Errorf("failed to parse JSON response: %w", err)
	}

	return tagsResponse.Models, nil
}

// GetModelNames returns just the model names as a string slice
func GetModelNames() ([]string, error) {
	models, err := GetOllamaModels()
	if err != nil {
		return nil, err
	}

	names := make([]string, len(models))
	for i, model := range models {
		names[i] = model.Name
	}

	return names, nil
}

func (qs *QueueServer) handleFileDownload(w http.ResponseWriter, r *http.Request) {
	var req FileDownloadRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		qs.logger.Error("Failed to decode request", zap.Error(err))
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	if req.APIKey != config.APIKey {
		qs.logger.Warn("Unauthorized access attempt (Invalid API Key)")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	filePath := "output/" + req.FileName
	file, err := os.Open(filePath)
	if err != nil {
		qs.logger.Error("Failed to open file", zap.String("file", req.FileName), zap.Error(err))
		http.Error(w, "File not found", http.StatusNotFound)
		return
	}
	defer file.Close()

	// Read the entire file content
	data, err := io.ReadAll(file)
	if err != nil {
		panic(err)
	}

	// Encode to base64 string
	encodedFile := base64.StdEncoding.EncodeToString(data)
	// Create a sha256 hash of the encoded filename
	hash := sha256.Sum256([]byte(encodedFile))

	// return the file in json format
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"encodedFile": encodedFile,
		"hash":        fmt.Sprintf("%x", hash),
	})
	qs.logger.Info("File list retrieved", zap.String("hash", fmt.Sprintf("%x", hash)))

}

// handleFileList returns a list of files in the output directory
func (qs *QueueServer) handleFileList(w http.ResponseWriter, r *http.Request) {
	var req FileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		qs.logger.Error("Failed to decode request", zap.Error(err))
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if req.APIKey != config.APIKey {
		qs.logger.Warn("Unauthorized access attempt (Invalid API Key)")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	files, err := os.ReadDir("output")
	if err != nil {
		qs.logger.Error("Failed to read output directory", zap.Error(err))
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	var fileNames []string
	for _, file := range files {
		if !file.IsDir() {
			fileNames = append(fileNames, file.Name())
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"files": fileNames,
	})
	qs.logger.Info("File list retrieved", zap.Int("file_count", len(fileNames)))
}

// handlePostJob processes incoming POST requests
func (qs *QueueServer) handlePostJob(w http.ResponseWriter, r *http.Request) {
	var req Request
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		qs.logger.Error("Failed to decode request", zap.Error(err))
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	//fmt.Println("Received API Key: " + req.APIKey)
	//fmt.Println("Expected API Key: " + config.APIKey)
	if req.APIKey != config.APIKey {
		qs.logger.Warn("Unauthorized access attempt (Invalid API Key)")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Generate ID for job
	req.ID = generateID()
	req.Timestamp = time.Now()

	// Create job with result channel
	resultChan := make(chan string, 1)
	job := Job{
		Request: req,
		Result:  resultChan,
	}

	// Try to add to queue with timeout
	select {
	case qs.jobQueue <- job:
		ollamaRequest, err := base64.StdEncoding.DecodeString(job.Request.Request)
		if err != nil {
			qs.logger.Error("Failed to decode request", zap.Error(err))
			job.Result <- fmt.Sprintf("Error decoding request: %v", err)
			return
		}
		ollamaRequestString := FilterAlphanumericWithSpaces(string(ollamaRequest))
		if len(ollamaRequestString) <= 30 {
			ollamaRequestString = ollamaRequestString + "_" + GenerateRandomString(30-len(ollamaRequestString))
		}
		qs.logger.Info("Job queued successfully",
			zap.String("job_id", req.ID),
			zap.Any("data", ollamaRequestString[:30]))
		// Wait for processing result with timeout
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusAccepted)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"id":        req.ID,
			"status":    "queued",
			"message":   "Job accepted",
			"timestamp": time.Now(),
		})

	case <-time.After(5 * time.Second):
		qs.logger.Error("Queue full", zap.String("job_id", req.ID))
		http.Error(w, "Queue full", http.StatusServiceUnavailable)
	}
}

// handleHealthCheck returns server status
func (qs *QueueServer) handleHealthCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":     "healthy",
		"timestamp":  time.Now(),
		"queue_size": len(qs.jobQueue),
		"capacity":   cap(qs.jobQueue),
	})
}

// handleStats returns queue statistics
func (qs *QueueServer) handleStats(w http.ResponseWriter, r *http.Request) {
	qs.mu.Lock()
	defer qs.mu.Unlock()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"total_jobs_processed": qs.jobCounter,
		"current_queue_size":   len(qs.jobQueue),
		"queue_capacity":       cap(qs.jobQueue),
		"timestamp":            time.Now(),
	})
}

// processJobs handles job processing in the background
func (qs *QueueServer) processJobs() {
	qs.wg.Add(1)
	defer qs.wg.Done()

	qs.logger.Info("Starting job processor")

	for {
		select {
		case job := <-qs.jobQueue:
			qs.processJob(job)
		case <-qs.ctx.Done():
			qs.logger.Info("Job processor shutting down")
			return
		}
	}
}

// processJob simulates job processing
func (qs *QueueServer) processJob(job Job) {
	var err error
	startTime := time.Now()

	if config.APIKey != job.Request.APIKey {
		qs.logger.Warn("Invalid API Key with Submission", zap.String("job_id", job.Request.ID))
		job.Result <- "Unauthorized"
		return
	}

	ollamaRequest, err := base64.StdEncoding.DecodeString(job.Request.Request)
	if err != nil {
		qs.logger.Error("Failed to decode request", zap.Error(err))
		job.Result <- fmt.Sprintf("Error decoding request: %v", err)
		return
	}

	ollamaRequestString := FilterAlphanumericWithSpaces(string(ollamaRequest))
	if len(ollamaRequestString) <= 30 {
		ollamaRequestString = ollamaRequestString + "_" + GenerateRandomString(30-len(ollamaRequestString))
	}

	qs.logger.Info("Processing job started",
		zap.String("job_id", job.Request.ID),
		zap.Any("data", ollamaRequestString[:30]))

	//fmt.Println(job.Request.Model)
	var oReq OllamaRequest
	oReq.Stream = false

	ollamaModel, err := base64.StdEncoding.DecodeString(job.Request.Model)
	if err != nil {
		qs.logger.Error("Failed to decode model", zap.Error(err))
		job.Result <- fmt.Sprintf("Error decoding model: %v", err)
		return
	}
	oReq.Model = string(ollamaModel)

	ollamaSystemPrompt, err := base64.StdEncoding.DecodeString(job.Request.SystemPrompt)
	if err != nil {
		qs.logger.Error("Failed to decode system prompt", zap.Error(err))
		job.Result <- fmt.Sprintf("Error decoding system prompt: %v", err)
		return
	}

	ollamaResponse, err := base64.StdEncoding.DecodeString(job.Request.Response)
	if err != nil {
		qs.logger.Error("Failed to decode response", zap.Error(err))
		job.Result <- fmt.Sprintf("Error decoding response: %v", err)
		return
	}

	var roleSystem OllamaMessages
	roleSystem.Role = "system"
	roleSystem.Content = string(ollamaSystemPrompt)

	var roleUser OllamaMessages
	roleUser.Role = "user"
	roleUser.Content = fmt.Sprintf("%s\n\n%s", string(ollamaRequest), string(ollamaResponse))

	oReq.Messages = append(oReq.Messages, roleSystem)
	oReq.Messages = append(oReq.Messages, roleUser)

	jsonData, err := json.Marshal(oReq)
	if err != nil {
		qs.logger.Error("Failed to marshal Ollama request", zap.Error(err))
		job.Result <- fmt.Sprintf("Error marshalling request: %v", err)
		return
	}

	response, err := sendToOllama([]byte(jsonData))
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	//fmt.Println("Response from Ollama:", string(response))
	// Create output directory if it doesn't exist
	if err := os.MkdirAll("output", 0755); err != nil {
		return
	}
	// Save the response to a file
	timestamp := time.Now().Format("2006-01-02_150405")
	outputFileName := fmt.Sprintf("output/results_req%s_%s_%s", job.Request.RequestNumber, timestamp, job.Request.ID)
	SaveOutputFile(string(response), outputFileName)

	// Send the jsonData to the Ollama API
	qs.logger.Info("Sending request to Ollama API",
		zap.String("job_id", job.Request.ID),
		zap.String("ollama_request", ollamaRequestString[:30]))

	// Simulate a delay for processing
	time.Sleep(1 * time.Second)

	// Update statistics
	qs.mu.Lock()
	qs.jobCounter++
	qs.mu.Unlock()

	processingDuration := time.Since(startTime)

	result := fmt.Sprintf("Processed in %v ID: %v Request: %s", processingDuration, job.Request.ID, ollamaRequestString[:30])

	qs.logger.Info("Job processing completed",
		zap.String("job_id", job.Request.ID),
		zap.Duration("processing_time", processingDuration),
		zap.String("result", result))

	// Send result back to HTTP handler
	select {
	case job.Result <- result:
	case <-time.After(1 * time.Second):
		qs.logger.Warn("Failed to send result - channel timeout",
			zap.String("job_id", job.Request.ID))
	}
}

// Start begins the HTTP server and job processing
func (qs *QueueServer) Start() error {
	// Start job processor
	go qs.processJobs()

	qs.logger.Info("Starting HTTPS server", zap.String("address", qs.server.Addr))

	// Start HTTPS server
	if err := qs.server.ListenAndServeTLS(config.TLSCert, config.TLSKey); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("failed to start server: %w", err)
	}

	return nil
}

// Stop gracefully shuts down the server
func (qs *QueueServer) Stop() error {
	qs.logger.Info("Initiating graceful shutdown")

	qs.cancel() // Signal job processor to stop

	// Shutdown HTTP server with timeout
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := qs.server.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("server shutdown failed: %w", err)
	}

	// Wait for job processor to finish
	qs.wg.Wait()
	qs.logger.Info("Server shutdown completed")

	return qs.logger.Sync()
}

// generateID creates a unique job ID
func generateID() string {
	return fmt.Sprintf("job_%d_%d", time.Now().Unix(), time.Now().UnixNano()%1000)
}

func FilterAlphanumericWithSpaces(input string) string {
	var result []rune

	for _, char := range input {
		if unicode.IsLetter(char) || unicode.IsDigit(char) || unicode.IsSpace(char) {
			result = append(result, char)
		}
	}

	return string(result)
}

func sendToOllama(jsonData []byte) ([]byte, error) {
	// Create HTTP request
	ollamaChatURL := config.OllamaURL + "/api/chat"
	req, err := http.NewRequest("POST", ollamaChatURL,
		bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{
		Timeout: 15 * time.Minute,
	}

	// Send request
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error sending request (waited 15 minutes): %w", err)
	}
	defer resp.Body.Close()

	// Check status code
	var body []byte
	if resp.StatusCode != http.StatusOK {
		body, _ = io.ReadAll(resp.Body)
		return nil, fmt.Errorf("HTTP error %d: %s", resp.StatusCode, string(body))
	} else {
		body, err = io.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("error reading response body: %w", err)
		}
	}

	return body, nil
}

var config Configuration

func main() {
	//config.OllamaURL = "http://10.27.20.160:11434/api/chat"
	//config.OllamaModel = "llama3.2:3b"

	ConfigPtr := flag.String("config", "config.json", "Configuration file to load for the proxy")
	flag.Parse()

	configFile := *ConfigPtr
	log.Println("Loading the following config file: " + configFile + "\n")
	if err := config.LoadConfig(configFile); err != nil {
		//fmt.Println("Could not load the configuration file, creating a new default config.json")
		config.CreateConfig(configFile)
		log.Fatalf("Modify the %s file to customize how the tool functions: %v\n", configFile, err)
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

	if len(config.APIKey) > 0 {
		fmt.Printf("Using Existing API Key from the config loaded: %s\n\n", config.APIKey)
	} else {
		fmt.Printf("No API Key declared in the config file, generating a one-time use Key\n")
		// Create Random API Key to be Used in the Burp Extension
		config.APIKey = GenerateRandomString(64)
		fmt.Printf("\nGenerated API Key: %s\n\n", config.APIKey)
	}

	// Create queue server with 100 job capacity
	serverIPPort := config.QueueServerURL
	serverIPPort = strings.TrimPrefix(serverIPPort, "https://")
	serverIPPort = strings.TrimPrefix(serverIPPort, "http://")
	server, err := NewQueueServer(serverIPPort, 100, "0") // The request number is 0 until populated
	if err != nil {
		log.Fatalf("Failed to create server: %v", err)
	}

	// Handle graceful shutdown
	go func() {
		sigChan := make(chan os.Signal, 1)
		<-sigChan

		if err := server.Stop(); err != nil {
			log.Printf("Error during shutdown: %v", err)
		}
		os.Exit(0)
	}()

	// Start the server
	if err := server.Start(); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
