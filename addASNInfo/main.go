package main

import (
	"bufio"
	"encoding/csv"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"math/big"
	"net"
	"os"
	"runtime"
	"strings"
	"sync"
)

/**
Downloaded the original ASN file from https://ipinfo.io/products/free-ip-database it has IP to Country + ASN JSON file

Added the feature to use half of the cores provided by the device
- During testing I found that 10 semaphores it took about 11 minutes
- With all of my cores about 4 minutes on the same above test file

**/

type ASNStruct struct {
	ASNInfo []ASNInfoStruct `json:"asnInfo"`
}

type ASNInfoStruct struct {
	Network       string   `json:"network"`
	StartIP       string   `json:"start_ip,omitempty"`
	StartDecIP    *big.Int `json:"startDecIP,omitempty"`
	EndIP         string   `json:"end_ip,omitempty"`
	EndDecIP      *big.Int `json:"endDecIP,omitempty"`
	Country       string   `json:"country"`
	CountryCode   string   `json:"country_code"`
	Continent     string   `json:"continent"`
	ContinentCode string   `json:"continent_code"`
	ASN           string   `json:"asn"`
	ASNName       string   `json:"as_name"`
	ASDomain      string   `json:"as_domain"`
}

func GetFirstAndLastIP(cidr string) (string, string, error) {
	ip, ipNet, err := net.ParseCIDR(cidr)
	if err != nil {
		return "", "", err
	}

	// Convert IP mask to 4-byte mask
	mask := ipNet.Mask
	networkIP := ip.Mask(mask)

	// First IP (network IP + 1)
	firstIP := make(net.IP, len(networkIP))
	copy(firstIP, networkIP)

	// Last IP = Broadcast IP - 1
	broadcastIP := make(net.IP, len(networkIP))
	copy(broadcastIP, networkIP)
	for i := 0; i < len(mask); i++ {
		broadcastIP[i] |= ^mask[i]
	}
	lastIP := broadcastIP

	return firstIP.String(), lastIP.String(), nil
}

func (s *ASNStruct) LoadFile(sPtr string) error {
	asnFile, err := os.Open(sPtr)
	if err != nil {
		return err
	}
	defer asnFile.Close()
	decoder := json.NewDecoder(asnFile)
	if err := decoder.Decode(&s); err != nil {
		return err
	}

	return nil
}

func (s *ASNStruct) SaveNewFile(sPtr string) error {
	jsonData, err := json.MarshalIndent(s, "", "    ")
	if err != nil {
		return err
	}

	err = os.WriteFile(sPtr, jsonData, 0644)
	if err != nil {
		return err
	}

	return nil
}

func restructureJSON(f string) ASNStruct {
	var asn ASNStruct
	file, err := os.Open(f)
	if err != nil {
		log.Fatalf("Failed to open file: %s", err)
	}
	defer file.Close()

	// Create a new Scanner to read the file line by line
	scanner := bufio.NewScanner(file)

	// Read each line
	for scanner.Scan() {

		var asnInfo ASNInfoStruct
		line := scanner.Text() // Get the current line as a string
		//decoder := json.NewDecoder([]byte(line))
		err = json.Unmarshal([]byte(line), &asnInfo)
		if err != nil {
			log.Fatalf("Unable to decode: %v", err)
		}
		if strings.Contains(asnInfo.Network, "/") {
			asnInfo.StartIP, asnInfo.EndIP, err = GetFirstAndLastIP(asnInfo.Network)
			if err != nil {
				fmt.Println("\nDebug...")
				fmt.Println(line)
				log.Fatalf("unable to get first and last IP Address\n%v", err)
			}
		} else {
			asnInfo.StartIP = asnInfo.Network
			asnInfo.EndIP = asnInfo.Network
		}

		//fmt.Printf("Start IP: %s  -  End IP: %s\n", asnInfo.StartIP, asnInfo.EndIP)

		asnInfo.StartDecIP, err = ipToDecimal(asnInfo.StartIP)
		if err != nil {
			log.Printf("Unable to convert IP Address: %s\n%v\n", asnInfo.StartIP, err)
		}
		asnInfo.EndDecIP, err = ipToDecimal(asnInfo.EndIP)
		if err != nil {
			log.Printf("Unable to convert IP Address: %s\n%v\n", asnInfo.EndIP, err)
		}

		asn.ASNInfo = append(asn.ASNInfo, asnInfo)
		//fmt.Println(line) // Print the line
	}
	// Check for any errors that occurred during scanning
	if err := scanner.Err(); err != nil {
		log.Fatalf("Error reading file: %s", err)
	}
	file.Close()
	return asn

}

func ipToDecimal(ipStr string) (*big.Int, error) {
	// Parse the IP address
	ip := net.ParseIP(ipStr)
	if ip == nil {
		return nil, fmt.Errorf("invalid IP address: %s", ipStr)
	}

	// Convert the IP address to a byte slice
	var ipBytes []byte
	if ip.To4() != nil {
		// IPv4 address
		ipBytes = ip.To4()
	} else {
		// IPv6 address
		ipBytes = ip.To16()
	}

	// Convert the byte slice to a big.Int
	ipInt := new(big.Int)
	ipInt.SetBytes(ipBytes)

	return ipInt, nil
}

