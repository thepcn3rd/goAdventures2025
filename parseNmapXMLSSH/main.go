package main

import (
	"encoding/csv"
	"encoding/xml"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strconv"
)

/**

Built to parse XML for the nmap script ssh-enum-algos for multiple hosts and create a simple csv file with the results

**/

type NmapRun struct {
	XMLName  xml.Name `xml:"nmaprun"`
	Scanner  string   `xml:"scanner,attr"`
	Args     string   `xml:"args,attr"`
	Start    int64    `xml:"start,attr"`
	StartStr string   `xml:"startstr,attr"`
	Version  string   `xml:"version,attr"`
	ScanInfo ScanInfo `xml:"scaninfo"`
	Host     []Host   `xml:"host"`
}

type ScanInfo struct {
	Type        string `xml:"type,attr"`
	Protocol    string `xml:"protocol,attr"`
	NumServices int    `xml:"numservices,attr"`
	Services    string `xml:"services,attr"`
}

type Host struct {
	StartTime int64      `xml:"starttime,attr"`
	EndTime   int64      `xml:"endtime,attr"`
	Addresses []Address  `xml:"address"`
	Hostnames []Hostname `xml:"hostnames>hostname"`
	Ports     []Port     `xml:"ports>port"`
}

type Address struct {
	Addr     string `xml:"addr,attr"`
	AddrType string `xml:"addrtype,attr"`
}

type Hostname struct {
	Name string `xml:"name,attr"`
	Type string `xml:"type,attr"`
}

type Port struct {
	Protocol string   `xml:"protocol,attr"`
	PortID   int      `xml:"portid,attr"`
	State    State    `xml:"state"`
	Service  Service  `xml:"service"`
	Scripts  []Script `xml:"script"`
}

type State struct {
	State  string `xml:"state,attr"`
	Reason string `xml:"reason,attr"`
	TTL    int    `xml:"reason_ttl,attr"`
}

type Service struct {
	Name    string `xml:"name,attr"`
	Product string `xml:"product,attr"`
	Version string `xml:"version,attr"`
	Tunnel  string `xml:"tunnel,attr"`
	Method  string `xml:"method,attr"`
	Conf    int    `xml:"conf,attr"`
	CPEs    []CPE  `xml:"cpe"`
}

type CPE struct {
	Text string `xml:",chardata"`
}

type Script struct {
	ID     string  `xml:"id,attr"`
	Output string  `xml:"output,attr"`
	Tables []Table `xml:"table"`
	Elems  []Elem  `xml:"elem"`
}

type Table struct {
	Key    string  `xml:"key,attr"`
	Tables []Table `xml:"table"`
	Elems  []Elem  `xml:"elem"`
}

type Elem struct {
	Key   string `xml:"key,attr"`
	Value string `xml:",chardata"`
}

type SSHInformation struct {
	SSHInfo []SSHProtocolInformation
}

type SSHProtocolInformation struct {
	IP                 string
	Hostname           string
	Port               string
	Protocol           string
	ServiceName        string
	ServiceProduct     string
	ServiceVersion     string
	KexAlgorithms      []string
	ServerHostKeyAlgos []string
	EncryptionAlgos    []string
	MACAlgorithms      []string
	CompressionAlgos   []string
}

func valueExists(i []string, e Elem) []string {
	exists := false
	for _, u := range i {
		if u == e.Value {
			exists = true
		}
	}
	if !exists {
		i = append(i, e.Value)
	}

	return i
}

func AddHeaderValue(i []string, header []string, columnHeader string) []string {
	header = append(header, columnHeader)
	header = append(header, i...)
	return header
}

func appendCSVLine(line []string, unique []string, hostInfo []string) []string {
	countItem := len(hostInfo)
	line = append(line, strconv.Itoa(countItem))
	for _, u := range unique {
		exists := false
		for _, h := range hostInfo {
			if u == h {
				exists = true
			}
		}
		if exists {
			line = append(line, "x")
		} else {
			line = append(line, "-")
		}
	}

	return line
}

