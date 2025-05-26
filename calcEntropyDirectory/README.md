# File Entropy Analyzer

A Go program that calculates file entropy, hashes, and other metadata for files in a specified directory and does recursion of the directory.

## Description

This tool analyzes files in a given directory and calculates:
- File entropy (Shannon entropy)
- Cryptographic hashes (MD5, SHA1, SHA256)
- File metadata (size, permissions, last modified)
- MIME type detection
- Chunk-based entropy analysis (configurable chunk size)

Entropy analysis is particularly useful for detecting potentially suspicious files, as encrypted or compressed data often exhibits high entropy values.

## Features

- Recursive directory scanning with depth control
- Configurable chunk size for entropy analysis
- Multiple output formats (JSON, CSV, or both)
- File size limits for chunk analysis
- Cross-platform compatibility (Linux/Windows)

## Usage

Create the binary for the parser:
```bash
./prep.sh
```

Execute the binary
```bash
./calcEntropyDirectory.bin -d . -o output -depth 3
```



```
Usage: calcEntropy [options]

Options:
  -d string
        Calculate Entropy of Files in Specified Directory
  -o string
        Save output to this file, extension depends on the format selected (default "output")
  -format string
        Output in CSV and JSON Format or specify (default "both")
  -size int
        Size of Chunk Evaluated (default 256)
  -depth int
        Maximum recursion depth (0 for current directory only, -1 for unlimited) (default -1)
  -maxsize int
        Maximum size of file in MB to evaluate chunks (default 10)
  -debug
        Enable Debug Information, creates a debug file
```

### Example Commands

```bash
# Analyze current directory with default settings
./calcEntropy -d . -o analysis

# Analyze specific directory with custom chunk size and max depth
./calcEntropy -d /path/to/files -size 512 -depth 2 -o results

# Output only in CSV format
./calcEntropy -d /path/to/files -format csv -o report
```

## Output Formats

The program generates detailed reports including:
- File metadata (name, path, size, permissions)
- Cryptographic hashes
- MIME type
- Overall file entropy and rating (Low/Medium/High)
- Chunk-based entropy analysis (if enabled)

### JSON Output Example

```json
{
   "BaseDir": "/path/to/files",
   "ChunkSize": 256,
   "MaxDepth": -1,
   "Created": "15:04:05",
   "FileList": ["file1.txt", "subdir/file2.exe"],
   "EntropyFiles": [
      {
         "Name": "file1.txt",
         "FilePath": "/path/to/files/file1.txt",
         "FileSize": 1024,
         "Permissions": "-rw-r--r--",
         "LastModified": "2023-01-01 12:00:00 +0000 UTC",
         "MD5": "d41d8cd98f00b204e9800998ecf8427e",
         "SHA1": "da39a3ee5e6b4b0d3255bfef95601890afd80709",
         "SHA256": "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
         "MIMEType": "text/plain",
         "Entropy": 3.1415,
         "EntropyRating": "Low",
         "ChunksEvaluated": true,
         "ChunkSize": 256,
         "TotalChunks": 4,
         "LowEntropyChunks": 4,
         "MediumEntropyChunks": 0,
         "HighEntropyChunks": 0
      }
   ]
}
```

### CSV Output Columns

| Column | Description |
|--------|-------------|
| Name | File name |
| FilePath | Full file path |
| FileSize | File size in bytes |
| Permissions | File permissions |
| LastModified | Last modification timestamp |
| MD5 | MD5 hash |
| SHA1 | SHA1 hash |
| SHA256 | SHA256 hash |
| MIMEType | Detected MIME type |
| Entropy | Calculated entropy value |
| EntropyRating | Low/Medium/High rating |
| ChunksEvaluated | Whether chunks were analyzed |
| ChunkSize | Size of chunks analyzed |
| TotalChunks | Total number of chunks |
| LowEntropyChunks | Count of low entropy chunks |
| MediumEntropyChunks | Count of medium entropy chunks |
| HighEntropyChunks | Count of high entropy chunks |

## Entropy Interpretation

- **Low Entropy (≤5.0)**: Indicates more orderly or non-random data (typical for plain text, uncompressed files)
- **Medium Entropy (5.0-6.5)**: Intermediate randomness
- **High Entropy (>6.5)**: Indicates more random data (typical for encrypted, compressed, or binary files)

