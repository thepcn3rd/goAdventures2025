# Gmail Search API

This document describes how to use the Gmail API provided in the Google Cloud Console to search your email.  I was using the search to find keywords among thousands of emails.  The primary purpose is to identify passwords or tax information and remove it from my email.  

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

I have used gmail for over 10 years and sometimes will go back through my emails and find items that I meant to delete after I sent or received them.  This allows a search of the contents of my inbox to find items I should delete or items that I should archive.

Challenges:
1. Configure in Google Cloud Console the Gmail API and understand how to do it with least privileges.
2. Create an OAuth 2.0 Client ID, build with least privileges and then not allow complete automation so that I can control the access to my email with the testing conducted.
3. Learn how to capture the OAuth 2.0 authorization code to then automate the usage of it with a web server that is temporarily running with the golang program.


![Scorpion Soldier](/picts/scorpionDroidSoldier.png)

## Solution

In the Google Cloud Console I learned about creating OAuth 2.0 Client ID's and providing them the gmail.readonly if I am searching and archiving; or gmail.modify permissions if I am moving messages to the trash.  Then with the web server that is spun up, creating an ID with type "web application" with the URI of the listening port of the web server.  A note here on security is the parameters are sent by way of a GET request so the values can be seen in the browsers history.  It works well! This could be used to tag messages also, this could allow a tag to be placed so you are not reviewing the same 10 messages in the stack each time you go through.  This tool was also created to be a red team tool, in the event of access to an email account it could be searched quickly for any passwords to lead to additional access.

## Compiling and Configuration

A `prep.sh` script is provided to streamline the compilation process. Ensure the `GOPATH` environment variable is correctly set before execution:

```bash
go env -w GOPATH="/home/thepcn3rd/go/workspaces/2025/gmailSearch"
```

Run the script with the following command:

```bash
. prep.sh
```

You can modify the script to change binary names or enable the creation of Windows binaries by uncommenting relevant lines.

---

## Configuration File

The `config.json` file contains parameters for network-specific settings. Below is a sample configuration:

```json
{
        "listeningPort": ":8080",
        "tlsEnabled": true,
        "sslConfig": "keys/certConfig.json",
        "sslCert": "keys/server.crt",
        "sslKey": "keys/server.key",
        "credentialFile": "credentials.json",
        "tokenFile": "token.json",
        "archiveDirectory": "archive",
        "query": "(in:inbox OR in:sent) (password OR taxes)"
}
```

#### Key Configuration Details
* Listening Port: The temporary web server that is setup to capture the authorization code in OAuth2 to create the token
* TLS Enabled: Toggle if you want Encrypted Communication
* SSL Config: Allows you to configure the Self-Signed Certificate that is automatically created for the Listening Port, after initial execution if this file is missing it will be created
* SSL Cert: The location of the certificate file for the web server, this allows customization and flexibility
* SSL Key: The location of the key file for the web server
* Credential File: After creating the OAuth2 ID you will need to download the file associated with it and save it to the directory with this name. 
* Token File: After the initial load using the credential file a token file is created that can be reused instead of the negotiating that occurs to authorize the token
* Archive Directory: The location of where messages are saved if they are archived
* Query: The query that is executed to find the messages


## Execution Instructions

To execute the binary and specify a custom configuration file:

```bash
./gmailSearch -config config.json
```

### Command-Line Usage

```txt
Usage of ./gmailSearch.bin:
  -config string
    	Configuration file to load for the proxy (default "config.json")
```

---

## Troubleshooting

1. **Binary Not Found:** Ensure `prep.sh` has executed successfully.
2. Webserver does not start, unless running as a privileged user or setup to run privileged use a port greater than 1024 in the configuration
3. Verify the URI is specified correctly in the OAuth2 ID configuration and that the consent is completed
4. If the TLS certificate is not setup it will end the program and prompt you to complete the configuration for the self-signed certificate
5. Verify that your test account includes the email address you are using
6. Verify no typos exist in the config.json file
7. Expect the browser to prompt you about allowing this application.  The authorization stage requires about 3-4 different pages to click through prior to creating the token, however secure the token afterwards.

---

## License

This project is licensed under the [MIT License](/LICENSE.md).
