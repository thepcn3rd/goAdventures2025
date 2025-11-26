# nmap TLS Enumeration of Ciphers Parser

A Go utility for parsing Nmap XML output from the `ssh2-enum-algos` script and generating a CSV report of SSH algorithms used by scanned hosts.  As part of new PCI 4 requirements, 12.3.3 is about creating, maintaining and managing risk based on a cipher and protocol inventory.  This tool can be used to create an inventory for TLS protocol and ciphers utilized.

Note: I built this tool prior to the SSH Enumeration.  Need to circle back and rewrite this to be more dynamic with the detected ciphers.

## Features

- Parses multiple Nmap XML files containing `ssl-enum-ciphers` script results
- Extracts TLS cipher information including:
  - Protocol Used (TLS 1.0, TLS 1.1, or etc.)
  - Ciphers available for negotiation
- Generates a CSV report showing which protocols and ciphers each host supports
- Preserves host information (IP, hostname, port)

## Usage

1. Run Nmap scans with the `ssl-enum-ciphers` script and save output as XML (Could be a loop over IP Addresses):
```bash
sudo nmap -p 443 -sV --script ssl-enum-ciphers -oX output.xml 10.10.10.0/24
```

2. Place all XML files in an `output/` directory where the binary is located

3. Create the binary for the parser:
```bash
./prep.sh
```

4. Execute the binary
```bash
./parseTLS.bin
```

5. The tool will generate a `ssl_ciphers.csv` file with the results

## Output Format

The CSV output contains:
- Basic host information (IP, hostname, port, protocol, service details)
- Columns for each discovered cipher
- "x" marks when a host supports a particular cipher
- "-" marks when a host doesn't support a cipher

## Requirements

- Go 1.16+
- Nmap with `ssl-enum-ciphers` script

## Example Output

The generated CSV will look similar to:

```
IP,Hostname,Port,TLS1_0Supported,TLS1_1Supported,TLS1_2Supported,TLS1_3Supported,TLS_AKE_WITH_AES_128_GCM_SHA256,TLS_AKE_WITH_AES_256_GCM_SHA384,TLS_AKE_WITH_CHACHA20_POLY1305_SHA256,TLS_DHE_RSA_WITH_3DES_EDE_CBC_SHA,TLS_DHE_RSA_WITH_AES_128_CBC_SHA256,TLS_DHE_RSA_WITH_AES_128_CBC_SHA,TLS_DHE_RSA_WITH_AES_128_GCM_SHA256,TLS_DHE_RSA_WITH_AES_256_CBC_SHA256,TLS_DHE_RSA_WITH_AES_256_CBC_SHA,TLS_DHE_RSA_WITH_AES_256_GCM_SHA384,TLS_DHE_RSA_WITH_CAMELLIA_128_CBC_SHA,TLS_DHE_RSA_WITH_CAMELLIA_256_CBC_SHA,TLS_DHE_RSA_WITH_CHACHA20_POLY1305_SHA256,TLS_DHE_RSA_WITH_SEED_CBC_SHA,TLS_ECDHE_RSA_WITH_3DES_EDE_CBC_SHA,TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA256,TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA,TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA384,TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA,TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,TLS_ECDHE_RSA_WITH_ARIA_128_GCM_SHA256,TLS_ECDHE_RSA_WITH_ARIA_256_GCM_SHA384,TLS_ECDHE_RSA_WITH_CAMELLIA_128_CBC_SHA256,TLS_ECDHE_RSA_WITH_CAMELLIA_256_CBC_SHA384,TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305_SHA256,TLS_ECDHE_RSA_WITH_RC4_128_SHA,TLS_RSA_WITH_3DES_EDE_CBC_SHA,TLS_RSA_WITH_AES_128_CBC_SHA256,TLS_RSA_WITH_AES_128_CBC_SHA,TLS_RSA_WITH_AES_128_CCM_8,TLS_RSA_WITH_AES_128_CCM,TLS_RSA_WITH_AES_128_GCM_SHA256,TLS_RSA_WITH_AES_256_CBC_SHA256,TLS_RSA_WITH_AES_256_CBC_SHA,TLS_RSA_WITH_AES_256_CCM_8,TLS_RSA_WITH_AES_256_CCM,TLS_RSA_WITH_AES_256_GCM_SHA384,TLS_RSA_WITH_ARIA_128_GCM_SHA256,TLS_RSA_WITH_ARIA_256_GCM_SHA384,TLS_RSA_WITH_CAMELLIA_128_CBC_SHA256,TLS_RSA_WITH_CAMELLIA_128_CBC_SHA,TLS_RSA_WITH_CAMELLIA_256_CBC_SHA256,TLS_RSA_WITH_CAMELLIA_256_CBC_SHA,TLS_RSA_WITH_IDEA_CBC_SHA,TLS_RSA_WITH_RC4_128_MD5,TLS_RSA_WITH_RC4_128_SHA,TLS_RSA_WITH_SEED_CBC_SHA,TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA,TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA256,TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,TLS_ECDHE_ECDSA_WITH_AES_256_CBC_SHA384,TLS_ECDHE_ECDSA_WITH_AES_256_CBC_SHA,TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305_SHA256
74.6.231.21,yahoo.com,443/tcp http-proxy - Apache Traffic Server,true,true,true,true,A,A,A,,,,,,,,,,,,,A,A,A,A,A,A,,,,,A,,,A,,,,A,A,A,,,A,,,,,,,,,,,A,A,A,A,A,A,A
```