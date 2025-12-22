# Neo4j IP Address Relationship Mapper

This tool is designed to create nodes and relationships in Neo4j based on output provided containing a source IP address, destination IP address, protocol and destination port.  The nodes can have additional properties added to them to enrich the data queried about a node.  The power of the tool is to identify relationships between IP Addresses that you did not recognize and paths that could be analyzed for operational or security purposes.

## Overview

The application provides functionality to:
- Create nodes for IP addresses with various properties
- Establish relationships between IP addresses with connection details
- Update node properties in bulk from CSV files
- Query complex network relationships using Neo4j's powerful graph traversal capabilities

## Practical Application

1. **Network Path Analysis**:
   - Identify critical communication paths
   - Detect unusually long connection chains
   - Identify paths from the internet through to internal nodes

2. **Security Investigations**:
   - Trace attack paths through networks
   - Identify systems with non-compliant configurations
   - Identify risk of systems based on where they are at from the internet

3. **Infrastructure Planning**:
   - Visualize network topology
   - Plan segmentation strategies

Below in the picture you can see in the node details that the dynamic fields of VulnScan and Agent are created as specified in the config.json file.  These are properties that you can populate manually.
![Node Details](/picts/nodeDetails.png)


The picture below shows how you can analyze paths from a host connecting to an external IP externally then NATed to an internal host and the path relationships the internal host has.
![Node Paths](/picts/nodePaths.png)

A hopCount property is generated for each node that is in a path connected to something on the internet.  The hopCount is the shortest hop count to an internet resource from the internet. The hopSource is the external IP that it is closest to based on hopCount.
![Hop Count](/picts/hopCount.png)

## Query Examples

