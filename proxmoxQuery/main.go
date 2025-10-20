package main

import (
	"crypto/tls"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"math"
	"net/http"
	"os"
	"sort"
	"strings"

	cf "github.com/thepcn3rd/goAdvsCommonFunctions"
)

// Define the struct for the configuration
type ConfigStruct struct {
	TokenID     string `json:"tokenID"`
	TokenSecret string `json:"tokenSecret"`
	APIURL      string `json:"apiURL"`
}

// Define the struct to represent the JSON structure
type NodeData struct {
	Data []NodeStruct `json:"data"`
}

type NodeStruct struct {
	Status         string  `json:"status"`
	Level          string  `json:"level"`
	Disk           int64   `json:"disk"`
	ID             string  `json:"id"`
	CPU            float64 `json:"cpu"`
	Uptime         int64   `json:"uptime"`
	MaxDisk        int64   `json:"maxdisk"`
	MaxCPU         int     `json:"maxcpu"`
	Node           string  `json:"node"`
	MaxMem         int64   `json:"maxmem"`
	Mem            int64   `json:"mem"`
	Type           string  `json:"type"`
	SSLFingerprint string  `json:"ssl_fingerprint"`
}

type VMData struct {
	Data []VMStruct `json:"data"`
}

type VMStruct struct {
	MaxMem    int64          `json:"maxmem"`
	NetIn     int64          `json:"netin"`
	DiskRead  int64          `json:"diskread"`
	Name      string         `json:"name"`
	DiskWrite int64          `json:"diskwrite"`
	NetOut    int64          `json:"netout"`
	MaxDisk   int64          `json:"maxdisk"`
	VMID      int            `json:"vmid"`
	Uptime    int64          `json:"uptime"`
	Mem       int64          `json:"mem"`
	Status    string         `json:"status"`
	CPUs      int            `json:"cpus"`
	CPU       float64        `json:"cpu"`
	Disk      int64          `json:"disk"`
	PID       int            `json:"pid,omitempty"`
	VMConfig  VMConfigStruct `json:"vmConfig,omitempty"`
}

type VMConfigData struct {
	Data VMConfigStruct `json:"data"`
}

type VMConfigStruct struct {
	Smbios1 string `json:"smbios1"`
	Onboot  int    `json:"onboot"`
	Meta    string `json:"meta"`
	Machine string `json:"machine,omitempty"`
	Net0    string `json:"net0"`
	CPU     string `json:"cpu"`
	Sockets int    `json:"sockets"`
	Ide0    string `json:"ide0,omitempty"`
	Scsi0   string `json:"scsi0,omitempty"`
	Vmgenid string `json:"vmgenid"`
	Memory  int    `json:"memory"`
	Name    string `json:"name"`
	Digest  string `json:"digest"`
	Cores   int    `json:"cores"`
	Numa    int    `json:"numa"`
	Scsihw  string `json:"scsihw"`
	Ostype  string `json:"ostype"`
	Ide2    string `json:"ide2"`
	Boot    string `json:"boot"`
}

type VMStorageData struct {
	Data []StorageDataStruct `json:"data"`
}

type StorageDataStruct struct {
	Content      string  `json:"content"`
	Storage      string  `json:"storage"`
	Type         string  `json:"type"`
	Total        int     `json:"total,omitempty"`
	Enabled      int     `json:"enabled,omitempty"`
	Used         int     `json:"used,omitempty"`
	UsedFraction float64 `json:"used_fraction,omitempty"`
	Shared       int     `json:"shared,omitempty"`
	Avail        int     `json:"avail,omitempty"`
	Active       int     `json:"active,omitempty"`
}

var config ConfigStruct
var nodeInfo NodeData
var vmInfo VMData
var storageInfo VMStorageData

func gatherNodeName() {
	createURL := fmt.Sprintf("%s/api2/json/nodes", config.APIURL) // Needed to gather the node name of proxymox...
	nodeResponse := webRequest(createURL, "GET")
	err := json.Unmarshal(nodeResponse, &nodeInfo)
	cf.CheckError("Unable to unmarshall the json configuration", err, true)
}

