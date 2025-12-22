package main

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
	"sync"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

/** Script to launch the docker image

#!/bin/bash
docker run -d \
  --name neo4j \
  --rm \
  -p 127.0.0.1:7474:7474 \
  -p 127.0.0.1:7687:7687 \
  -v $(pwd)/neo4j_data:/data \
  -v $(pwd)/neo4j_logs:/logs \
  -e NEO4J_AUTH=neo4j/l0st1nSpac3 \
  neo4j:latest

**/

// With waitgroups and semaphores we were able to speed up the creation of the nodes and connections
// on list2025.txt from 2 minutes to 13.3 seconds.

// View the data in a graph
// MATCH (n)-[r]->(m) RETURN n, r, m

// Find the longest path between nodes
/**
MATCH path = (start:IPAddress)-[:TO*]->(end:IPAddress)
WHERE start <> end
RETURN path, length(path) AS pathLength
ORDER BY pathLength DESC
LIMIT 1
**/

/** Specify a start IP Address - Using list2025.txt as a csv list
MATCH path = (start:IPAddress)-[:TO*]->(end:IPAddress)
WHERE start <> end AND start.address = "10.0.0.117"
RETURN path, length(path) AS pathLength
ORDER BY pathLength DESC
**/

/** Specify a start and end IP Address - Using list2025.txt as a csv list
MATCH path = (start:IPAddress)-[:TO*]->(end:IPAddress)
WHERE start <> end AND start.address = "10.0.0.117" AND end.address = "10.4.25.206"
RETURN path, length(path) AS pathLength
ORDER BY pathLength DESC
**/

/** Specify a start IP address and an end Subnet but with the starts with ... list2025.txt
MATCH path = (start:IPAddress)-[:TO*]->(end:IPAddress)
WHERE start <> end AND start.address = "10.0.0.117" AND end.address STARTS WITH "10.4.25."
RETURN path, length(path) AS pathLength
ORDER BY pathLength DESC
**/

// Find the longest path between nodes with a specific protocol
/**
MATCH path = (start:IPAddress)-[rels:TO*]->(end:IPAddress)
WHERE start <> end AND ALL(r IN rels WHERE r.protocol = 'tcp')
RETURN path, length(path) AS pathLength
ORDER BY pathLength DESC
LIMIT 2    // Adjust the limit as needed to capture the longest paths...
**/

/** Find the paths with a start.address and a specific protocol .. list2025.txt
MATCH path = (start:IPAddress)-[rels:TO*]->(end:IPAddress)
WHERE start <> end AND start.address = "10.0.0.117" AND ALL(r IN rels WHERE r.protocol = 'tcp')
RETURN path, length(path) AS pathLength
ORDER BY pathLength DESC
**/

/** Find the paths from a node with IPAddressInternet to IPAddressExternal to IPAddress...
MATCH path = (internet:IPAddressInternet)-[:TO*1..]->(external:IPAddressExternal)-[:TO*1..]->(internal:IPAddress)
RETURN path
**/

/** If you have 6 nodes and want to find the relationships going to the internet without 6 connections from those 6 nodes, this would be finding anomolous connections outbound
MATCH (n:IPAddress)-[r]->(end:IPAddressInternet)
WHERE start <> end
WITH start, r, end
MATCH ()-[rel]->(end)
WITH start, r, end, count(rel) as relationshipCount
WHERE relationshipCount < 6
RETURN start, r, end


/** Calculate the minimum number of hops from an internet node to an internal node
MATCH path = (internet:IPAddressInternet)-[*1..]->(external:IPAddressExternal)-[*1..]->(lastNode:IPAddress)
//WITH path, lastNode, length(path) AS currentHopLength, internet
WHERE lastNode.hopLength IS NULL OR length(path) < lastNode.hopLength
  SET lastNode.hopLength = length(path)
  SET lastNode.hopSource = internet.address  // Stores source IP if needed
RETURN
  lastNode.address AS internal_ip,
  lastNode.hopLength AS hops,
  [n IN nodes(path) | n.address] AS full_path
**/

// Remove the data in the neo4j database
// MATCH (n) DETACH DELETE n;

