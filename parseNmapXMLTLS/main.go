package main

import (
	"encoding/csv"
	"encoding/xml"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
)

/**
Sample nmap command:
nmap -sV --script ssl-enum-ciphers -p443 sample.local -oA sample_output
**/

type NmapRun struct {
	XMLName   xml.Name  `xml:"nmaprun"`
	Scanner   string    `xml:"scanner,attr"`
	Args      string    `xml:"args,attr"`
	Start     int64     `xml:"start,attr"`
	StartStr  string    `xml:"startstr,attr"`
	Version   string    `xml:"version,attr"`
	ScanInfo  ScanInfo  `xml:"scaninfo"`
	Verbose   Verbose   `xml:"verbose"`
	Debugging Debugging `xml:"debugging"`
	HostHint  HostHint  `xml:"hosthint"`
	Host      Host      `xml:"host"`
	RunStats  RunStats  `xml:"runstats"`
}

type ScanInfo struct {
	Type        string `xml:"type,attr"`
	Protocol    string `xml:"protocol,attr"`
	NumServices int    `xml:"numservices,attr"`
	Services    string `xml:"services,attr"`
}

type Verbose struct {
	Level int `xml:"level,attr"`
}

type Debugging struct {
	Level int `xml:"level,attr"`
}

type HostHint struct {
	Status    Status     `xml:"status"`
	Addresses []Address  `xml:"address"`
	Hostnames []Hostname `xml:"hostnames>hostname"`
}

type Host struct {
	StartTime int64      `xml:"starttime,attr"`
	EndTime   int64      `xml:"endtime,attr"`
	Status    Status     `xml:"status"`
	Addresses []Address  `xml:"address"`
	Hostnames []Hostname `xml:"hostnames>hostname"`
	Ports     []Port     `xml:"ports>port"`
	Times     Times      `xml:"times"`
}

