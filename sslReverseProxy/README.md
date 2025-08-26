# TLS Reverse Proxy

This document describes the setup and functionality of a custom TLS reverse proxy built in Go, designed to enhance the security of a Guacamole server running in a Docker environment. The proxy enables TLS encryption, logging, and flexible configuration options.

---

## Problem Statement

Guacamole, an open-source remote desktop gateway, does not support TLS by default when deployed via Docker. By default, Guacamole listens on HTTP, which poses a security risk, especially when authentication occurs over plain text. Apache recommends placing a reverse proxy (e.g., nginx) in front of Guacamole to secure connections.

This project addresses two primary use cases:

1. **Home Lab Management**: Managing multiple VMs on Proxmox requires a secure and seamless remote desktop experience. Tools like noVNC or Remmina have limitations, particularly with copy-paste functionality between hosts and VMs.  Use this reverse proxy to secure the authentication to guacamole
    
2. **Teaching Environment**: As a professor at Ensign College, I needed a secure and efficient way to allocate VMs to students, ensuring security and ease of access.
    

The custom Go-based TLS reverse proxy provides a lightweight, solution to secure Guacamole connections without modifying its Docker image or relying on nginx.

---
## Solution

### Design and Implementation

The solution builds on Docker-based Guacamole setup guides, such as the [Linode guide](https://www.linode.com/docs/guides/installing-apache-guacamole-through-docker/). After deploying Guacamole:

1. **Initial Setup**: The proxy was containerized, but logging features didn't function as I liked to get the logs to an Elastic SIEM.
    
2. **Refinement**: Hosting the proxy directly on the VM running Guacamole allowed direct communication with Guacamole's private Docker network (e.g., `172.17.0.4`), eliminating the need to expose port `8080` on the VMs interface. (View the network address by running 'docker inspect --random name of docker--)
    

This approach secures Guacamole's connection, simplifies configuration, allows copy/paste between VMs, and enables detailed logging via syslog or local files.

![Scorpion Fight](/picts/scorpionFight.png)

## Compile and Configure the SSL Reverse Proxy

### Compiling the SSL Reverse Proxy

A `prep.sh` script is provided to streamline the compilation process. Ensure the `GOPATH` environment variable is correctly set before execution:

```bash
go env -w GOPATH="/home/thepcn3rd/go/workspaces/2025/dnsServer"
```

Run the script with the following command:

```bash
. prep.sh
```

You can modify the script to change binary names or enable the creation of Windows binaries by uncommenting relevant lines.

### Configuration File

The `config.json` file contains parameters for network-specific settings. Below is a sample configuration:

```json
{
	"listeningPort": "0.0.0.0:8443",
	"proxiedHTTPEndpoint": "http://172.17.0.4:8080",
	"sslConfig": "keys/certConfig.json",
	"sslCert": "keys/server.crt",
	"sslKey": "keys/server.key",
	"syslogOptions": {
		"syslogEnabled": "True",
		"syslogServer": "10.27.20.210:514",
		"syslogOriginName": "proxy-server"
	},
	"saveFileOptions": {
		"saveFileEnabled": "True",
		"saveFileBaseName": "proxy",
		"saveFileExtension": ".log"
	},
	"basicAuthOptions": {
        "enabled": true,
        "username": "thepcn3rd",
        "password": "T3sting"
    }
}
```

#### Key Configuration Details

- **Listening Port**: Define the IP and port to bind (e.g., `0.0.0.0:8443`). For ports <1024, configure system permissions.
- **Proxied HTTP Endpoint**: Target endpoint for proxying (default: Guacamole's `http://172.17.0.4:8080`).
- **SSL Certificates**:
    - `sslConfig`: Path to a JSON file defining certificate creation parameters.
    - `sslCert` and `sslKey`: Paths to the certificate and key files.
- **Logging Options**:
    - `syslogOptions`: Enable syslog with server and origin name configuration.
    - `saveFileOptions`: Enable/disable local file logging and customize file naming.
- Basic Authentication Options:
	- enabled: Allows you to provide a username and password entry to what the SSL Proxy is in-front of in-the-event of not having authentication.  I am hosting a Revshells instance that does not have authentication, this works well to add a simple layer of protection
	- username: The username that is verified
	- password: Password that is verified after it creates a SHA256 checksum and compares it to what is input by the user

#### Certificate Configuration (`certConfig.json`)
This file specifies parameters for generating `server.crt` and `server.key`. If these files are missing, the Go program will recreate them using `certConfig.json`.

## Execution Instructions

To execute the binary and specify a custom configuration file:

```bash
./revProxy -config home.config.json
```

### Command-Line Usage

```txt
Usage of ./revProxy.bin:
  -config string
    	Configuration file to load for the proxy (default "config.json")
```

---

## Troubleshooting

1. **Binary Not Found:** Ensure `prep.sh` has executed successfully.
2. **Permission Issues:** Check file permissions for binaries and configuration files.
3. **Syslog Not Working:** Verify syslog server address and connectivity.
4. **Name Resolution Fails:** Verify you are connecting with the URL in the certConfig.json file that you configured.

---

## Optional Configuration with Docker
You can utilize the runDocker.sh script and modify the Dockerfile to run the reverse proxy in a docker.  However, logging does not function correctly.

---

## License

This project is licensed under the [MIT License](/LICENSE.md).

