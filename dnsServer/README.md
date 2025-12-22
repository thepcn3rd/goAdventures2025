# DNS Server

Simple DNS Server written in Golang to perform local lookups and then forward any additional requests to an upstream DNS server. Designed to be simple and to create logs or send logs via syslog to Elastic inside a Security Onion Solution.

---

## Table of Contents

1. [Problem Statement](#problem-statement)
2. [Partial Solution](#partial-solution)
3. [Compile and Configure the DNS Server](#compile-and-configure-the-dns-server)
4. [Configuration Details](#configuration-details)
5. [Execution Instructions](#execution-instructions)
6. [Troubleshooting](#troubleshooting)
7. [License](#license)

---

## Problem Statement

As I monitored DNS requests on my network using pfSense, I noticed that requests from my wireless network were coming from the NATed IP Address. I needed a DNS server deployed in the wireless network (e.g., `10.17.37.15`) to identify the source devices making requests.

Additionally, due to TLS certificate validation requirements, I needed a DNS server for proper name resolution. While modifying the hosts file was an option, I wanted an automated solution to handle the growing size of my home lab.

---

## Partial Solution

The DNS server works well for internal requests; however, my wireless system proxies DNS requests, preventing identification of the specific client making a request. Despite this limitation, the server successfully supports TLS certificate validation for my SSL reverse proxy in front of the Guacamole server and other systems.

---
![Fiery Red Armored Scorpion](/picts/fieryRedArmorScorpion.png)
## Compile and Configure the DNS Server

### Compiling the DNS Server

A `prep.sh` script is included to simplify the compilation process. Before executing the script, update the `GOPATH` variable to a valid directory on your system:

```bash
go env -w GOPATH="/home/thepcn3rd/go/workspaces/2025/dnsServer"
```

Run the script with the following command:

```bash
. prep.sh
```

You can modify the script to change binary names or enable the creation of Windows binaries by uncommenting relevant lines.

### Configuration File

The `config.json` file contains parameters for network-specific settings. Below is a sample configuration:

```json
{
  "upstreamDNS": "8.8.8.8:53",
  "serverBanner": "Golang DNS Server",
  "syslogOptions": {
    "syslogEnabled": "True",
    "syslogServer": "10.27.20.210:514",
    "syslogOriginName": "dns-server"
  },
  "saveFileOptions": {
    "saveFileEnabled": "True",
    "saveFileBaseName": "dns-server",
    "saveFileExtension": ".log"
  },
  "aRecords": [
    { "aName": "www.4gr8.local.", "ip": "10.27.20.174" },
    { "aName": "guac.4gr8.local.", "ip": "10.27.20.184" },
    { "aName": "kali.4gr8.local.", "ip": "10.27.20.173" }
  ],
  "txtRecords": [
    { "txtName": "stuff.", "txtMessage": "this_is_a_message_that_can_be_shared" }
  ]
}
```

---

## Configuration Details

- **Upstream DNS Server:** Default is Google DNS (`8.8.8.8:53`). Customize it for your network.
- **Server Banner:** Identifies the server instance in logs.
- **Syslog Options:** Configure syslog logging, including server address and origin name.
- **File Logging:** Enable/disable local file logging and define file naming conventions.
- **A Records:**
    1. Supports resolving based on the first word (e.g., `www` resolves to the IP).
    2. Automatically creates PTR records for reverse lookups.
- **TXT Records:** Added for experimentation and learning about DNS security.

---

## Execution Instructions

To execute the binary and specify a custom configuration file:

```bash
./dnsServer -config home.config.json
```

### Usage

```txt
Usage of ./dnsServer:
  -config string
        Configuration file to load for the proxy (default "config.json")
```

---

## Troubleshooting

1. **Binary Not Found:** Ensure `prep.sh` has executed successfully.
2. **Permission Issues:** Check file permissions for binaries and configuration files.
3. **Syslog Not Working:** Verify syslog server address and connectivity.
4. **Name Resolution Fails:** Confirm A records and upstream DNS settings in `config.json`.

---

## License

This project is licensed under the [MIT License](/LICENSE.md).