/**
Use cases to develop:
1. Create a query that reads a csv file to populate the networkNickname and compliance fields in the IPNode nodes.
2. Create a query that reads a csv file to populate if the IP Address shows up in a vulnerability scan
3. Create a query that reads a csv file to populate if an agent is installed on the IP Address
4. Create a Node that is an Asset which connects to 1 or more IPNodes (This could be a device with more than one IP Address)


**/

// IPNode represents an IP address node in the graph
type IPNode struct {
	IPAddress        string
	NetworkNickName  string
	NodeType         string
	Compliance       string
	DestinationPorts []int
	DynamicFields    map[string]any // Optional field for dynamic configuration

}

// Connection represents a relationship between source and destination IPs
type NewConnection struct {
	SourceIP            string
	SourceNodeType      string
	DestinationIP       string
	DestinationNodeType string
	Protocol            string
	DestinationPort     string
	RuleName            string
	ConnectionStatus    string // e.g., "Allowed", "Blocked"
}

type Configuration struct {
	Neo4jURI             string         `json:"neo4juri"`
	Neo4jUsername        string         `json:"username"`
	Neo4jPassword        string         `json:"password"`
	ConcurrentProcessing int            `json:"concurrentProcessing"`    // Number of concurrent connections to the database
	Note                 string         `json:"_note"`                   // Note field to add comments or notes about the configuration
	DynamicFields        map[string]any `json:"dynamicFields,omitempty"` // Optional field for dynamic configuration
	ExternalNetworks     []string       `json:"externalNetworks"`        // List of external networks to be used in the graph
	InternalNetworks     []string       `json:"internalNetworks"`        // List of internal networks to be used in the graph
	IgnoredNetworks      []string       `json:"ignoredNetworks"`         // List of networks to ignore in the graph
}

func (c *Configuration) CreateConfig(f string) error {
	c.Neo4jURI = "bolt://localhost:7687"
	c.Neo4jUsername = "neo4j"
	c.Neo4jPassword = "l0st1nSpac3"
	c.ConcurrentProcessing = 10
	c.Note = "When specifying a dynamicField for an IPNode, provide an example of the output whether a string, int, or ... The field needs to be capitalized... Fields below need to match the 1st column in a supporting csv"
	c.DynamicFields = map[string]any{
		"VulnScan": "False",
		"Agent":    "Missing",
	}
	c.ExternalNetworks = []string{"145.6.0.0/16", "165.6.0.0/16"}
	c.InternalNetworks = []string{"10.0.0.0/8", "172.16.0.0/12", "192.168.0.0/16"}
	c.IgnoredNetworks = []string{"169.254.0.0/16"}
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

type CSVFileStruct struct {
	Rows []CSVRowStruct
}

/**
// CSVRowStruct for IP Nodes - Leaving this here in case it is used in the future
type CSVRowStruct struct {
	IPAddress        string `csv:"ip_address"`
	NetworkNickName  string `csv:"network_nick_name"`
	Compliance       string `csv:"compliance"`
	DestinationPorts []int
}
**/

type CSVRowStruct struct {
	SourceIP         string `csv:"source_ip"`
	DestinationIP    string `csv:"destination_ip"`
	DestinationPort  string `csv:"destination_port"`
	Protocol         string `csv:"protocol"`
	RuleName         string `csv:"rule_name,omitempty"`
	ConnectionStatus string `csv:"connection_status,omitempty"`
}

/**
			SourceIP:        "192.168.1.1",
			DestinationIP:   "10.0.0.5",
			DestinationPort: 443,
			Protocol:        "tcp",
			ruleName:        "Allow HTTPS",
			ConnectionStatus: "Allowed",
**/

func (csvFile *CSVFileStruct) LoadCSVFile(filePath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("could not open CSV file: %v", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return fmt.Errorf("could not read CSV file: %v", err)
	}

	for _, record := range records[1:] { // Skip header row
		if len(record) < 4 {
			continue // Skip rows with insufficient data
		}
		csvRow := CSVRowStruct{
			SourceIP:         record[0],
			DestinationIP:    record[1],
			DestinationPort:  record[2],
			Protocol:         record[3],
			RuleName:         record[4],
			ConnectionStatus: record[5],
		}
		csvFile.Rows = append(csvFile.Rows, csvRow)
	}

	return nil
}

// createIPNode creates a node for an IP address with its properties
func createIPNode(ctx context.Context, session neo4j.SessionWithContext, ipNode IPNode) error {
	_, err := session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		query := `
		MERGE (ip:`
		switch ipNode.NodeType {
		case "Internet":
			query += "IPAddressInternet"
		case "External":
			query += "IPAddressExternal"
		default:
			query += "IPAddress"
		}
		query += ` {address: $ipAddress})
		ON CREATE SET
			ip.networkNickName = $networkNickName,
			ip.compliance = $compliance,
			ip.nodeType = $nodeType,
			ip.destinationPorts = $destinationPorts,
			ip.createdAt = datetime(),
			ip.name = $ipAddress`

		// Modify the query above to include dynamic fields if needed
		if len(ipNode.DynamicFields) > 0 {
			for key := range ipNode.DynamicFields {
				query += fmt.Sprintf(", ip.%s = $%s", key, key)
			}
		}

		query += `
		ON MATCH SET
			ip.networkNickName = $networkNickName,
			ip.compliance = $compliance,
			ip.nodeType = $nodeType,
			ip.destinationPorts = $destinationPorts,
			ip.updatedAt = datetime(),
			ip.name = $ipAddress`

		// Modify the query above to include dynamic fields if needed
		if len(ipNode.DynamicFields) > 0 {
			for key := range ipNode.DynamicFields {
				query += fmt.Sprintf(", ip.%s = $%s", key, key)
			}
		}

		query += `
		RETURN ip
		`

		parameters := map[string]interface{}{
			"ipAddress":        ipNode.IPAddress,
			"networkNickName":  ipNode.NetworkNickName,
			"compliance":       ipNode.Compliance,
			"destinationPorts": ipNode.DestinationPorts,
			"nodeType":         ipNode.NodeType,
		}

		// Modify the parameters to include dynamic fields if needed
		if len(ipNode.DynamicFields) > 0 {
			for key, value := range ipNode.DynamicFields {
				parameters[key] = value
			}
		}

		_, err := tx.Run(ctx, query, parameters)
		return nil, err
	})
	return err
}

