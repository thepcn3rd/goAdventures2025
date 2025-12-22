# IP to ASN Lookup Tool

This tool is designed to look up Autonomous System Number (ASN) information for given IP addresses using the IP to Country + ASN database from [ipinfo.io](https://ipinfo.io/products/free-ip-database). The tool processes the JSON file provided by ipinfo.io, converts IP addresses to their decimal equivalents, and allows users to search for ASN details based on IP addresses.

## Features

- **IP to ASN Lookup**: Given an IP address, the tool can find the corresponding ASN, ASN name, AS domain, country, country name, and continent.
- **Batch Processing**: The tool can process a list of IP addresses from a file.
- **JSON Restructuring**: The tool can restructure the original JSON file from ipinfo.io into a more usable format.

## Prerequisites

- **ipinfo.io Database**: You need to download the IP to Country + ASN JSON file from [ipinfo.io](https://ipinfo.io/products/free-ip-database). This file is required for the tool to function.


![Scorpion Soldier Loot](/picts/scorpionSoldierLoot.png)

## Installation

Created a prep script in the directory to assist in building the Go application:
   ```bash
   # Adjust the path in the script
   ./prep.sh
   ```

## Usage

### Restructuring the JSON File

If you have the original JSON file from ipinfo.io, you can restructure it to a usable format for this program using the following command:

```bash
./addASNInfo.bin -original <path_to_original_json_file>
```

This will create a new file named `restructured.json` in the current directory.

### Looking Up ASN Information

You can look up ASN information for a single IP address or a list of IP addresses.

#### Single IP Address

To look up ASN information for a single IP address, use the `-i` flag:

```bash
./addASNInfo.bin -a <path_to_restructured_json_file> -i
```

The tool will prompt you to enter an IP address, and it will output the corresponding ASN details.

#### Batch Processing

To process a list of IP addresses from a file, use the `-f` flag:

```bash
./addASNInfo.bin -a <path_to_restructured_json_file> -f <path_to_ip_list_file>
```

The tool will read the IP addresses from the file and output the ASN details for each IP address.

### Output Format

The tool outputs the results in CSV format to stdout with the following columns:

- `IP`: The IP address being looked up.
- `ASN`: The Autonomous System Number (ASN).
- `ASNName`: The name of the ASN.
- `ASDomain`: The domain associated with the ASN.
- `Country`: The country code.
- `CountryName`: The name of the country.
- `ContinentName`: The name of the continent.

## Example

```bash
./addASNInfo.bin -a restructured.json -f ip_list.txt | tee output.csv
```

This command will read IP addresses from `ip_list.txt` and output the ASN details for each IP address.

## Acknowledgments

- The IP to Country + ASN database is provided by [ipinfo.io](https://ipinfo.io/products/free-ip-database).