type Status struct {
	State  string `xml:"state,attr"`
	Reason string `xml:"reason,attr"`
	TTL    int    `xml:"reason_ttl,attr"`
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

type Times struct {
	SRTT   int `xml:"srtt,attr"`
	RTTVar int `xml:"rttvar,attr"`
	To     int `xml:"to,attr"`
}

type RunStats struct {
	Finished Finished `xml:"finished"`
	Hosts    Hosts    `xml:"hosts"`
}

type Finished struct {
	Time    int64  `xml:"time,attr"`
	TimeStr string `xml:"timestr,attr"`
	Summary string `xml:"summary,attr"`
	Elapsed string `xml:"elapsed,attr"`
	Exit    string `xml:"exit,attr"`
}

type Hosts struct {
	Up    int `xml:"up,attr"`
	Down  int `xml:"down,attr"`
	Total int `xml:"total,attr"`
}

// cat sample_output.xml | grep "key=\"name" | sort | uniq | sed 's/<elem key="name">//' | sed 's@</elem>@@' | sed 's/$/ string/'
type ProtocolInformation struct {
	IP                                            string
	Hostname                                      string
	Port                                          string
	TLS1_0Supported                               bool
	TLS1_1Supported                               bool
	TLS1_2Supported                               bool
	TLS1_3Supported                               bool
	TLS_AKE_WITH_AES_128_GCM_SHA256               string
	TLS_AKE_WITH_AES_256_GCM_SHA384               string
	TLS_AKE_WITH_CHACHA20_POLY1305_SHA256         string
	TLS_DHE_RSA_WITH_3DES_EDE_CBC_SHA             string
	TLS_DHE_RSA_WITH_AES_128_CBC_SHA256           string
	TLS_DHE_RSA_WITH_AES_128_CBC_SHA              string
	TLS_DHE_RSA_WITH_AES_128_GCM_SHA256           string
	TLS_DHE_RSA_WITH_AES_256_CBC_SHA256           string
	TLS_DHE_RSA_WITH_AES_256_CBC_SHA              string
	TLS_DHE_RSA_WITH_AES_256_GCM_SHA384           string
	TLS_DHE_RSA_WITH_CAMELLIA_128_CBC_SHA         string
	TLS_DHE_RSA_WITH_CAMELLIA_256_CBC_SHA         string
	TLS_DHE_RSA_WITH_CHACHA20_POLY1305_SHA256     string
	TLS_DHE_RSA_WITH_SEED_CBC_SHA                 string
	TLS_ECDHE_RSA_WITH_3DES_EDE_CBC_SHA           string
	TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA256         string
	TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA            string
	TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256         string
	TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA384         string
	TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA            string
	TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384         string
	TLS_ECDHE_RSA_WITH_ARIA_128_GCM_SHA256        string
	TLS_ECDHE_RSA_WITH_ARIA_256_GCM_SHA384        string
	TLS_ECDHE_RSA_WITH_CAMELLIA_128_CBC_SHA256    string
	TLS_ECDHE_RSA_WITH_CAMELLIA_256_CBC_SHA384    string
	TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305_SHA256   string
	TLS_ECDHE_RSA_WITH_RC4_128_SHA                string
	TLS_RSA_WITH_3DES_EDE_CBC_SHA                 string
	TLS_RSA_WITH_AES_128_CBC_SHA256               string
	TLS_RSA_WITH_AES_128_CBC_SHA                  string
	TLS_RSA_WITH_AES_128_CCM_8                    string
	TLS_RSA_WITH_AES_128_CCM                      string
	TLS_RSA_WITH_AES_128_GCM_SHA256               string
	TLS_RSA_WITH_AES_256_CBC_SHA256               string
	TLS_RSA_WITH_AES_256_CBC_SHA                  string
	TLS_RSA_WITH_AES_256_CCM_8                    string
	TLS_RSA_WITH_AES_256_CCM                      string
	TLS_RSA_WITH_AES_256_GCM_SHA384               string
	TLS_RSA_WITH_ARIA_128_GCM_SHA256              string
	TLS_RSA_WITH_ARIA_256_GCM_SHA384              string
	TLS_RSA_WITH_CAMELLIA_128_CBC_SHA256          string
	TLS_RSA_WITH_CAMELLIA_128_CBC_SHA             string
	TLS_RSA_WITH_CAMELLIA_256_CBC_SHA256          string
	TLS_RSA_WITH_CAMELLIA_256_CBC_SHA             string
	TLS_RSA_WITH_IDEA_CBC_SHA                     string
	TLS_RSA_WITH_RC4_128_MD5                      string
	TLS_RSA_WITH_RC4_128_SHA                      string
	TLS_RSA_WITH_SEED_CBC_SHA                     string
	TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA          string
	TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA256       string
	TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256       string
	TLS_ECDHE_ECDSA_WITH_AES_256_CBC_SHA384       string
	TLS_ECDHE_ECDSA_WITH_AES_256_CBC_SHA          string
	TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384       string
	TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305_SHA256 string
}

func main() {
	var xmlFiles []string

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

	// Create CSV file
	csvFile, err := os.Create("ssl_ciphers.csv")
	if err != nil {
		fmt.Printf("Error creating CSV file: %v\n", err)
		return
	}
	defer csvFile.Close()

	writer := csv.NewWriter(csvFile)
	defer writer.Flush()

	// Write CSV header
	// cat temp.list | awk '{print $1}' | sed 's/$/",/' | sed 's/^/"/' | tr -d '\n'
	header := []string{"IP", "Hostname", "Port", "TLS1_0Supported", "TLS1_1Supported", "TLS1_2Supported", "TLS1_3Supported", "TLS_AKE_WITH_AES_128_GCM_SHA256", "TLS_AKE_WITH_AES_256_GCM_SHA384", "TLS_AKE_WITH_CHACHA20_POLY1305_SHA256", "TLS_DHE_RSA_WITH_3DES_EDE_CBC_SHA", "TLS_DHE_RSA_WITH_AES_128_CBC_SHA256", "TLS_DHE_RSA_WITH_AES_128_CBC_SHA", "TLS_DHE_RSA_WITH_AES_128_GCM_SHA256", "TLS_DHE_RSA_WITH_AES_256_CBC_SHA256", "TLS_DHE_RSA_WITH_AES_256_CBC_SHA", "TLS_DHE_RSA_WITH_AES_256_GCM_SHA384", "TLS_DHE_RSA_WITH_CAMELLIA_128_CBC_SHA", "TLS_DHE_RSA_WITH_CAMELLIA_256_CBC_SHA", "TLS_DHE_RSA_WITH_CHACHA20_POLY1305_SHA256", "TLS_DHE_RSA_WITH_SEED_CBC_SHA", "TLS_ECDHE_RSA_WITH_3DES_EDE_CBC_SHA", "TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA256", "TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA", "TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256", "TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA384", "TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA", "TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384", "TLS_ECDHE_RSA_WITH_ARIA_128_GCM_SHA256", "TLS_ECDHE_RSA_WITH_ARIA_256_GCM_SHA384", "TLS_ECDHE_RSA_WITH_CAMELLIA_128_CBC_SHA256", "TLS_ECDHE_RSA_WITH_CAMELLIA_256_CBC_SHA384", "TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305_SHA256", "TLS_ECDHE_RSA_WITH_RC4_128_SHA", "TLS_RSA_WITH_3DES_EDE_CBC_SHA", "TLS_RSA_WITH_AES_128_CBC_SHA256", "TLS_RSA_WITH_AES_128_CBC_SHA", "TLS_RSA_WITH_AES_128_CCM_8", "TLS_RSA_WITH_AES_128_CCM", "TLS_RSA_WITH_AES_128_GCM_SHA256", "TLS_RSA_WITH_AES_256_CBC_SHA256", "TLS_RSA_WITH_AES_256_CBC_SHA", "TLS_RSA_WITH_AES_256_CCM_8", "TLS_RSA_WITH_AES_256_CCM", "TLS_RSA_WITH_AES_256_GCM_SHA384", "TLS_RSA_WITH_ARIA_128_GCM_SHA256", "TLS_RSA_WITH_ARIA_256_GCM_SHA384", "TLS_RSA_WITH_CAMELLIA_128_CBC_SHA256", "TLS_RSA_WITH_CAMELLIA_128_CBC_SHA", "TLS_RSA_WITH_CAMELLIA_256_CBC_SHA256", "TLS_RSA_WITH_CAMELLIA_256_CBC_SHA", "TLS_RSA_WITH_IDEA_CBC_SHA", "TLS_RSA_WITH_RC4_128_MD5", "TLS_RSA_WITH_RC4_128_SHA", "TLS_RSA_WITH_SEED_CBC_SHA", "TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA", "TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA256", "TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256", "TLS_ECDHE_ECDSA_WITH_AES_256_CBC_SHA384", "TLS_ECDHE_ECDSA_WITH_AES_256_CBC_SHA", "TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384", "TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305_SHA256"}
	err = writer.Write(header)
	if err != nil {
		fmt.Printf("Error writing CSV header: %v\n", err)
		return
	}

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

		//var TLSProtocol string
		//var ProtoCiphersCompressors string

		for _, p := range nmapRun.Host.Ports {
			var protoInfo ProtocolInformation
			evaluatedCiphers := false
			protoInfo.IP = nmapRun.Host.Addresses[0].Addr
			if len(nmapRun.Host.Hostnames) > 0 {
				protoInfo.Hostname = nmapRun.Host.Hostnames[0].Name
			}

			protoInfo.Port = strconv.Itoa(p.PortID) + "/" + p.Protocol + " " + p.Service.Name + " - " + p.Service.Product
			protoInfo.TLS1_0Supported, protoInfo.TLS1_1Supported, protoInfo.TLS1_2Supported, protoInfo.TLS1_3Supported = false, false, false, false
			//protoInfo.Hostname = nmapRun.Address.Addr
			for _, s := range p.Scripts {
				for _, t := range s.Tables {
					switch t.Key {
					case "TLSv1.0":
						protoInfo.TLS1_0Supported = true
					case "TLSv1.1":
						protoInfo.TLS1_1Supported = true
					case "TLSv1.2":
						protoInfo.TLS1_2Supported = true
					case "TLSv1.3":
						protoInfo.TLS1_3Supported = true
					}
					if protoInfo.TLS1_0Supported || protoInfo.TLS1_1Supported || protoInfo.TLS1_2Supported || protoInfo.TLS1_3Supported {
						//fmt.Println(t.Key)
						for _, t2 := range t.Tables {
							//ProtoCiphersCompressors = t2.Key
							//fmt.Println(t2.Key)
							for _, e := range t2.Tables {
								//fmt.Println(e.Elems)
								type cipherInfo struct {
									Strength string
									KexInfo  string
									Name     string
								}
								var c cipherInfo
								c.Strength = "NA"
								c.KexInfo = "NA"
								c.Name = "NA"
								for _, i := range e.Elems {
									if i.Key == "strength" {
										c.Strength = i.Value
									}
									if i.Key == "kex_info" {
										c.KexInfo = i.Value
									}
									if i.Key == "name" {
										c.Name = i.Value
									}
									//fmt.Println(i.Key)
									//cat sample_output.xml | grep "key=\"name" | sort | uniq | sed 's/<elem key="name">//' | sed 's@</elem>@@' | awk '{print "case \"" $1 "\":\n    protoInfo." $1 " = true" }'
									//cat listCiphers.txt | awk '{print "case \"" $1 "\":\n    protoInfo." $1 " = c.Strength\n    evaluatedCiphers = true" }'
									//fmt.Println(i.Value)
									// Found if the strength follows the name of the cipher the strength is not recorded due to this order.  Pulling out the strength for now...
									// Fix Y in Yahoo
									// Fix J in Yahoo
								}
								//fmt.Printf("DEBUG: %s, %s, %s, %s\n", t.Key, c.Name, c.KexInfo, c.Strength)
								switch c.Name {
								case "TLS_AKE_WITH_AES_128_GCM_SHA256":
									protoInfo.TLS_AKE_WITH_AES_128_GCM_SHA256 = c.Strength
									evaluatedCiphers = true
								case "TLS_AKE_WITH_AES_256_GCM_SHA384":
									protoInfo.TLS_AKE_WITH_AES_256_GCM_SHA384 = c.Strength
									evaluatedCiphers = true
								case "TLS_AKE_WITH_CHACHA20_POLY1305_SHA256":
									protoInfo.TLS_AKE_WITH_CHACHA20_POLY1305_SHA256 = c.Strength
									evaluatedCiphers = true
								case "TLS_DHE_RSA_WITH_3DES_EDE_CBC_SHA":
									protoInfo.TLS_DHE_RSA_WITH_3DES_EDE_CBC_SHA = c.Strength
									evaluatedCiphers = true
								case "TLS_DHE_RSA_WITH_AES_128_CBC_SHA256":
									protoInfo.TLS_DHE_RSA_WITH_AES_128_CBC_SHA256 = c.Strength
									evaluatedCiphers = true
								case "TLS_DHE_RSA_WITH_AES_128_CBC_SHA":
									protoInfo.TLS_DHE_RSA_WITH_AES_128_CBC_SHA = c.Strength
									evaluatedCiphers = true
								case "TLS_DHE_RSA_WITH_AES_128_GCM_SHA256":
									protoInfo.TLS_DHE_RSA_WITH_AES_128_GCM_SHA256 = c.Strength
									evaluatedCiphers = true
								case "TLS_DHE_RSA_WITH_AES_256_CBC_SHA256":
									protoInfo.TLS_DHE_RSA_WITH_AES_256_CBC_SHA256 = c.Strength
									evaluatedCiphers = true
								case "TLS_DHE_RSA_WITH_AES_256_CBC_SHA":
									protoInfo.TLS_DHE_RSA_WITH_AES_256_CBC_SHA = c.Strength
									evaluatedCiphers = true
								case "TLS_DHE_RSA_WITH_AES_256_GCM_SHA384":
									protoInfo.TLS_DHE_RSA_WITH_AES_256_GCM_SHA384 = c.Strength
									evaluatedCiphers = true
								case "TLS_DHE_RSA_WITH_CAMELLIA_128_CBC_SHA":
									protoInfo.TLS_DHE_RSA_WITH_CAMELLIA_128_CBC_SHA = c.Strength
									evaluatedCiphers = true
								case "TLS_DHE_RSA_WITH_CAMELLIA_256_CBC_SHA":
									protoInfo.TLS_DHE_RSA_WITH_CAMELLIA_256_CBC_SHA = c.Strength
									evaluatedCiphers = true
								case "TLS_DHE_RSA_WITH_CHACHA20_POLY1305_SHA256":
									protoInfo.TLS_DHE_RSA_WITH_CHACHA20_POLY1305_SHA256 = c.Strength
									evaluatedCiphers = true
								case "TLS_DHE_RSA_WITH_SEED_CBC_SHA":
									protoInfo.TLS_DHE_RSA_WITH_SEED_CBC_SHA = c.Strength
									evaluatedCiphers = true
								case "TLS_ECDHE_RSA_WITH_3DES_EDE_CBC_SHA":
									protoInfo.TLS_ECDHE_RSA_WITH_3DES_EDE_CBC_SHA = c.Strength
									evaluatedCiphers = true
								case "TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA256":
									protoInfo.TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA256 = c.Strength
									evaluatedCiphers = true
								case "TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA":
									protoInfo.TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA = c.Strength
									evaluatedCiphers = true
								case "TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256":
									protoInfo.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256 = c.Strength
									evaluatedCiphers = true
								case "TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA384":
									protoInfo.TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA384 = c.Strength
									evaluatedCiphers = true
								case "TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA":
									protoInfo.TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA = c.Strength
									evaluatedCiphers = true
								case "TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384":
									protoInfo.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384 = c.Strength
									evaluatedCiphers = true
								case "TLS_ECDHE_RSA_WITH_ARIA_128_GCM_SHA256":
									protoInfo.TLS_ECDHE_RSA_WITH_ARIA_128_GCM_SHA256 = c.Strength
									evaluatedCiphers = true
								case "TLS_ECDHE_RSA_WITH_ARIA_256_GCM_SHA384":
									protoInfo.TLS_ECDHE_RSA_WITH_ARIA_256_GCM_SHA384 = c.Strength
									evaluatedCiphers = true
								case "TLS_ECDHE_RSA_WITH_CAMELLIA_128_CBC_SHA256":
									protoInfo.TLS_ECDHE_RSA_WITH_CAMELLIA_128_CBC_SHA256 = c.Strength
									evaluatedCiphers = true
								case "TLS_ECDHE_RSA_WITH_CAMELLIA_256_CBC_SHA384":
									protoInfo.TLS_ECDHE_RSA_WITH_CAMELLIA_256_CBC_SHA384 = c.Strength
									evaluatedCiphers = true
								case "TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305_SHA256":
									protoInfo.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305_SHA256 = c.Strength
									evaluatedCiphers = true
								case "TLS_ECDHE_RSA_WITH_RC4_128_SHA":
									protoInfo.TLS_ECDHE_RSA_WITH_RC4_128_SHA = c.Strength
									evaluatedCiphers = true
								case "TLS_RSA_WITH_3DES_EDE_CBC_SHA":
									protoInfo.TLS_RSA_WITH_3DES_EDE_CBC_SHA = c.Strength
									evaluatedCiphers = true
								case "TLS_RSA_WITH_AES_128_CBC_SHA256":
									protoInfo.TLS_RSA_WITH_AES_128_CBC_SHA256 = c.Strength
									evaluatedCiphers = true
								case "TLS_RSA_WITH_AES_128_CBC_SHA":
									protoInfo.TLS_RSA_WITH_AES_128_CBC_SHA = c.Strength
									evaluatedCiphers = true
								case "TLS_RSA_WITH_AES_128_CCM_8":
									protoInfo.TLS_RSA_WITH_AES_128_CCM_8 = c.Strength
									evaluatedCiphers = true
								case "TLS_RSA_WITH_AES_128_CCM":
									protoInfo.TLS_RSA_WITH_AES_128_CCM = c.Strength
									evaluatedCiphers = true
								case "TLS_RSA_WITH_AES_128_GCM_SHA256":
									protoInfo.TLS_RSA_WITH_AES_128_GCM_SHA256 = c.Strength
									evaluatedCiphers = true
								case "TLS_RSA_WITH_AES_256_CBC_SHA256":
									protoInfo.TLS_RSA_WITH_AES_256_CBC_SHA256 = c.Strength
									evaluatedCiphers = true
								case "TLS_RSA_WITH_AES_256_CBC_SHA":
									protoInfo.TLS_RSA_WITH_AES_256_CBC_SHA = c.Strength
									evaluatedCiphers = true
								case "TLS_RSA_WITH_AES_256_CCM_8":
									protoInfo.TLS_RSA_WITH_AES_256_CCM_8 = c.Strength
									evaluatedCiphers = true
								case "TLS_RSA_WITH_AES_256_CCM":
									protoInfo.TLS_RSA_WITH_AES_256_CCM = c.Strength
									evaluatedCiphers = true
								case "TLS_RSA_WITH_AES_256_GCM_SHA384":
									protoInfo.TLS_RSA_WITH_AES_256_GCM_SHA384 = c.Strength
									evaluatedCiphers = true
								case "TLS_RSA_WITH_ARIA_128_GCM_SHA256":
									protoInfo.TLS_RSA_WITH_ARIA_128_GCM_SHA256 = c.Strength
									evaluatedCiphers = true
								case "TLS_RSA_WITH_ARIA_256_GCM_SHA384":
									protoInfo.TLS_RSA_WITH_ARIA_256_GCM_SHA384 = c.Strength
									evaluatedCiphers = true
								case "TLS_RSA_WITH_CAMELLIA_128_CBC_SHA256":
									protoInfo.TLS_RSA_WITH_CAMELLIA_128_CBC_SHA256 = c.Strength
									evaluatedCiphers = true
								case "TLS_RSA_WITH_CAMELLIA_128_CBC_SHA":
									protoInfo.TLS_RSA_WITH_CAMELLIA_128_CBC_SHA = c.Strength
									evaluatedCiphers = true
								case "TLS_RSA_WITH_CAMELLIA_256_CBC_SHA256":
									protoInfo.TLS_RSA_WITH_CAMELLIA_256_CBC_SHA256 = c.Strength
									evaluatedCiphers = true
								case "TLS_RSA_WITH_CAMELLIA_256_CBC_SHA":
									protoInfo.TLS_RSA_WITH_CAMELLIA_256_CBC_SHA = c.Strength
									evaluatedCiphers = true
								case "TLS_RSA_WITH_IDEA_CBC_SHA":
									protoInfo.TLS_RSA_WITH_IDEA_CBC_SHA = c.Strength
									evaluatedCiphers = true
								case "TLS_RSA_WITH_RC4_128_MD5":
									protoInfo.TLS_RSA_WITH_RC4_128_MD5 = c.Strength
									evaluatedCiphers = true
								case "TLS_RSA_WITH_RC4_128_SHA":
									protoInfo.TLS_RSA_WITH_RC4_128_SHA = c.Strength
									evaluatedCiphers = true
								case "TLS_RSA_WITH_SEED_CBC_SHA":
									protoInfo.TLS_RSA_WITH_SEED_CBC_SHA = c.Strength
									evaluatedCiphers = true
								case "TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA":
									protoInfo.TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA = c.Strength
									evaluatedCiphers = true
								case "TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA256":
									protoInfo.TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA256 = c.Strength
									evaluatedCiphers = true
								case "TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256":
									protoInfo.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256 = c.Strength
									evaluatedCiphers = true
								case "TLS_ECDHE_ECDSA_WITH_AES_256_CBC_SHA384":
									protoInfo.TLS_ECDHE_ECDSA_WITH_AES_256_CBC_SHA384 = c.Strength
									evaluatedCiphers = true
								case "TLS_ECDHE_ECDSA_WITH_AES_256_CBC_SHA":
									protoInfo.TLS_ECDHE_ECDSA_WITH_AES_256_CBC_SHA = c.Strength
									evaluatedCiphers = true
								case "TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384":
									protoInfo.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384 = c.Strength
									evaluatedCiphers = true
								case "TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305_SHA256":
									protoInfo.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305_SHA256 = c.Strength
									evaluatedCiphers = true
								}
							}
						}
					}
				}
			}
			//fmt.Println(protoInfo)
			// cat temp.list | awk '{print "strconv.FormatBool(protoInfo." $1 ")"}' | sed 's/$/,/' | tr -d '\n'
			if evaluatedCiphers && (protoInfo.TLS1_0Supported || protoInfo.TLS1_1Supported || protoInfo.TLS1_2Supported || protoInfo.TLS1_3Supported) {
				csvLine := []string{protoInfo.IP, protoInfo.Hostname, protoInfo.Port, strconv.FormatBool(protoInfo.TLS1_0Supported), strconv.FormatBool(protoInfo.TLS1_1Supported), strconv.FormatBool(protoInfo.TLS1_2Supported), strconv.FormatBool(protoInfo.TLS1_3Supported), protoInfo.TLS_AKE_WITH_AES_128_GCM_SHA256, protoInfo.TLS_AKE_WITH_AES_256_GCM_SHA384, protoInfo.TLS_AKE_WITH_CHACHA20_POLY1305_SHA256, protoInfo.TLS_DHE_RSA_WITH_3DES_EDE_CBC_SHA, protoInfo.TLS_DHE_RSA_WITH_AES_128_CBC_SHA256, protoInfo.TLS_DHE_RSA_WITH_AES_128_CBC_SHA, protoInfo.TLS_DHE_RSA_WITH_AES_128_GCM_SHA256, protoInfo.TLS_DHE_RSA_WITH_AES_256_CBC_SHA256, protoInfo.TLS_DHE_RSA_WITH_AES_256_CBC_SHA, protoInfo.TLS_DHE_RSA_WITH_AES_256_GCM_SHA384, protoInfo.TLS_DHE_RSA_WITH_CAMELLIA_128_CBC_SHA, protoInfo.TLS_DHE_RSA_WITH_CAMELLIA_256_CBC_SHA, protoInfo.TLS_DHE_RSA_WITH_CHACHA20_POLY1305_SHA256, protoInfo.TLS_DHE_RSA_WITH_SEED_CBC_SHA, protoInfo.TLS_ECDHE_RSA_WITH_3DES_EDE_CBC_SHA, protoInfo.TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA256, protoInfo.TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA, protoInfo.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256, protoInfo.TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA384, protoInfo.TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA, protoInfo.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384, protoInfo.TLS_ECDHE_RSA_WITH_ARIA_128_GCM_SHA256, protoInfo.TLS_ECDHE_RSA_WITH_ARIA_256_GCM_SHA384, protoInfo.TLS_ECDHE_RSA_WITH_CAMELLIA_128_CBC_SHA256, protoInfo.TLS_ECDHE_RSA_WITH_CAMELLIA_256_CBC_SHA384, protoInfo.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305_SHA256, protoInfo.TLS_ECDHE_RSA_WITH_RC4_128_SHA, protoInfo.TLS_RSA_WITH_3DES_EDE_CBC_SHA, protoInfo.TLS_RSA_WITH_AES_128_CBC_SHA256, protoInfo.TLS_RSA_WITH_AES_128_CBC_SHA, protoInfo.TLS_RSA_WITH_AES_128_CCM_8, protoInfo.TLS_RSA_WITH_AES_128_CCM, protoInfo.TLS_RSA_WITH_AES_128_GCM_SHA256, protoInfo.TLS_RSA_WITH_AES_256_CBC_SHA256, protoInfo.TLS_RSA_WITH_AES_256_CBC_SHA, protoInfo.TLS_RSA_WITH_AES_256_CCM_8, protoInfo.TLS_RSA_WITH_AES_256_CCM, protoInfo.TLS_RSA_WITH_AES_256_GCM_SHA384, protoInfo.TLS_RSA_WITH_ARIA_128_GCM_SHA256, protoInfo.TLS_RSA_WITH_ARIA_256_GCM_SHA384, protoInfo.TLS_RSA_WITH_CAMELLIA_128_CBC_SHA256, protoInfo.TLS_RSA_WITH_CAMELLIA_128_CBC_SHA, protoInfo.TLS_RSA_WITH_CAMELLIA_256_CBC_SHA256, protoInfo.TLS_RSA_WITH_CAMELLIA_256_CBC_SHA, protoInfo.TLS_RSA_WITH_IDEA_CBC_SHA, protoInfo.TLS_RSA_WITH_RC4_128_MD5, protoInfo.TLS_RSA_WITH_RC4_128_SHA, protoInfo.TLS_RSA_WITH_SEED_CBC_SHA, protoInfo.TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA, protoInfo.TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA256, protoInfo.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256, protoInfo.TLS_ECDHE_ECDSA_WITH_AES_256_CBC_SHA384, protoInfo.TLS_ECDHE_ECDSA_WITH_AES_256_CBC_SHA, protoInfo.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384, protoInfo.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305_SHA256}
				err = writer.Write(csvLine)
				if err != nil {
					fmt.Printf("Error writing CSV record: %v\n", err)
					return
				}
			}
		}
	}

}