// createConnection creates a relationship between source and destination IPs
func createConnection(ctx context.Context, session neo4j.SessionWithContext, conn NewConnection) error {
	_, err := session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		query := `
		MATCH (source:`
		switch conn.SourceNodeType {
		case "Internet":
			query += "IPAddressInternet"
		case "External":
			query += "IPAddressExternal"
		default:
			query += "IPAddress"
		}
		query += ` {address: $sourceIP})
		MATCH (dest:`
		switch conn.DestinationNodeType {
		case "Internet":
			query += "IPAddressInternet"
		case "External":
			query += "IPAddressExternal"
		default:
			query += "IPAddress"
		}
		query += ` {address: $destIP})
		SET dest.destinationPorts =
    		CASE
    		WHEN dest.destinationPorts IS NULL THEN [$destinationPort]
    		WHEN NOT $destinationPort IN dest.destinationPorts THEN dest.destinationPorts + [$destinationPort]
    		ELSE dest.destinationPorts
    		END
		MERGE (source)-[r:TO {
    		source: $sourceIP,
    		dest: $destIP,
			destinationPort: $destinationPort,
    		protocol: $protocol,
			ruleName: $ruleName,
			connectionStatus: $connectionStatus
		}]->(dest)
		ON CREATE SET r.createdAt = datetime()
		ON MATCH SET r.updatedAt = datetime()
		RETURN r
		`

		//if conn.SourceNodeType == "External" || conn.SourceNodeType == "Internet" {
		//	fmt.Println(query)
		//}

		parameters := map[string]interface{}{
			"sourceIP":         conn.SourceIP,
			"destIP":           conn.DestinationIP,
			"protocol":         conn.Protocol,
			"destinationPort":  conn.DestinationPort,
			"ruleName":         conn.RuleName,
			"connectionStatus": conn.ConnectionStatus,
		}

		_, err := tx.Run(ctx, query, parameters)
		return nil, err
	})
	return err
}

