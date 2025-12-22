# Cryptographic Key & Hash Generator

A Go-based utility for generating various cryptographic keys and hashes, including SSH public keys and password hashes.  This was used to create random keys and hashes for the "hashScanner" project.

## Features

- **SSH Key Generation**: Generate SSH public keys in multiple formats:
  - RSA (2048-bit)
  - ECDSA P-256
  - ECDSA P-384

- **Password Hashing**: 
  - SHA-512 hashing
  - SHA-512 hash crypt

## Configuration

Modify the main.go file to the type of hashed password you want generated.

## Installation

```bash
./prep.sh
```

## Usage

### Basic Usage

```bash
# Generate ECDSA P-384 SSH public key (default)
./genHash

# Generate with a specific password
./genHash -p "your-password-here"
```

### Command Line Options

- `-p`: Password to hash (if empty, generates random passwords)

## Dependencies

- Go 1.16 or higher
- `golang.org/x/crypto/ssh` for SSH key handling
- Standard crypto libraries for cryptographic operations

## License

[License](LICENSE.md)

