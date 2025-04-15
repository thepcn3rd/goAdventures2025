# goPwnCheck

This Go program checks if a password or its hash (SHA-1 or NTLM) has been compromised using the k-Anonymity model with the Have I Been Pwned (HIBP) API. It supports interactive input, file input, and can be configured via a JSON configuration file.  The program also checks and stores in offline files the hash information gathered from HIBP.

## Features

- **Interactive Mode**: Allows users to input a plain-text password or a hash directly from the command line.
- **File Input**: Reads a list of hashes or plain-text passwords from a file and checks each one against HIBP.
- **SHA-1 and NTLM Hash Support**: Supports both SHA-1 and NTLM hash formats.
- **k-Anonymity Model**: Uses the k-Anonymity model to securely check passwords against the HIBP API without exposing the full hash.
- **Configuration File**: Allows customization of the API URL, user agent, request delay, skip the load or saving of offline files.

## Installation

Compile, install dependencies and create the binary
```bash
./prep.sh   
```


![Scorpion Soldier Entering a Password](/picts/scorpionSoldierPasswords.png)


## Usage

```bash
Usage of ./pwnCheck.bin:
  -config string
        Configuration file to load (default "config.json")
  -f string
        File to load and read line-by-line that contains SHA1 or NTLM hashes
  -i    Use Interactive Mode
  -ntlm string
        File to load and read plain-text passwords and convert into NTLM hashes
  -sha1 string
        File to load and read plain-text passwords and convert into SHA1 hashes
```


### Interactive Mode Example

```bash
$ ./pwnCheck.bin -i
Interactive Mode - Select Option
1. Input plain text password
2. Input SHA1 or NTLM hash
> 1

Select Hash to use for the Password
1. SHA1
2. NTLM
> 1

Input Plain-text Password
> mypassword

[+] Password Hash Exists: 5BAA61E4C9B93F3F0682250B6CF8331B7EE68FD8
```

### File Input Mode of Hashes Example

```bash
$ ./pwnCheck.bin -f hashes.txt
[*] Processing File: hashes.txt

[+] Password Hash Exists: 5BAA61E4C9B93F3F0682250B6CF8331B7EE68FD8
[-] Password Hash Not Available: 098F6BCD4621D373CADE4E832627B4F6
```

### File Input Mode of Plain-text Passwords to SHA1 Example

```bash
$ ./pwnCheck.bin -sha1 plaintextPasswords.txt
[*] Processing File: plaintextPasswords.txt

[+] Password Hash Exists: 5BAA61E4C9B93F3F0682250B6CF8331B7EE68FD8
```


### File Input Mode of Plain-text Passwords to NTLM Example

```bash
$ ./pwnCheck.bin -ntlm plaintextPasswords.txt
[*] Processing File: plaintextPasswords.txt

[-] Password Hash Not Available: 098F6BCD4621D373CADE4E832627B4F6
```



## Offline Files

The information gathered from the HIBP API by default is saved offline to the Offline Files directory stored in config.json in respective directories for SHA1 and NTLM.  The offline files have no delay for lookup but to communicate to the HIBP it has a default delay of 3 seconds configured in the config.json file.

The current examples of the passwords; "Password123" and "Welcome123" exist in the offline files.  They both exist in the SHA1 location of the offline files, and "Welcome123" exists as an NTLM hash.