func inputFromStdin() string {

	reader := bufio.NewReader(os.Stdin)
	fmt.Println("Enter IP to find ASN:")
	input, err := reader.ReadString('\n')
	if err != nil {
		log.Fatalln("[E] Error reading input:", err)
	}
	fmt.Println()
	input = strings.Replace(input, "\r", "", -1)
	input = strings.Replace(input, "\n", "", -1)
	return input
}

func loopIPList(ipList []string) {
	var wg sync.WaitGroup
	halfCores := runtime.NumCPU() / 2
	semaphore := make(chan struct{}, halfCores*2)
	for _, ip := range ipList {
		wg.Add(1)
		go func(ip string) {
			defer wg.Done()
			semaphore <- struct{}{}
			defer func() { <-semaphore }()
			findASN(ip)
		}(ip)
	}

	wg.Wait()
}

func removeBadChars(s string) string {
	s = strings.ReplaceAll(s, ",", "") // Some of the ASNames have a comma
	return s
}

func findASN(ip string) {
	ipDec, err := ipToDecimal(ip)
	if err != nil {
		log.Printf("Unable to convert IP %s to Decimal\n%v\n", ip, err)
	}
	for _, asn := range asn.ASNInfo {
		matchStart := false
		matchEnd := false
		// ipDec should be equal or greater than the asn.StartDecIP
		switch ipDec.Cmp(asn.StartDecIP) {
		case -1:
			continue
		case 0:
			matchStart = true
		case 1:
			matchStart = true
		}

		// ipDec should be equal or less than the asn.EndDecIP
		switch ipDec.Cmp(asn.EndDecIP) {
		case -1:
			matchEnd = true
		case 0:
			matchEnd = true
		case 1:
			continue
		}

		if matchStart && matchEnd {
			// Output the csv of the information...

			outputString := []string{ip, asn.ASN, removeBadChars(asn.ASNName), removeBadChars(asn.ASDomain), asn.Country, removeBadChars(asn.CountryCode), asn.ContinentCode}

			writer := csv.NewWriter(os.Stdout)

			// Write the strings as a single CSV record
			err := writer.Write(outputString)
			if err != nil {
				fmt.Println("Error writing CSV:", err)
				return
			}

			// Flush the writer to ensure all data is written
			writer.Flush()

			// Check if there were any errors during flushing
			if err := writer.Error(); err != nil {
				fmt.Println("Error flushing CSV writer:", err)
				return
			}
		}
	}

}

func loadIPFile(f string) []string {
	file, err := os.Open(f)
	if err != nil {
		log.Fatalf("Failed to open file: %s", err)
	}
	defer file.Close()

	// Create a new Scanner to read the file line by line
	scanner := bufio.NewScanner(file)

	var ipList []string
	for scanner.Scan() {
		line := scanner.Text() // Get the current line as a string
		line = strings.ReplaceAll(line, "\r", "")
		line = strings.ReplaceAll(line, "\n", "")
		ipList = append(ipList, line)
	}
	// Check for any errors that occurred during scanning
	if err := scanner.Err(); err != nil {
		log.Fatalf("Error building IP List from file: %s", err)
	}
	file.Close()
	ipList = dedupStrings(ipList)
	return ipList
}

func dedupStrings(slice []string) []string {
	// Create a map to track unique strings
	seen := make(map[string]bool)
	result := []string{}

	// Iterate over the slice
	for _, item := range slice {
		// If the string hasn't been seen, add it to the result
		if !seen[item] {
			seen[item] = true
			result = append(result, item)
		}
	}

	return result
}

var asn ASNStruct

func main() {
	// Use half of the cores provided by the device
	halfCores := runtime.NumCPU() / 2
	//fmt.Printf("Using %d CPU cores to accomplish the task...\n", halfCores)
	runtime.GOMAXPROCS(halfCores)

	//var asn ASNStruct
	origFilePtr := flag.String("original", "", "Load the original file from ipinfo to restructure JSON")
	asnFilePtr := flag.String("a", "", "Load ASN File to be Used")
	gatherInputPtr := flag.Bool("i", false, "Gather user input to search for the IP Address")
	ipAddrFilePtr := flag.String("f", "", "Load a list of IP Addresses to analyze")
	flag.Parse()

	if *origFilePtr != "" {
		asn = restructureJSON(*origFilePtr)
		asn.SaveNewFile("restructured.json")
		fmt.Println("Created new JSON file as restructured.json")
		os.Exit(0)
	}

	if *asnFilePtr != "" {
		err := asn.LoadFile(*asnFilePtr)
		if err != nil {
			log.Fatalf("Unable to load ASN File specified %s: %v\n", *asnFilePtr, err)
		}

		ipList := []string{}
		if *gatherInputPtr {
			ipList = append(ipList, inputFromStdin())
		} else if *ipAddrFilePtr != "" {
			ipList = loadIPFile(*ipAddrFilePtr)
		}

		if len(ipList) > 0 {
			fmt.Println("\"IP\",\"ASN\",\"ASNName\",\"ASDomain\",\"Country\",\"CountryName\",\"ContinentName\"")
			loopIPList(ipList)
		} else {
			fmt.Println("No IP Addresses were loaded to be evaluated...")
			flag.PrintDefaults()
		}
	} else {
		flag.PrintDefaults()
	}

}