func gatherVMList() {
	createURL := fmt.Sprintf("%s/api2/json/nodes/%s/qemu", config.APIURL, nodeInfo.Data[0].Node)
	vmResponse := webRequest(createURL, "GET")
	err := json.Unmarshal(vmResponse, &vmInfo)
	cf.CheckError("Unable to unmarshall the json configuration for the VM List", err, true)
}

func gatherVMConfig() {
	for i, vm := range vmInfo.Data {
		var vmData VMConfigData
		createURL := fmt.Sprintf("%s/api2/json/nodes/%s/qemu/%d/config", config.APIURL, nodeInfo.Data[0].Node, vm.VMID)
		vmConfigResponse := webRequest(createURL, "GET")
		err := json.Unmarshal(vmConfigResponse, &vmData)
		cf.CheckError("Unable to unmarshall the VM Configuration", err, true)
		vmInfo.Data[i].VMConfig = vmData.Data
	}
}

func gatherVMStorage() {
	//var vmStorageData VMStorageData
	createURL := fmt.Sprintf("%s/api2/json/nodes/%s/storage", config.APIURL, nodeInfo.Data[0].Node)
	vmStorageResponse := webRequest(createURL, "GET")
	//fmt.Println(string(vmStorageResponse))
	err := json.Unmarshal(vmStorageResponse, &storageInfo)
	cf.CheckError("Unable to unmarshall the VM Configuration", err, true)
	//vmInfo.Data[i].VMConfig = vmData.Data
	//storageInfo.Data = vmStorageData.Data
}

func webRequest(url string, httpVerb string) []byte {
	req, err := http.NewRequest(httpVerb, url, nil)
	cf.CheckError("Unable to create the GET Request", err, true)
	token := fmt.Sprintf("PVEAPIToken=%s=%s", config.TokenID, config.TokenSecret)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", token)

	// Create a Custom HTTP client that skips TLS Certificate Validation
	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	resp, err := client.Do(req)
	cf.CheckError("Unable to GET a Response", err, true)
	defer resp.Body.Close()
	respBody, _ := io.ReadAll(resp.Body)
	return respBody
}

/**
func createVM(apiURL, token string, config VMConfig) error {
	createURL := fmt.Sprintf("%s/api2/json/nodes/%s/qemu", apiURL, config.Node)

	// Build the VM creation request
	body := map[string]interface{}{
		"vmid":    config.VMID,
		"name":    config.Name,
		"memory":  config.Memory,
		"sockets": config.Sockets,
		"cores":   config.Cores,
		"ide2":    config.ISO + ",media=cdrom",
		"scsihw":  "virtio-scsi-pci",
		"scsi0":   fmt.Sprintf("dataB:%s,format=qcow2", config.DiskSize),
		"net0":    fmt.Sprintf("virtio,bridge=%s", config.Bridge),
		"boot":    "cdrom",
	}
	bodyJSON, _ := json.Marshal(body)

	// Create HTTP request
	req, err := http.NewRequest("POST", createURL, bytes.NewBuffer(bodyJSON))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "PVEAPIToken="+token)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to create VM, status code: %d", resp.StatusCode)
	}

	return nil
}
**/

