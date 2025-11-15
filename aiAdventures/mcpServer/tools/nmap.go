package tools

import (
	"context"
	"fmt"
	"log"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type NmapReconParams struct {
	DestinationIP string `json:"destinationIP" jsonschema:"Destination IP Address to Get Nmap Scan Results For"`
}

func NmapRecon(ctx context.Context, req *mcp.CallToolRequest, params *NmapReconParams) (*mcp.CallToolResult, any, error) {
	command := fmt.Sprintf("nmap -Pn -p- --min-rate=1000 -T4 %s", params.DestinationIP)

	output, err := ExecuteCommand(command)
	if err != nil {
		return nil, "", err
	}

	log.Printf("Output from tool:\n%s", output)

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: output},
		},
	}, nil, nil
}

type NmapScanParams struct {
	DestinationIP    string `json:"destinationIP" jsonschema:"Destination IP Address to conduct a scan and gather information about open ports and services"`
	DestinationPorts string `json:"destinationPorts" jsonschema:"A comma seperated list of ports to include in the scan"`
}

func NmapScan(ctx context.Context, req *mcp.CallToolRequest, params *NmapScanParams) (*mcp.CallToolResult, any, error) {
	command := fmt.Sprintf("nmap -Pn -p%s -sV %s", params.DestinationPorts, params.DestinationIP)

	output, err := ExecuteCommand(command)
	if err != nil {
		return nil, "", err
	}

	log.Printf("Output from tool:\n%s", output)

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: output},
		},
	}, nil, nil
}
