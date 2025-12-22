# calcEntropy - File Entropy Analysis Tool

## Overview

calcEntropy is a command-line tool written in Go that analyzes files by calculating their entropy (measure of randomness) to help identify potentially suspicious content. High entropy sections in files can indicate encrypted or compressed data, which is often used in malware.

This is an update of my previous tool located [here](https://github.com/thepcn3rd/goAdventures/blob/main/projects/calcEntropy/main.go). This tool only evaluates 1 file at a time.  If you need to evaluate files in bulk refer to this [tool](/calcEntropyDirectory/README.md)

## Features

- Calculates overall file entropy (Shannon entropy)
- Analyzes file in chunks to identify high entropy sections
- Provides file metadata and hash values (MD5, SHA1, SHA256)
- Detects MIME type
- On on Linux, includes output from the `file` command
- Color-coded output for easy interpretation
- Configurable chunk size for analysis

## Usage

Create the binary for the parser:
```bash
./prep.sh
```

Execute the binary
```bash
./calcEntropy.bin -f <filename> 
```


## Additional Usage 

```bash
calcEntropy -f <filename> [-s <chunk_size>] [-d]
```

### Options

- `-f`: File to analyze (required)
- `-s`: Size of chunks to evaluate (default: 256 bytes)
- `-d`: Disable output of chunk information (only show summary)

### Information 

- **Entropy values**:
  - Low entropy (≤ 5.0 bits/byte) - indicates more orderly, less random data
  - Medium entropy (5.0 < entropy ≤ 6.5)
  - High entropy (> 6.5 bits/byte) - indicates more random data, potentially encrypted/compressed

- The tool also provides:
  - Basic file information (size, permissions, timestamps)
  - Cryptographic hashes
  - MIME type detection
  - On Linux: output from the `file` command

## Example Output

```
Basic File Information
----------------------
Name: sample.exe
Size: 102400 bytes
Permissions: -rw-r--r--
Last Modified: 2023-01-01 12:34:56 +0000 UTC
MD5: d41d8cd98f00b204e9800998ecf8427e
SHA1: da39a3ee5e6b4b0d3255bfef95601890afd80709
SHA256: e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855
MIME Type: application/octet-stream

Entropy of File: 7.4523 bits/byte (High Entropy)

Chunk size is set to: 256
Total Chunks: 400
Chunks with Low: 12 - Med: 45 - High: 343 Entropy
```

## Background

Entropy analysis is useful for detecting suspicious files because:
- Encrypted or compressed data typically has high entropy
- Malware often uses encryption or packing to hide its payload
- Legitimate files usually have lower, more varied entropy patterns

As referenced in the code comments:
> "Variations in the entropy values in the file might indicate that suspect content is hidden in files. For example, the high entropy values might be an indication that the data is stored encrypted and compressed and the lower values might indicate that at runtime the payload is decrypted and stored in different sections." (IBM)

