# goRunAs - Run as Different User in Windows

A Go implementation of Windows' `runas` command that allows executing programs with different user credentials, similar to the built-in `runas` command but with more flexibility.

## Features

- Execute programs with a different users credentials
- Supports both command-line arguments and interactive mode
- Works with domain and local accounts
- Handles program arguments

## Installation

Compile, install dependencies and create the binary
```bash
./prep.sh   
```


## Usage

```text
Usage: goRunAs.exe [options]

Options:
-u string OR -user string
        Username
-p string OR -password string
        Password
-d string OR -domain string
        Domain or use a "." for local computer
-e string OR -exec string
        Program to execute
-a string OR -args string
        Arguments for the program enclosed by quotes
-i OR -interactive
        Interactive Mode
```

## Example
```bash
goRunAs.exe -u bradley.marks -p <password> -d 4gr8.local -e notepad.exe
```

## Interactive Mode Example

The below executes notepad.exe with the specified username, password and domain
```bash
C:\Users\Administrator>goRunAs.exe -i
Enter username: bradley.marks
Enter password: <password>
Enter domain: 4gr8.local
Enter program to execute: notepad.exe
```

