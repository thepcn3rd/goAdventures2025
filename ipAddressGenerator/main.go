package main

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net"
	"os"
)

// Config represents the structure of the config.json file
type Config struct {
	SourceSubnet      string   `json:"source_subnet"`
	DestinationSubnet string   `json:"destination_subnet"`
	DestinationPorts  []int    `json:"destination_ports"`
	Protocols         []string `json:"protocols"`
	Count             int      `json:"count"`
}

// FlowRecord represents a single network flow record
type FlowRecord struct {
	SourceIP      string `json:"source_ip"`
	DestinationIP string `json:"destination_ip"`
	Protocol      string `json:"protocol"`
	Port          int    `json:"port"`
}

func main() {
	// Read the config file
	config, err := readConfig("config.json")
	if err != nil {
		fmt.Printf("Error reading config: %v\n", err)
		os.Exit(1)
	}

	// Generate random flow records
	records := generateIPRecords(config)

	// Print the generated records
	for _, record := range records {
		fmt.Printf("%s,%s,%s,%d\n", record.SourceIP, record.DestinationIP, record.Protocol, record.Port)
	}
}

func readConfig(filename string) (*Config, error) {
	file, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var config Config
	err = json.Unmarshal(file, &config)
	if err != nil {
		return nil, err
	}

	return &config, nil
}

func generateIPRecords(config *Config) []FlowRecord {
	// Parse the source and destination subnets
	_, sourceNet, err := net.ParseCIDR(config.SourceSubnet)
	if err != nil {
		panic(fmt.Sprintf("Invalid source subnet: %v", err))
	}

	_, destNet, err := net.ParseCIDR(config.DestinationSubnet)
	if err != nil {
		panic(fmt.Sprintf("Invalid destination subnet: %v", err))
	}

	records := make([]FlowRecord, 0, config.Count)

	for i := 0; i < config.Count; i++ {
		// Generate random source and destination IPs within their subnets
		sourceIP := generateRandomIP(sourceNet)
		destIP := generateRandomIP(destNet)

		// Select random protocol and port
		protocol := config.Protocols[rand.Intn(len(config.Protocols))]
		port := config.DestinationPorts[rand.Intn(len(config.DestinationPorts))]

		records = append(records, FlowRecord{
			SourceIP:      sourceIP,
			DestinationIP: destIP,
			Protocol:      protocol,
			Port:          port,
		})
	}

	return records
}

func generateRandomIP(network *net.IPNet) string {
	// Generate a random IP address within the given network
	ip := make(net.IP, len(network.IP))
	copy(ip, network.IP)

	for i := 0; i < len(ip); i++ {
		if network.Mask[i] == 0 {
			ip[i] = byte(rand.Intn(256)) // Random byte for this octet
		} else {
			ip[i] = network.IP[i] & network.Mask[i] // Keep the network part
		}
	}

	return ip.String()
}
