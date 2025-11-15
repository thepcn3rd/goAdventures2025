package tools

import (
	"bytes"
	"fmt"
	"log"
	"os/exec"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func LoadTools(s *mcp.Server) *mcp.Server {
	// Recon - nmap discovery scan
	mcp.AddTool(s, &mcp.Tool{
		Name:        "nmapRecon",
		Description: "Scan using nmap a specified IP address.",
	}, NmapRecon)

	log.Printf("Available tool: nmapRecon executes nmap against a specified IP Address (i.e. nmap 127.0.0.1)")

	// Recon - nmap port scan -sV
	mcp.AddTool(s, &mcp.Tool{
		Name:        "nmapScan",
		Description: "Scan using nmap scan a specified IP address and Port(s).",
	}, NmapScan)

	log.Printf("Available tool: nmapScan executes nmap against a specified IP Address and Port(s) (i.e. nmap 127.0.0.1 22,11434)")

	// Recon - CVE Search
	mcp.AddTool(s, &mcp.Tool{
		Name:        "cveSearch",
		Description: "Search for a CVE based on a keyword provided (i.e. wordpress) and timeframe to look back (i.e. 7 days).",
	}, CVESearch)

	log.Printf("Available tool: cveSearch searches the NIST NVD database for CVEs based on a keyword and timeframe to look back (i.e. wordpress 7 days)")

	return s
}

func ExecuteCommand(command string) (string, error) {
	// Create the command
	cmd := exec.Command("sh", "-c", command)

	// Capture stdout and stderr
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	log.Printf("Command Executing: \n%s", command)

	// Execute the command
	err := cmd.Run()
	if err != nil {
		return "", fmt.Errorf("%s execution failed: %v, stderr: %s", command, err, stderr.String())
	}

	// Wait for the command to finish and return the output
	//time.Sleep(3 * time.Minute) // Ensure command has time to complete - 3 minute timeout

	return stdout.String(), nil
}