// UpdateNetworkNickname updates the network nickname for all IP addresses that match subnets listed in the CSV file
func updateKeyValue(ctx context.Context, session neo4j.SessionWithContext, csvPath string) error {
	fmt.Println("\n\nUpdating key values from CSV file:", csvPath)
	// Open the CSV file
	file, err := os.Open(csvPath)
	if err != nil {
		return fmt.Errorf("could not open CSV file: %v", err)
	}
	defer file.Close()

	// Read the CSV file
	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return fmt.Errorf("could not read CSV file: %v", err)
	}

	// Process each record (skip header row if present)
	for i, record := range records {
		// Skip header row if it exists
		if i == 0 && (strings.ToLower(record[0]) == "startsWith" || strings.ToLower(record[1]) == "key" || strings.ToLower(record[2]) == "value") {
			//fmt.Println("Skipping header row")
			continue
		}

		if len(record) < 3 {
			continue // Skip rows without both subnet and nickname
		}

		startsWith := record[0]
		key := record[1]
		value := record[2]

		// Update all IP addresses that start with this subnet
		_, err := session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
			// Come back and modify this if external and nat networks need a nickname...
			query := `
            MATCH (ip:IPAddress)
            WHERE ip.address STARTS WITH $subnet
            SET ip.`
			query += key
			query += ` = $value,
                ip.updatedAt = datetime()
            RETURN count(ip) as updatedCount
            `

			parameters := map[string]interface{}{
				"subnet": startsWith,
				"key":    key,
				"value":  value,
			}

			result, err := tx.Run(ctx, query, parameters)
			if err != nil {
				return nil, err
			}

			if result.Next(ctx) {
				count := result.Record().Values[0].(int64)
				fmt.Printf("\nUpdated %d IPs in subnet %s with key '%s' and value '%s'\n", count, startsWith, key, value)
			}
			return nil, nil
		})

		if err != nil {
			log.Printf("Error updating subnet %s: %v", startsWith, err)
			continue
		}
	}

	return nil
}

func IPInSubnet(ipStr, cidrStr string) bool {
	// Parse the IP address
	ip := net.ParseIP(ipStr)
	if ip == nil {
		return false
	}

	// Parse the CIDR notation
	_, subnet, err := net.ParseCIDR(cidrStr)
	if err != nil {
		return false
	}

	// Check if the IP is within the subnet
	return subnet.Contains(ip)
}

func CalcNodeType(ipNode IPNode, config Configuration) IPNode {
	// Check if the IP address is in any of the external networks
	for _, externalNetwork := range config.ExternalNetworks {
		if IPInSubnet(ipNode.IPAddress, externalNetwork) {
			ipNode.NodeType = "External"
		}
	}

	// Check if the IP address is in any of the internal networks
	for _, internalNetwork := range config.InternalNetworks {
		if IPInSubnet(ipNode.IPAddress, internalNetwork) {
			ipNode.NodeType = "Internal"
		}
	}

	// Check if the IP address is in any of the internal networks
	for _, n := range config.IgnoredNetworks {
		if IPInSubnet(ipNode.IPAddress, n) {
			ipNode.NodeType = "Ignored"
		}
	}

	if ipNode.NodeType != "External" && ipNode.NodeType != "Internal" && ipNode.NodeType != "Ignored" {
		ipNode.NodeType = "Internet"
	}

	return ipNode // Default type if no match found
}