After loading data, you can query Neo4j directly (via browser at http://localhost:7474):

1. **View all relationships**:
   ```cypher
   MATCH (n)-[r]->(m) RETURN n, r, m
   ```

2. **Find longest path between IP Address Nodes**:
   ```cypher
   MATCH path = (start:IPAddress)-[:TO*]->(end:IPAddress)
   WHERE start <> end
   RETURN path, length(path) AS pathLength
   ORDER BY pathLength DESC
   LIMIT 1
   ```

3. **Filter by protocol**:
   ```cypher
   MATCH path = (start:IPAddress)-[rels:TO*]->(end:IPAddress)
   WHERE start <> end AND ALL(r IN rels WHERE r.protocol = 'tcp')
   RETURN path, length(path) AS pathLength
   ORDER BY pathLength DESC
   ```

4. **Find paths between specific IPs**:
   ```cypher
   MATCH path = (start:IPAddress)-[:TO*]->(end:IPAddress)
   WHERE start.address = "10.0.0.1" AND end.address = "10.0.0.5"
   RETURN path
   ```

5. Find paths between a specific IP and a subnet:
```cypher
MATCH path = (start:IPAddress)-[:TO*]->(end:IPAddress)
WHERE start <> end AND start.address = "10.0.0.117" AND end.address STARTS WITH "10.4.25."
RETURN path, length(path) AS pathLength
ORDER BY pathLength DESC
```

6. Find the paths that start with a specific IP and with a particular protocol
```cypher
MATCH path = (start:IPAddress)-[rels:TO*]->(end:IPAddress)
WHERE start <> end AND start.address = "10.0.0.117" AND ALL(r IN rels WHERE r.protocol = 'tcp')
RETURN path, length(path) AS pathLength
ORDER BY pathLength DESC
```

7. Find the paths from the internet --> External IP --> NAT --> Internal IP Addresses
```cypher
MATCH path = (internet:IPAddressInternet)-[:TO*1..]->(external:IPAddressExternal)-[:TO*1..]->(internal:IPAddress)
RETURN path
```

8. Find the longest path through internal IP Address Nodes
```cypher
MATCH path = (start:IPAddress)-[rels:TO*]->(end:IPAddress)
WHERE start <> end AND ALL(r IN rels WHERE r.protocol = 'tcp') AND length(path) > 10
RETURN path
LIMIT 5
```



## Installation

### Prerequisites
- Docker (for running Neo4j)
- Go 1.20+ (to build the tool)

### Neo4j Setup with Docker 
```bash
# Create directories for Neo4j data
mkdir neo4j_data neo4j_logs

# Run Neo4j container
docker run -d --name neo4j --rm -p 7687:7687 -p 7474:7474 \
  -v $(pwd)/neo4j_data:/data -v $(pwd)/neo4j_logs:/logs \
  -e NEO4J_AUTH=neo4j/l0st1nSpac3 neo4j:latest
```

### Build the Tool
Modify the necessary variables in the prep.sh file to generate the ipMap.bin binary
```bash
./prep.sh
```

## Usage

### Basic Commands
```bash
# Load data from CSV and specify the config to use
./ipMap.bin -csv test.csv -config config.json

# Update node properties
./ipMap.bin -keyupdate updateKey.csv -config.json
```

### Example config.json File

An example config.json file is provided among the files and is shown below

```json
{
    "neo4juri": "bolt://localhost:7687",
    "username": "neo4j",
    "password": "l0st1nSpac3",
    "concurrentProcessing": 10,
    "_note": "When specifying a dynamicField for an IPNode, provide an example of the output whether a string, int, or ... The field needs to be capitalized... Fields below need to match the 1st column in a supporting csv",
    "dynamicFields": {
            "VulnScan": "False",
            "Agent": "Missing"
    },
    "externalNetworks": [
            "145.6.0.0/16",
            "165.6.0.0/16"
    ],
    "internalNetworks": [
            "10.0.0.0/8",
            "172.16.0.0/12",
            "192.168.0.0/16"
    ],
    "ignoredNetworks": [
		    "169.254.0.0/16"
    ],
    "_noteColors": "Set the protocol on an import to 'nat' and the color will be set to the specified below",
    "networkNodeColors": {
            "internet": "Magenta",
            "external": "Red",
            "internal": "Blue"
    }
}
```

Here's a Markdown explanation of your JSON configuration for a README file:

#### Configuration JSON Explanation

This JSON file contains configuration settings for a Neo4j-based application. Below is an explanation of each field:

Neo4j Connection Settings
- `neo4juri`: The connection URI for the Neo4j database  
- `username`: Username for Neo4j authentication  
- `password`: Password for Neo4j authentication  

Processing Settings
- `concurrentProcessing`: Number of concurrent processes to use for importing data 

Dynamic Fields
- `dynamicFields`: Custom fields that can be added to IP nodes  
  - These are fields that can be in the first column in supporting keyvalue CSV files
  - An example is provided in the config.json file above  
    
Network Definitions
- `externalNetworks`: List of CIDR ranges considered external networks
	- Nodes are created based on CIDR Range
	- Nodes are labelled and created as IPAddressExternal
	- If the IP Address is not externally listed and is outside of RFC1918 then it is labelled and created as an IPAddressInternet
- `internalNetworks`: List of CIDR ranges considered internal networks  
- `ignoredNetworks`: If these show up do not create connections or nodes for them
- The networks are important to have listed as they provide path information from the internet to an external IP to a NATed internal IP to other internal IP paths
  
Node Coloring
- `networkNodeColors`: Defines colors for different network types  
  - This does not work as expected and will be removed...
Notes
- `_note`: Important implementation details about dynamic fields
- `_noteColors`: Explanation of how node coloring works (Will be removed)


Below is a picture of a node created in neo4j has the following properties by default:
1. createdAt - Creation Date and Time
2. updatedAt - Last Updated Date and Time
3. address - IP Address as the Node
4. name - Same as the IP Address
5. compliance - Meant to list which compliance frameworks the node needs applied (Like PCI, GLBA, GDPR, etc.)
6. networkNickname - A nickname given to the subnet or zone the IP Address is related to.


### Example CSV Files Formats

**IP Relationship Connections:** This could be an export from a firewall log.  Provided with the files are 2 that are examples of IP Relationship CSV files called,  test.csv and list2025.txt
```
source_ip,destination_ip,protocol,destination_port
10.0.0.1,10.0.0.2,tcp,80
10.0.0.2,10.0.0.3,udp,53
10.0.0.3,10.0.0.4,nat,80 # How to create a NAT that is represented in the DB
```

**Property Updates:** An example file is among the files called updateKey.csv
```
startsWith,key,value
10.0.0,networkNickName,Core Network
10.0.1,compliance,PCI
```

## Neo4j Concepts

### Graph Database Basics
Neo4j is a graph database that stores data as:
- **Nodes**: Represent entities (like IP addresses in this tool)
- **Relationships**: Connect nodes and can have properties (like protocol and port)
- **Properties**: Key-value pairs stored on both nodes and relationships

### Why Use Neo4j for Network Mapping?
- Naturally represents network topologies as graphs
- Efficiently handles complex relationship queries
- Visualizes connections intuitively
- Scales well for interconnected data

## Key Features

1. **IP Node Creation**:
   - Stores IP addresses as nodes with metadata (nickname, compliance status)
   - Supports dynamic fields for custom properties

2. **Connection Mapping**:
   - Creates directed relationships between IPs
   - Tracks protocol and destination port information
   - Maintains timestamps for creation/updates

3. **Bulk Updates**:
   - Updates node properties from CSV files
   - Can target nodes by IP prefix patterns
## Future Enhancements
1. The import process which creates the nodes and the connections is slow, this could be enhanced with go routines or parallel processing. (Completed)
2. Continue to build queries that can be used for practical application of this program
3. Include different nodes and create connections to them from the internet and external IP Addresses (Completed)
4. Include in the configuration the ability to specify the internal and external IP Addresses used by the company (Completed)
5. Provide the ability in an import to specify NATs that are configured (Completed, in the import of a connection place the protocol as NAT)
6. Provide files that can be used to test the features mentioned above of importing connections, creating internet nodes, creating external nodes and complimenting them with dynamic properties. (Completed)