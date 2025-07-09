# goHashScanner

goHashScanner is a Go program designed to scan files or directories for various types of hashes. It uses regular expressions to identify potential hash values within text files and can be customized to ignore specific patterns or files. The program is particularly useful for security professionals who need to identify sensitive information, such as password hashes or patterns within code, file systems, or log files.

## Features

- **Customizable Hash Detection**: The program uses a JSON file (`customPrototypes.json`) to define the types of strings or hashes to search for. Each hash/string type has an associated regular expression and can be enabled or disabled.
- **Exclusion Lists**: The program supports an exclusion list (`exclusions.json`) to ignore specific strings, files, or patterns that might otherwise be flagged as hashes.
- **File and Directory Scanning**: The program can scan individual files or recursively scan directories for hashes.
- **Binary File Detection**: The program skips binary files by checking the first 512 bytes for non-printable characters unless the setting is modified in the exclusions.json file.
- **Max File Size**: The program can be configured to skip files larger than a specified size with a setting in the exclusions.json file.


![Scorpion Soldier Scanning](/picts/scorpionSoldierScanning.png)

## Usage

A prep.sh file is included with the program and can be used to download the modules needed and compile the program:

```bash
./prep.sh
```

After compiling it creates the following binary to execute:
```bash
./hashID -h
```
### Flags

- `-p`: Specifies the path to the prototypes JSON file (default: `customPrototypes.json`).
- `-e`: Specifies the path to the exclusions JSON file (default: `exclusions.json`).
- `-s`: Reads input from stdin to search for a match.
- `-f`: Specifies a file to search for hashes.
- `-d`: Specifies a directory to recursively search for hashes.
- `-t`: Creates a template `templatePrototypes.json` file for complete customization.

### Examples

1. **Scan a Single File**:
```bash
./hashID.bin -f example.txt
```
   This command will scan `example.txt` for any hashes defined in `customPrototypes.json`.

2. **Scan a Directory**:
```bash
./hashID.bin -d /path/to/directory
```
   This command will recursively scan all files in the specified directory for hashes.

3. **Read from Standard Input**:
```bash
echo "some text with a hash 5f4dcc3b5aa765d61d8327deb882cf99" | ./hashID.bin -s
   ```
   This command will read input from stdin and search for hashes.

4. **Create a Template Prototypes File**:
```bash
./hashID -t
```
   This command will create a `templatePrototypes.json` file that you can customize to define your own hash detection rules.

5. **Custom Prototypes and Exclusions**:
```bash
./hashID -p customPrototypes.json -e exclusions.json -f example.txt
```
   This command will use `customPrototypes.json` for hash detection rules and `exclusions.json` for exclusion rules while scanning `example.txt`.

## Configuration

### Prototypes JSON (`customPrototypes.json`)

The `customPrototypes.json` file defines the types of hashes the program will search for. Each hash type includes a regular expression, a name, and optional metadata such as John the Ripper and Hashcat modes.

Example:
```json
{
  "prototypes": [
    {
      "regex": "^[a-f0-9]{32}$",
      "newRegex": "[a-f0-9]{32}",
      "enabled": true,
      "notes": "MD5 Hash",
      "modes": [
        {
          "john": "raw-md5",
          "hashcat": 0,
          "extended": false,
          "name": "MD5"
        }
      ]
    }
  ]
}
```

### Exclusions JSON (`exclusions.json`)

The `exclusions.json` file defines strings, patterns, and files that should be ignored during the scan.

Example:
```json
{
  "maxFileSize": 1048576,
  "binaryCheck": 512,
  "matchStrings": [
    "ignoreThisString"
  ],
  "regexs": [
    "^ignoreThisPattern$"
  ],
  "files": [
    "ignoreThisFile.txt"
  ]
}
```

## Dependencies

- **Go Modules**: The program uses the `slices` package, which is available in Go 1.18 or later.  (This code can be modified to not use slices...)
- **External Libraries**: The program uses `github.com/thepcn3rd/goAdvsCommonFunctions` for common utility functions.

## Conclusion

The goHashScanner is a versatile tool for identifying hashes within text files. Its customizable nature allows it to be adapted to various use cases, making it a valuable tool for security audits and code reviews.