func main() {
	ConfigPtr := flag.String("config", "config.json", "Configuration file to load for the proxy")
	flag.Parse()

	// ANSI Escape colors do not work on Windows compiled binaries...
	red := "\033[31m"
	green := "\033[32m"
	yellow := "\033[33m"
	reset := "\033[0m"

	// Load config.json file
	fmt.Printf("\n%sLoading the following config file: %s%s\n", green, *ConfigPtr, reset)
	//go logToSyslog(fmt.Sprintf("Loading the following config file: %s\n", *ConfigPtr))
	configFile, err := os.Open(*ConfigPtr)
	cf.CheckError("Unable to open the configuration file", err, true)
	defer configFile.Close()
	decoder := json.NewDecoder(configFile)
	if err := decoder.Decode(&config); err != nil {
		cf.CheckError("Unable to decode the configuration file", err, true)
	}

	// Populate the Node Data
	gatherNodeName()
	fmt.Printf("\n%sNode Name: %s%s\n", green, nodeInfo.Data[0].Node, reset)
	maxCPU := nodeInfo.Data[0].MaxCPU
	var totalUsedCPU int
	fmt.Printf("Max CPUs: %d\n", maxCPU)
	maxMemory := int64(math.Round(float64(nodeInfo.Data[0].MaxMem) / (1024 * 1024 * 1024)))
	var totalUsedMemory int
	fmt.Printf("Max Memory: %dGB\n", maxMemory)
	fmt.Printf("Host Disk Space: %dGB\n\n", int64(math.Round(float64(nodeInfo.Data[0].MaxDisk)/(1024*1024*1024))))
	// Build out the node information that is displayed...

	// Populate the VM Data
	gatherVMList()

	// Sort the VM Data by VMID
	sort.Slice(vmInfo.Data, func(i, j int) bool {
		return vmInfo.Data[i].VMID < vmInfo.Data[j].VMID
	})

	// Gather the VMID Configuration
	gatherVMConfig()

	// Display Results
	// Identify the Next VMID that can be used for provisioning
	nextVMID := 0
	largestVMID := 0
	fmt.Printf("%sVM List%s\n", green, reset)
	fmt.Printf("%-5s | %-30s | %-5s | %-6s | %-30s\n", "VMID", "VMName", "CPU", "Memory", "Disk")
	fmt.Println(strings.Repeat("-", 105))
	for _, v := range vmInfo.Data {
		// Display the disk that is present...
		var disk string
		if len(v.VMConfig.Ide0) > 0 {
			disk = v.VMConfig.Ide0
		} else {
			disk = v.VMConfig.Scsi0
		}
		fmt.Printf("%-5d | %-30s | %-5d | %-6d | %-30s\n", v.VMID, v.Name, v.CPUs, int64(math.Round(float64(v.MaxMem)/(1024*1024*1024))), disk)
		totalUsedCPU += v.CPUs
		totalUsedMemory += int(math.Round(float64(v.MaxMem) / (1024 * 1024 * 1024)))
		if v.VMID > largestVMID {
			largestVMID = v.VMID
		}
	}
	if totalUsedCPU >= maxCPU {
		fmt.Printf("%s*Total CPUs Provisioned %d, is greater than the maximum number of CPUs Available %d%s\n", yellow, totalUsedCPU, maxCPU, reset)
	}

	if totalUsedMemory >= int(maxMemory) {
		fmt.Printf("%s*Total Memory Provisioned %d, is greater than the maximum amount of Memory Available %d%s\n", red, totalUsedMemory, maxMemory, reset)
	}

	// The Next VMID
	nextVMID = largestVMID + 1
	fmt.Printf("\nThe next VMID is: %d\n", nextVMID)

	// List the total of CPU and Memory...
	// Compare to the node and show if over-provisioned...

	fmt.Printf("\n%sVMStorage Available%s\n", green, reset)
	gatherVMStorage()
	fmt.Printf("%-12s | %-30s | %-5s\n", "Name", "Type", "Available")
	fmt.Println(strings.Repeat("-", 105))
	for _, storage := range storageInfo.Data {
		fmt.Printf("%-12s | %-30s | %-5dGB\n", storage.Storage, storage.Type, int64(math.Round(float64(storage.Avail)/(1024*1024*1024))))
	}

	// Create VM from Options
	// Show ISO files to select from...

	/**
	// Create the VM
	err = createVM("https://10.27.20.75:8006", token, config)
	if err != nil {
		log.Fatalf("Failed to create VM: %v", err)
	}

	fmt.Println("VM created successfully!")
	**/
}
