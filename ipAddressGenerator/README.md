IP Address Record Generator

This Go program generates random source and destination IP Addresses with given protocols and ports.  This was created to generate an import file for the project called "ipAddressRelationships", which uses neo4j to identify relationships between nodes.

## Overview

The program generates random a csv file containing:
- Source IP addresses within a specified subnet
- Destination IP addresses within a specified subnet
- Randomly selected protocols
- Randomly selected destination ports

## Configuration

The program reads from a `config.json` file with the following structure:

```json
{
  "source_subnet": "192.168.1.0/24",
  "destination_subnet": "10.0.0.0/23",
  "destination_ports": [80, 443, 22, 3389],
  "protocols": ["tcp", "udp", "icmp"],
  "count": 100
}
```

## How It Works

1. **Configuration Loading**:
   - Reads and parses the `config.json` file
   - Validates the subnet formats
   - Stores all configuration parameters

2. **Record Generation**:
   - For each requested record (based on the `count` parameter):
     - Generates a random source IP within the source subnet
     - Generates a random destination IP within the destination subnet
     - Selects a random protocol from the configured list
     - Selects a random destination port from the configured list

3. **Output**:
   - Prints each generated record in CSV format to stdout
   - Capture the output as it is generated to stdout

## Example Output

```
192.168.1.123,10.0.45.67,tcp,443
192.168.1.87,10.0.12.34,udp,53
192.168.1.201,10.0.99.1,icmp,0
```

## Building and Running

1. Ensure Go is installed
2. Create a `config.json` file with your desired parameters
3. Run the prep program after modifying the variables:
   ```bash
   ./prep.sh
   ```
4. Run the binary that is created and capture the output
```bash
./ipGen.bin > outputList.csv
```


