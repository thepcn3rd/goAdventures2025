# Validate OTP

This document describes the setup of a simple program to validate an one-time passcode that is input by a user.

---

## Table of Contents

- [Problem Statement](#problem-statement)
- [Solution](#solution)
- [Compiling and Configuration](#compiling-and-configuration)
- [Configuration File](#configuration-file)
- [Execution Instructions](#execution-instructions)
- [Troubleshooting](#troubleshooting)

---

## Problem Statement

I started a project to build an application that required the creation and validation of a one-time passcode for multi-factor authentication.  I found a few projects that could be used as dependencies that were less than 1000 lines of code.  I used a couple of them and then created only the necessary elements that I needed from the code.

Challenges:
1. Reading Code: Understand and recreate in a simple way the code of others on the web
2. Validate Input Code: I observed a time skew of about 5 minutes and 30 seconds on the codes that I could not figure out mathematically

---

## Solution

The golang program is built to verify a code input by the user from a google authenticator compared to a stored Secret Key that is contained in the config.json.

![Scorpion Droid Jumping a Crevasse](/picts/scorpionDroidJumping.png)

## Compiling and Configuration

A `prep.sh` script is provided to streamline the compilation process. Ensure the `GOPATH` environment variable is correctly set before execution:

```bash
go env -w GOPATH="/home/thepcn3rd/go/workspaces/2025/validateOTP"
```

Run the script with the following command:

```bash
. prep.sh
```

You can modify the script to change binary names or enable the creation of Windows binaries by uncommenting relevant lines.

## Configuration File

The `config.json` file contains parameters for network-specific settings. Below is a sample configuration:

```json
{
	"secretKey": "AAABBBCCCDDDEEEFFF11122233344455"
}

```

#### Configuration Details

* Secret Key - This was generated initially when the Secret Code and/or the QR Code was generated to be stored in the google authenticator

---

## Execution Instructions

To execute the binary and specify a custom configuration file:

```bash
./validateOTP -config config.json
```

### Command-Line Usage

```txt
Usage of ./validateOTP.bin:
  -config string
    	Configuration file to load for the proxy (default "config.json")
```

---

## Troubleshooting

1. **Binary Not Found:** Ensure `prep.sh` has executed successfully.
2. Verify the Security Key is correct and is upper-case letter

---

## License

This project is licensed under the [MIT License](/LICENSE.md).