func main() {
	ConfigPtr := flag.String("config", "config.json", "Configuration file to load for the proxy")
	CSVPtr := flag.String("csv", "", "CSV file to load into the database")
	KeyUpdatePtr := flag.String("keyupdate", "", "CSV file to update the specified key in the file")

	// Custom help message
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
		fmt.Fprintln(os.Stderr, "\nFor testing I use neo4j in a docker container, I pull the latest by running: ")
		fmt.Fprintln(os.Stderr, "   docker pull neo4j:latest")
		fmt.Fprintln(os.Stderr, "\nIn the directory where you execute the next docker command create the following directories:")
		fmt.Fprintln(os.Stderr, "   mkdir neo4j_data neo4j_logs")
		fmt.Fprintln(os.Stderr, "\nTo setup and run the container, I use the following commands in the directory where the directories were created:")
		fmt.Fprintln(os.Stderr, "   docker run -d --name neo4j --rm -p 127.0.0.1:7687:7687 -p 127.0.0.1:7474:7474 -v $(pwd)/neo4j_data:/data -v $(pwd)/neo4j_logs:/logs -e NEO4J_AUTH=neo4j/l0st1nSpac3 neo4j:latest")
		fmt.Fprintln(os.Stderr, "\nYou can then access the Neo4j browser at http://localhost:7474")
		fmt.Fprintln(os.Stderr, "\nOptions:")
		flag.PrintDefaults()
	}
	flag.Parse()

	// Load the Configuration file
	var config Configuration
	configFile := *ConfigPtr
	log.Println("Loading the following config file: " + configFile + "\n")
	if err := config.LoadConfig(configFile); err != nil {
		//fmt.Println("Could not load the configuration file, creating a new default config.json")
		config.CreateConfig(configFile)
		log.Fatalf("Modify the %s file to customize how the tool functions: %v\n", configFile, err)
	}

	// Create Neo4j driver
	driver, err := neo4j.NewDriverWithContext(config.Neo4jURI, neo4j.BasicAuth(config.Neo4jUsername, config.Neo4jPassword, ""))
	//driver, err := neo4j.NewDriver(uri, neo4j.BasicAuth(username, password, ""))
	if err != nil {
		log.Fatalf("Error creating Neo4j driver: %v", err)
	}
	// Create the context for the driver
	ctx := context.Background()
	// Ensure the driver is closed when done
	defer func() {
		if err := driver.Close(ctx); err != nil {
			log.Fatalf("Error closing Neo4j driver: %v", err)
		}
	}()
	defer driver.Close(ctx)

	// Read a csv file if provided with the -csv flag
	var csvFile CSVFileStruct
	var nodes []IPNode
	var connections []NewConnection
	if *CSVPtr != "" {
		fmt.Println("Loading CSV file:", *CSVPtr)
		if err := csvFile.LoadCSVFile(*CSVPtr); err != nil {
			log.Fatalf("Error loading CSV file: %v", err)
		}
		fmt.Printf("CSV file loaded successfully. Number of rows: %d. Processing...\n", len(csvFile.Rows))
		for _, row := range csvFile.Rows {

			//destPort, _ := strconv.Atoi(row.DestinationPort)
			//fmt.Printf("Processing row: SourceIP=%s, DestinationIP=%s, Protocol=%s, DestinationPort=%s\n", row.SourceIP, row.DestinationIP, row.Protocol, row.DestinationPort)
			ipConn := NewConnection{
				SourceIP:         row.SourceIP,
				DestinationIP:    row.DestinationIP,
				DestinationPort:  row.DestinationPort,
				Protocol:         row.Protocol,
				RuleName:         row.RuleName,
				ConnectionStatus: row.ConnectionStatus,
			}

			srcNode := IPNode{
				IPAddress:        row.SourceIP,
				NetworkNickName:  "Unknown", // Default value, can be modified later
				Compliance:       "Unknown", // Default value, can be modified later
				NodeType:         "Unknown",
				DestinationPorts: []int{},
				DynamicFields:    config.DynamicFields,
			}

			// Calculate the node type and color based on the IP address
			srcNode = CalcNodeType(srcNode, config)
			ipConn.SourceNodeType = srcNode.NodeType

			destNode := IPNode{
				IPAddress:        row.DestinationIP,
				NetworkNickName:  "Unknown", // Default value, can be modified later
				Compliance:       "Unknown", // Default value, can be modified later
				NodeType:         "Unknown",
				DestinationPorts: []int{},
				DynamicFields:    config.DynamicFields,
			}

			// Calculate the node type and color based on the IP address
			destNode = CalcNodeType(destNode, config)
			ipConn.DestinationNodeType = destNode.NodeType

			if destNode.NodeType == "Ignored" || srcNode.NodeType == "Ignored" {
				// Skip connections where either source or destination is in an ignored network
				fmt.Printf("Skipping connection from %s to %s as one of the nodes is in an ignored network\n", srcNode.IPAddress, destNode.IPAddress)
				continue
			} else {
				// Add the ipConn to the connections slice if it does not already exist
				exists := false
				for _, conn := range connections {
					if conn.SourceIP == ipConn.SourceIP && conn.DestinationIP == ipConn.DestinationIP &&
						conn.Protocol == ipConn.Protocol && conn.DestinationPort == ipConn.DestinationPort {
						exists = true
						break
					}
				}
				if !exists {
					connections = append(connections, ipConn)
				}

				// Add the srcNode and destNode to the nodes slice if they do not already exist
				nodeExists := false
				for _, node := range nodes {
					if node.IPAddress == srcNode.IPAddress {
						nodeExists = true
						break
					}
				}
				if !nodeExists {
					nodes = append(nodes, srcNode)
				}
				nodeExists = false
				for _, node := range nodes {
					if node.IPAddress == destNode.IPAddress {
						nodeExists = true
						break
					}
				}
				if !nodeExists {
					nodes = append(nodes, destNode)
				}
			}
		}

		session := driver.NewSession(ctx, neo4j.SessionConfig{})
		if session == nil {
			log.Fatalf("Error creating Neo4j session")
		}
		defer session.Close(ctx)

		var wg sync.WaitGroup
		semaphore := make(chan struct{}, config.ConcurrentProcessing)
		fmt.Printf("Creating IP Nodes... Total Nodes: %d\n", len(nodes))
		for _, node := range nodes {
			wg.Add(1)
			go func(node IPNode) {
				defer wg.Done()
				semaphore <- struct{}{}        // Acquire a semaphore slot
				defer func() { <-semaphore }() // Release the semaphore slot
				//fmt.Printf("Creating IP Node: %s\n", node.IPAddress)

				err := createIPNode(ctx, session, node)
				if err != nil {
					log.Printf("Error creating IP node %s: %v", node.IPAddress, err)
				}
			}(node)
		}
		wg.Wait()

		//session.Close(ctx)
		semaphore = make(chan struct{}, config.ConcurrentProcessing)
		fmt.Println("Creating Connections... Total Connections:", len(connections))
		for _, conn := range connections {
			wg.Add(1)
			go func(conn NewConnection) {
				defer wg.Done()
				semaphore <- struct{}{}        // Acquire a semaphore slot
				defer func() { <-semaphore }() // Release the semaphore slot
				err = createConnection(ctx, session, conn)
				if err != nil {
					log.Printf("Error creating connection from %s to %s with protocol %s and destination port %s:\n%v", conn.SourceIP, conn.DestinationIP, conn.Protocol, conn.DestinationPort, err)
				}
			}(conn)
		}
		wg.Wait()
		//session.Close(ctx)

		// Create the hopCount from the internet to the internal nodes
		fmt.Println("Calculating minimum hop counts from Internet to Internal nodes...")
		_, err = session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
			query := `
			MATCH path = (internet:IPAddressInternet)-[*1..]->(external:IPAddressExternal)-[*1..]->(lastNode:IPAddress)
			WHERE lastNode.hopLength IS NULL OR length(path) < lastNode.hopLength
  				SET lastNode.hopLength = length(path)
  				SET lastNode.hopSource = internet.address  // Stores source IP if needed
			RETURN
  				lastNode.address AS internal_ip,
  				lastNode.hopLength AS hops,
  				[n IN nodes(path) | n.address] AS full_path
			`

			_, err := tx.Run(ctx, query, nil)
			if err != nil {
				return nil, err
			}

			/** Removed the stdout of the results to avoid cluttering the console
			result, err := tx.Run(ctx, query, nil)
			if err != nil {
				return nil, err
			}
			for result.Next(ctx) {
				record := result.Record()
				internalIP := record.Values[0].(string)
				hops := record.Values[1].(int64)
				fullPath := record.Values[2].([]any)
				fmt.Printf("Internal IP: %s, Hops: %d, Full Path: %v\n", internalIP, hops, fullPath)
			}
			**/
			return nil, nil
		})
		if err != nil {
			log.Fatalf("Error calculating hop counts: %v", err)
		}
		session.Close(ctx)
	}

	if len(csvFile.Rows) == 0 || *CSVPtr == "" {
		fmt.Println("No CSV file provided or no rows found in the CSV file.")
		fmt.Println("You can provide a CSV file with the -csv flag to load data into Neo4j.")
		fmt.Println("Example CSV format:")
		fmt.Println("source_ip,destination_ip,destination_port,protocol,rule_name,connection_status")
		fmt.Println("Example row:")
		fmt.Println("10.1.1.1,10.2.2.2,80,tcp,AllowHTTP,allowed")
	} else {
		fmt.Println("Data successfully loaded into Neo4j")
	}

	if *KeyUpdatePtr != "" {
		if err := updateKeyValue(ctx, driver.NewSession(ctx, neo4j.SessionConfig{}), *KeyUpdatePtr); err != nil {
			log.Fatalf("Error updating key values from CSV file: %v", err)
		} else {
			fmt.Println("Key values updated successfully from CSV file.")
		}
	}
}
