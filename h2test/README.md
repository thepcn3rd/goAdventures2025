# HTTP/2 Scanner

A Go-based tool to check if websites support HTTP/2 protocol.

## Features

- **HTTP/2 Detection**: Verify if a target website supports HTTP/2 protocol
- **Single URL Scanning**: Test individual URLs for HTTP/2 support
- **Batch Processing**: Scan multiple URLs from a file
- **Simple Output**: Clear results indicating HTTP/2 support status

## Installation

```bash
./prep.sh
```

## Usage

### Basic Usage

```bash
# Scan a single URL
./h2test -url https://example.com

# Scan multiple URLs from a file
./h2test -list urls.txt
```

## Input File Format

For batch processing, provide a text file with one URL per line:
```
https://example.com
https://google.com
https://github.com
```

## Output

The tool outputs results in the format:
```
https://example.com, HTTP/2 is supported
https://old-website.com, HTTP/2 is not supported
```

## Dependencies

- Go 1.16 or higher
- `golang.org/x/net/http2` for HTTP/2 protocol support
- `github.com/thepcn3rd/goAdvsCommonFunctions` for common utility functions

## License

[License](LICENSE.md)
