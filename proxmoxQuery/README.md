# Proxmox API Query

This document describes the setup of a simple client using the proxmox API to query my proxmox for information.  This project was created to build a foundation for creating a VM from a template.


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

I love using proxmox as a hypervisor.  The GUI works to gather this information, however to consolidate in a quick report I built this.  While building I had to solve the following challenges.

Challenges:
1. Configure the permissions in proxmox to Allow the use of the API.  I saw a comment on one post as I was researching that said, once you know how you know how...
2. Build the script to automate the pull of the proxmox node name, then pull the VM information of CPUs, Memory, and where the storage is provisioned.  Show if the CPUs or memory are over-provisions.
3. Show the storage and how much was used.

Working through these challenges allowed the discovery of how to build from a template a VM.


![Scorpion Soldiers](/picts/scorpionDroidSoldiers.png)


## Solution

The golang program allowed for an efficient view of the VM's where their storage was located across multiple drives and shows if over provisioning is occurring.  The script also displays the storage that is available for usage.  With this foundation it will assist in automating the creation of a VM in a later program.

---

## Compiling and Configuration

A `prep.sh` script is provided to streamline the compilation process. Ensure the `GOPATH` environment variable is correctly set before execution:

```bash
go env -w GOPATH="/home/thepcn3rd/go/workspaces/2025/proxmoxQuery"
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
	"tokenID": "tokenIDcreatedinProxmox",
	"tokenSecret": "ffffffff-ffff-ffff-ffff-ffffffffffff",
	"APIURL": "https://10.10.10.10:8006"
}
```

#### Key Configuration Details
* Token ID: Provided when an API User is created in the Proxmox Interface
* Token Secret: A GUID provided upon the creation of an account
* API URL: Proxmox URL which is used for authentication on port 8006


---

## Execution Instructions

To execute the binary and specify a custom configuration file:

```bash
./proxmoxQuery -config config.json
```

### Command-Line Usage

```txt
Usage of ./proxmoxQuery.bin:
  -config string
    	Configuration file to load for the proxy (default "config.json")
```

---

## Troubleshooting

1. **Binary Not Found:** Ensure `prep.sh` has executed successfully.
2. Verify in Proxmox that the permissions are setup for read-only
3. Verify no typos exist in the config.json file

---

## License

This project is licensed under the [MIT License](/LICENSE.md).