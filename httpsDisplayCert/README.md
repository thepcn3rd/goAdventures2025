# SSL/TLS Certificate Scanner

A Go-based tool for scanning SSL/TLS certificates from websites and exporting the results to CSV and JSON

## Features

- **Certificate Analysis**: Extract detailed SSL/TLS certificate information including:
  - Subject and Issuer details
  - Validity periods (Not Before/Not After)
  - DNS names (SAN)
  - Serial numbers
  - Signature and public key algorithms
  - Email addresses and IP addresses

- **Browser Impersonation**: Configurable HTTP headers to bypass bot detection and WAFs
- **Batch Processing**: Support for single URL or a list of URLs from a file
- **Customizable Headers**: Modify or Add request headers via configuration file

## Usage

### Basic Usage

```bash
# Execute the prep.sh script to compile the program
./prep.sh

# Scan a single URL
./displayCert -u https://example.com

# Scan multiple URLs from a file
./displayCert -f urls.txt
```

### Command Line Options

- `-u`: Single URL to scan (default: http://127.0.0.1)
- `-f`: File containing list of URLs to scan
- `-b`: Browser headers configuration file (default: browserHeaders.json)

### Configuration

The tool uses `browserHeaders.json` to configure HTTP headers. A default configuration file will be created automatically on first run if it doesn't exist.

Example configuration:
```json
{
    "userAgent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36...",
    "acceptLanguage": "en-US",
    "upgradeInsecureRequests": "1",
    "acceptEncoding": "gzip, deflate, br",
    "connection": "keep-alive",
    "additionalHeaders": [
        {
            "key": "Content-Type",
            "value": "text/plain;charset=UTF-8"
        }
    ]
}
```

## Output

The tool generates two output files:

1. **certInformation.json**: Structured JSON data with all certificate details
2. **certInformation.csv**: CSV format for easy analysis in spreadsheet applications

### Output Fields

- URL: The scanned website URL
- Notes: Any errors or notes from the scan
- Subject: Certificate subject
- Issuer: Certificate issuer
- CommonName: Common Name from issuer
- Country: Country from issuer
- IssuerSerialNumber: Issuer's serial number
- NotBefore/NotAfter: Certificate validity period
- DNSNames: Subject Alternative Names
- SerialNumber: Certificate serial number
- SignatureAlgo/PublicKeyAlgo: Cryptographic algorithms
- Version: Certificate version
- Emails: Email addresses from certificate
- IPAddresses: IP addresses from certificate

## Input File Format

For batch processing, provide a text file with one URL per line:
```
https://example.com
https://google.com
https://github.com
```

## Requirements

- Go 1.16 or higher
- Internet connection for scanning external websites

## Notes

- The tool skips certificate verification (`InsecureSkipVerify: true`) to handle self-signed certificates
- No redirects are followed during requests
- Default timeout is 30 seconds per request

## License

[License Documentation](LICENSE.md)