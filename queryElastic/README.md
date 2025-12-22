# Elastic Search API Query

This document describes the setup of a simple client using the go Elasticsearch API to query my security onion instance.  

---

## Table of Contents

- [Problem Statement](#problem-statement)
- [Solution](#solution)
- [Compiling and Configuration](#compiling-and-configuration)
- [Configuration File](#configuration-file)
- [Execution Instructions](#execution-instructions)
- [Troubleshooting](#troubleshooting)

---

## Problem Statement

The DNS Server that I previously created, generates a syslog that is sent to the Security Onion instance that I am running.  To analyze the queries that are logged across a time frame of 1 week I created the following client.  I was faced with these 2 challenges:

Challenges:
1. Security Onion Configuration: To run this script I had to allow communication from my development instance and where I need to run my script to talk to port 9200.  This required configuration for my IP Address to communicate.
2. Go ElasticSearch Module: This had a level of difficulty to identify how to use the ElasticSearch Module and then to implement it.  The struct is also provided for future development if necessary.
3. Analysis of the Data: The pfsense and the security onion is conducting security analysis of the URLs that pass through, however I wanted to look closer and do some statistical analysis on the information.  That is why I wrote this specific script.

---

## Solution

The golang program is specific to parsing out the Real Message that is returned in the JSON.  The JSON is formulated into a struct.  Then the regular expression is applied to pull out only the DNS that was forwarded to my upstream query, which is my pfsense firewall.  

![Army of Scorpions](/picts/armyScorpions.png)

## Compiling and Configuration

A `prep.sh` script is provided to streamline the compilation process. Ensure the `GOPATH` environment variable is correctly set before execution:

```bash
go env -w GOPATH="/home/thepcn3rd/go/workspaces/2025/queryElastic"
```

Run the script with the following command:

```bash
. prep.sh
```

You can modify the script to change binary names or enable the creation of Windows binaries by uncommenting relevant lines.

## Configuration File

The `config.json` file contains parameters for network-specific settings. Below is a sample configuration:

```json
{
	"elasticURL": "https://10.10.10.10:9200",
	"username": "username@email.local",
	"password": "longpasswordlongpassword",
	"elasticSettings": {
		"index": "logs-syslog*",
		"keywords": "golang dns server - forwarding query",
		"lookbackTimeDays": -7,
		"pageSize": 100,
		"maxPages": 5,
		"regexpRealMessage": "`for\\s([^\\s]+)\\sto`",
		"_comment": "Currently the regex applies to the RealMessage returned by the elastic search query conducted. Remember to escape the backslash!"
	}
}
```

#### Configuration Details

* Elastic URL: This is the listening IP Address and Port of the Elastic Server.  If it uses a self signed certificate it ignores and does not validate
* Username: Elastic Username for the Query
* Password: Elastic Password for the Query
* Elastic Settings
	* Index - The index to query in Elastic
	* Keywords - The keywords to identify in the real message
	* Look Back - The number of days to look back, this needs to be a negative number
	* Page Size - The number of results returned in a given page of information
	* Max Pages - You can set this to truncate the number of pages returned
	* Regular Expression - The Regular Expression applied to the message to extract exactly what you need in it.  This case was to extract the domain name being queried

---

## Execution Instructions

To execute the binary and specify a custom configuration file:

```bash
./queryElastic -config config.json
```

### Command-Line Usage

```txt
Usage of ./queryElastic.bin:
  -config string
    	Configuration file to load for the proxy (default "config.json")
```

---

## Troubleshooting

1. **Binary Not Found:** Ensure `prep.sh` has executed successfully.
2. Make sure to escape any backslashes in the config.json file that is utilized
3. Verify the index exists in your instance of Elastic Search
4. Verify the firewall allows the connection to Elastic Search
5. Setup minimal permissions for the access to Elastic Search

---

## License

This project is licensed under the [MIT License](/LICENSE.md).