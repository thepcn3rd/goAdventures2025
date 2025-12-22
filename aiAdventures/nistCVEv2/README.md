# CVE Monitoring and Analysis Tool

A Go-based tool that monitors, analyzes, and generates reports for Common Vulnerabilities and Exposures (CVEs) using NIST NVD data and Brave Search API with AI-powered analysis.  Built as an experiment to use the go ADK and an ollama connector that I created for a home lab.

## Overview

This tool performs CVE monitoring by:

1. **Pulling recent CVEs** from the NIST National Vulnerability Database (NVD)
2. **Searching for related information** using the Brave Search API
3. **AI-powered analysis** using Ollama to verify relevance of the search results
4. **AI-powered analysis** using Ollama to create and select a summary of the search results
5. **Markdown report generation** for security teams

## Prerequisites

- Go 1.19 or higher
- Brave Search API key
- Ollama instance running locally or remotely

## Installation

1. Compile the Binary:
```bash
./prep.sh
```

2. Set up configuration:
```bash
go run main.go -config config.json
```
This will generate a default `config.json` file if it doesn't exist.

## Configuration

Edit `config.json` to customize the tool:

```json
{
    "braveAPIKey": "your-brave-api-key",
    "ollamaURL": "http://localhost:11434/api/chat",
    "ollamaWaitTime": 10
}
```

### Configuration Fields

- `braveAPIKey`: API key for Brave Search service
- `ollamaURL`: URL for Ollama API endpoint
- `ollamaWaitTime`: Timeout in minutes for Ollama responses

## Usage

```bash
./nistCVEv2.bin
```

## How It Works

### 1. CVE Data Collection
- Pulls CVE data from NIST NVD API for the specified timeframe
- Filters results based on configured keywords
- Processes vulnerability metadata and descriptions

### 2. Web Intelligence Gathering
- Searches Brave Search for each CVE ID
- Retrieves relevant news articles and technical content
- Collects additional references and context

### 3. AI-Powered Analysis
- **Verification**: Multiple AI agents verify if search results match CVE topics
- **Summarization**: AI generates concise summaries of CVE impact and context
- **Quality Control**: Requires multiple matching sources for report generation

### 4. Report Generation
- Creates markdown files for each significant CVE
- Includes severity ratings, descriptions, AI summaries, and references
- Outputs to `output/` directory with CVE ID as filename
## Output

The tool generates markdown reports in the `output/` directory with the following structure:

```markdown
## CVE-YYYY-XXXXX
- **URL:** [NVD Link]
- **Severity:** CRITICAL/HIGH/MEDIUM/LOW
- **Vulnerability Status:** Analyzed/Modified/Rejected

### CVE Description
[Official CVE description from NVD]

### AI Summary from Search Results
[AI-generated summary of related information]

### Additional URL References
[List of relevant articles and references]
```

## Project Structure

```
nistCVEv2/
├── main.go                # Main application entry point
├── agentsVerifySearch.go  # Custom code for running Agentic AI
├── braveCustom.go         # Brave custom code for searching
├── nvdCustom.go           # Custom code for filtering NVD Information
├── config.json            # Configuration file (auto-generated)
├── output/                # Generated markdown reports
├── ollama/                # Ollama Connector struct
├── nvd/                   # NIST NVD API integration struct
├── brave/                 # Brave Search API integration struct
└── README.md              # README
```

![Process Diagram](/aiAdventures/nistCVEv2/processDiagram.png)

## License

[License Information](/LICENSE.md)