func main() {
	var xmlFiles []string
	var sshInfo SSHInformation
	var UniqueKexAlgorithms []string
	var UniqueServerHostKeyAlgos []string
	var UniqueEncryptionAlgos []string
	var UniqueMACAlgorithms []string
	var UniqueCompressionAlgos []string

	fmt.Println("Parsing XML and Parsing SSH Algorithm Information")

	err := filepath.Walk("output/", func(path string, info os.FileInfo, err error) error {

		if err != nil {
			return fmt.Errorf("[E] error accessing path %q: %w", path, err)
		}
		if info.IsDir() {
			return nil
		}
		xmlFiles = append(xmlFiles, path)
		return nil
	})
	// Handle any errors from filepath.Walk
	if err != nil {
		log.Printf("[W] Warning walking the directory: %v", err)
	}

	// Gather information from SSH Files
	for _, f := range xmlFiles {
		// Read XML file
		fmt.Printf("Processing File: %s\n", f)
		xmlFile, err := os.ReadFile(f)
		if err != nil {
			fmt.Printf("Error reading XML file: %v\n", err)
			return
		}

		// Parse XML
		var nmapRun NmapRun
		err = xml.Unmarshal(xmlFile, &nmapRun)
		if err != nil {
			fmt.Printf("Error parsing XML: %v\n", err)
			return
		}
		/**
		for i, a := range nmapRun.Host.Addresses {
			fmt.Printf("Count: %d Address: %s\n", i, a.Addr)
		}
		**/
		for _, h := range nmapRun.Host {
			for _, p := range h.Ports {
				var protoInfo SSHProtocolInformation
				protoInfo.IP = h.Addresses[0].Addr
				//fmt.Println(protoInfo.IP)
				if len(h.Hostnames) > 0 {
					protoInfo.Hostname = h.Hostnames[0].Name
				}
				for _, s := range p.Scripts {
					if s.ID == "ssh2-enum-algos" {
						protoInfo.Port = strconv.Itoa(p.PortID)
						protoInfo.Protocol = p.Protocol
						protoInfo.ServiceName = p.Service.Name
						protoInfo.ServiceProduct = p.Service.Product
						protoInfo.ServiceVersion = p.Service.Version
						for _, t := range s.Tables {
							//fmt.Println(t.Key)
							switch t.Key {
							case "kex_algorithms":
								for _, e := range t.Elems {
									protoInfo.KexAlgorithms = append(protoInfo.KexAlgorithms, e.Value)
									UniqueKexAlgorithms = valueExists(UniqueKexAlgorithms, e)
									/**
									exists := false
									for _, u := range UniqueKexAlgorithms {
										if u == e.Value {
											exists = true
										}
									}
									if !exists {
										UniqueKexAlgorithms = append(UniqueKexAlgorithms, e.Value)
									}
									**/
								}
							case "mac_algorithms":
								for _, e := range t.Elems {
									protoInfo.MACAlgorithms = append(protoInfo.MACAlgorithms, e.Value)
									UniqueMACAlgorithms = valueExists(UniqueMACAlgorithms, e)
								}
							case "encryption_algorithms":
								for _, e := range t.Elems {
									protoInfo.EncryptionAlgos = append(protoInfo.EncryptionAlgos, e.Value)
									UniqueEncryptionAlgos = valueExists(UniqueEncryptionAlgos, e)
								}
							case "compression_algorithms":
								for _, e := range t.Elems {
									protoInfo.CompressionAlgos = append(protoInfo.CompressionAlgos, e.Value)
									UniqueCompressionAlgos = valueExists(UniqueCompressionAlgos, e)
								}
							case "server_host_key_algorithms":
								for _, e := range t.Elems {
									protoInfo.ServerHostKeyAlgos = append(protoInfo.ServerHostKeyAlgos, e.Value)
									UniqueServerHostKeyAlgos = valueExists(UniqueServerHostKeyAlgos, e)
								}
							}

						}
					}
				}
				//fmt.Println(protoInfo)
				if protoInfo.Port != "" {
					sshInfo.SSHInfo = append(sshInfo.SSHInfo, protoInfo)
				}
			}
		}
	}
	//fmt.Println("****************")
	//fmt.Println(UniqueCompressionAlgos)

	// Create CSV file
	csvFile, err := os.Create("ssh_algos.csv")
	if err != nil {
		fmt.Printf("Error creating CSV file: %v\n", err)
		return
	}
	defer csvFile.Close()

	writer := csv.NewWriter(csvFile)
	defer writer.Flush()

	// Build CSV Header
	headerStr := []string{"IP", "Hostname", "Port", "Protocol", "ServiceName", "ServiceProduct", "ServiceVersion"}
	// Add Kex Algorithms
	sort.Strings(UniqueKexAlgorithms)
	sort.Strings(UniqueServerHostKeyAlgos)
	sort.Strings(UniqueEncryptionAlgos)
	sort.Strings(UniqueMACAlgorithms)
	sort.Strings(UniqueCompressionAlgos)
	headerStr = AddHeaderValue(UniqueKexAlgorithms, headerStr, "KexAlgos")
	headerStr = AddHeaderValue(UniqueServerHostKeyAlgos, headerStr, "ServerHostKeyAlgos")
	headerStr = AddHeaderValue(UniqueEncryptionAlgos, headerStr, "EncryptionAlgos")
	headerStr = AddHeaderValue(UniqueMACAlgorithms, headerStr, "MACAlgorithms")
	headerStr = AddHeaderValue(UniqueCompressionAlgos, headerStr, "CompressionAlgos")

	//fmt.Println(headerStr)
	err = writer.Write(headerStr)
	if err != nil {
		fmt.Printf("Error writing CSV header: %v\n", err)
	}

	// Go through each host with ports where the ssh-enum-algos script executed...
	for _, host := range sshInfo.SSHInfo {
		csvLine := []string{}
		csvLine = append(csvLine, host.IP)
		csvLine = append(csvLine, host.Hostname)
		csvLine = append(csvLine, host.Port)
		csvLine = append(csvLine, host.Protocol)
		csvLine = append(csvLine, host.ServiceName)
		csvLine = append(csvLine, host.ServiceProduct)
		csvLine = append(csvLine, host.ServiceVersion)
		// Kex Algorithms
		csvLine = appendCSVLine(csvLine, UniqueKexAlgorithms, host.KexAlgorithms)
		// Server Host Key Algos
		csvLine = appendCSVLine(csvLine, UniqueServerHostKeyAlgos, host.ServerHostKeyAlgos)
		// Encryption Algos
		csvLine = appendCSVLine(csvLine, UniqueEncryptionAlgos, host.EncryptionAlgos)
		// MAC Algos
		csvLine = appendCSVLine(csvLine, UniqueMACAlgorithms, host.MACAlgorithms)
		// Compression Algos
		csvLine = appendCSVLine(csvLine, UniqueCompressionAlgos, host.CompressionAlgos)
		err = writer.Write(csvLine)
		if err != nil {
			fmt.Printf("Error writing CSV header: %v\n", err)
		}
	}
}
