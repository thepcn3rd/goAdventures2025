# Create OTP

This document describes the setup of a simple program to generate an one-time passcode that will use CyberChef to generate a QR Code to scan using Google Authenticator or similar application.

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

I started a project to build an application that required the creation of a one-time passcode for multi-factor authentication.  I found a few projects that could be used as dependencies that were less than 1000 lines of code.  I used a couple of them and then created only the necessary elements that I needed from the code.

Challenges:
1. Reading Code: Understand and recreate in a simple way the code of others on the web
2. Building a QR Code to be Scanned: This was an unusual obstacle but decided to use CyberChef because it can be executed in a static web server that I built in

---

## Solution

The golang program is built to generate the Secret Code that can be entered manually into Google Authenticator or a URL is used to generate a QR Code.  The program uses CyberChef either online or offline to generate the QR Code.  If you are using it offline you need to unzip CyberChef in the static directory and rename the HTML file to index.html.

![Scorpion Droid in the Jungle](/picts/scorpionDroidGold.png)

## Compiling and Configuration

A `prep.sh` script is provided to streamline the compilation process. Ensure the `GOPATH` environment variable is correctly set before execution:

```bash
go env -w GOPATH="/home/thepcn3rd/go/workspaces/2025/createOTP"
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
        "appName": "MyApp",
        "accountName": "user@example.com",
        "webServerPort": 8088,
        "useCyberChefLocal": true
}
```

#### Configuration Details

* App Name - This is the name of the app where the OTP is being used
* Account Name - This is the account that you typically use the OTP with
* Web Server Port - This is the port the local web server will use if you run Cyber Chef locally
* Use CyberChef Local - If this is set to true it runs a local web server to create the QR Code, if this is set to false it uses the online version of CyberChef 

---

## Execution Instructions

To execute the binary and specify a custom configuration file:

```bash
./otp -config config.json
```

### Command-Line Usage

```txt
Usage of ./otp.bin:
  -config string
    	Configuration file to load for the proxy (default "config.json")
```

---

## Troubleshooting

1. **Binary Not Found:** Ensure `prep.sh` has executed successfully.
2. Make sure you unzip CyberChef into the static directory, sometimes it will create a directory inside static
3. In the static directory after you unzip CyberChef verify you rename the main HTML file to index.html
4. Make sure you copy the full URL that is provided

---

## License

This project is licensed under the [MIT License](/LICENSE.md).