# HTTP Method Tester

A simple Go tool to test HTTP methods on web servers and endpoints.

## Features

- **HTTP Method Testing**: Test various HTTP methods (GET, POST, PUT, DELETE, etc.)
- **Response Analysis**: Check status codes and response details
- **TLS Support**: Handles HTTPS connections with insecure skip verify
- **Colored Output**: Clear color-coded results for easy reading

## Installation

```bash
./prep.sh
```

## Usage

### Basic Usage

```bash
# Test GET method on a URL (default)
./methodTest -u https://example.com

# Test specific HTTP method
./methodTest -u https://api.example.com/users/1 -m DELETE
```

## Output

The tool provides color-coded output showing:

- Method allowance status
- Response headers
- Response body (if any)
- HTTP status codes

### Sample Output
```
DELETE method is Allowed and Succeeded

Response Headers
Content-Type: application/json
Server: nginx

Response Body
{"status": "deleted"}
```


## License

[License](LICENSE.md)