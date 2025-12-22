# nmap SSH Algorithm Enumeration Parser

A Go utility for parsing Nmap XML output from the `ssh2-enum-algos` script and generating a CSV report of SSH algorithms used by scanned hosts.  As part of new PCI 4 requirements, 12.3.3 is about creating, maintaining and managing risk based on a cipher and protocol inventory.  This tool can be used to create an inventory for SSH algorithms.

## Features

- Parses multiple Nmap XML files containing `ssh2-enum-algos` script results
- Extracts SSH algorithm information including:
  - Key exchange algorithms (kex)
  - Server host key algorithms
  - Encryption algorithms
  - MAC algorithms
  - Compression algorithms
- Generates a CSV report showing which algorithms each host supports
- Preserves host information (IP, hostname, port, service details)

## Usage

1. Run Nmap scans with the `ssh2-enum-algos` script and save output as XML (Could be a loop over IP Addresses):
```bash
sudo nmap -p 22 -sV --script ssh2-enum-algos -oX output.xml 10.10.10.0/24
```

2. Place all XML files in an `output/` directory where the binary is located

3. Create the binary for the parser:
```bash
./prep.sh
```

4. Execute the binary
```bash
./parseXML.bin
```

5. The tool will generate a `ssh_algos.csv` file with the results

## Output Format

The CSV output contains:
- Basic host information (IP, hostname, port, protocol, service details)
- Columns for each discovered algorithm type
- "x" marks when a host supports a particular algorithm
- "-" marks when a host doesn't support an algorithm

## Requirements

- Go 1.16+
- Nmap with `ssh2-enum-algos` script

## Example Output

The generated CSV will look similar to:

```
IP,Hostname,Port,Protocol,ServiceName,ServiceProduct,ServiceVersion,KexAlgos,curve25519-sha256,curve25519-sha256@libssh.org,diffie-hellman-group-exchange-sha256,diffie-hellman-group14-sha256,diffie-hellman-group16-sha512,diffie-hellman-group18-sha512,ecdh-sha2-nistp256,ecdh-sha2-nistp384,ecdh-sha2-nistp521,ext-info-s,kex-strict-s-v00@openssh.com,sntrup761x25519-sha512@openssh.com,ServerHostKeyAlgos,ecdsa-sha2-nistp256,rsa-sha2-256,rsa-sha2-512,ssh-ed25519,EncryptionAlgos,aes128-ctr,aes128-gcm@openssh.com,aes192-ctr,aes256-ctr,aes256-gcm@openssh.com,chacha20-poly1305@openssh.com,MACAlgorithms,hmac-sha1,hmac-sha1-etm@openssh.com,hmac-sha2-256,hmac-sha2-256-etm@openssh.com,hmac-sha2-512,hmac-sha2-512-etm@openssh.com,umac-128-etm@openssh.com,umac-128@openssh.com,umac-64-etm@openssh.com,umac-64@openssh.com,CompressionAlgos,none,zlib@openssh.com
10.10.10.10,i,22,tcp,ssh,OpenSSH,8.9p1 Ubuntu 3ubuntu0.11,11,x,x,x,x,x,x,x,x,x,-,x,x,4,x,x,x,x,6,x,x,x,x,x,x,10,x,x,x,x,x,x,x,x,x,x,2,x,x
10.10.10.11,j,22,tcp,ssh,OpenSSH,9.6p1 Ubuntu 3ubuntu13.9,12,x,x,x,x,x,x,x,x,x,x,x,x,4,x,x,x,x,6,x,x,x,x,x,x,10,x,x,x,x,x,x,x,x,x,x,2,x,x
```