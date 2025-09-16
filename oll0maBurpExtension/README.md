# Oll0ma Queue Server: Asynchronous LLM Analysis for Burp Suite

## Overview
The oll0ma queue server is built to work with "oll0ma Submit to API" burp extension (oll0maExtension.py).  The extension will prepare information provided to it of an HTTP Request and HTTP Response from Burp to the Queue Server.  The Queue Server will then manage the queue and processing of the information going to a LLM.  The LLM based on the system prompt is supposed to identify weaknesses and vulnerabilities. After the information is processed it is saved on the queue server and then can be viewed in Burp with the "oll0ma View" burp extension (oll0maView.py)

The oll0ma queue server has an index.html page where you can see the API Endpoints and conduct a direct submission to the LLM.  I also created a simple 1 page app using MIT App Inventor to interact with /api/submit, however the TLS certificate if it is not valid on mobile will reject the connection. (androidPOCApp.apk)

After working with the openwebui, Verizon Red Team AI and the burpference project I learned from them and created oll0ma.  The core reason was I have a local LLM that would take upwards of 3 minutes to process a request.  I needed a server that would queue the request and then handle the processing with the LLM.  

oll0ma contains the following components:

1. queueServer - This receives an HTTPS request containing the system prompt, request and response received from Burp
2. oll0ma Extension - This is a UI similar to burpference where the queue server IP address is configured and visually you can see what is sent to the queue server
3. oll0ma File Viewer - Through the queue server the files created with the responses from the LLM can be viewed

## Queue Server (`main.go`)

The queue server is the core backend service written in Go. It performs the following functions:
*   **Manages a Job Queue:** Accepts analysis jobs from the Burp extension and queues them for processing to prevent overwhelming the LLM.
*   **Processes Jobs Asynchronously:** Handles communication with the Ollama API, sending the HTTP request/response and a system prompt for analysis.
*   **Stores Results:** Saves the LLM's analysis output to timestamped files in an `output/` directory.
*   **Provides a Management API:** Offers endpoints for health checks, statistics, and file retrieval for the Burp View extension.
*   **Secures Communications:** Uses HTTPS with TLS and API key authentication for all client-server interactions.

### API Endpoints

All endpoints require a valid `APIKey` to be provided in the request body (for POST) and expect/provide JSON.

* **`GET /index.html`**
    *   **Purpose:** Home page that references the API endpoints and a direct submission.    
     
<center><img src="/oll0maBurpExtension/picts/indexPage.png" width="70%" style="border:2px solid black;" /></center>

* **`GET /api/direct.html`**
    *   **Purpose:** Allows direct submission of prompts to the LLM.  The results can be found with oll0ma View the burp suite extension.
    *   **Request Body:** Complete the form with the following `Model`, `API Key`, `Request Number`, `System Prompt`, and the `Prompt`. Then click submit.
    *   **Response:** Returns that the request is queued and can be viewed through oll0maView
    
<center><img src="/oll0maBurpExtension/picts/directSubmission.png" width="70%" /></center>

*   **`POST /api/submit`**
    *   **Purpose:** Submits a new job for LLM analysis.
    *   **Request Body:** A `Request` object containing base64-encoded `model`, `request`, `response`, `systemPrompt`, and the `apiKey`.
    *   **Response:** Returns a job ID with a status of `queued`.

*   **`POST /api/files`**
    *   **Purpose:** Retrieves a list of all available result files stored in the `output/` directory.
    *   **Request Body:** A `FileRequest` object containing only the `apiKey`.
    *   **Response:** A JSON object with a `files` array listing all result files.

*   **`POST /api/file`**
    *   **Purpose:** Downloads the content of a specific result file.
    *   **Request Body:** A `FileDownloadRequest` object containing the `apiKey` and `fileName`.
    *   **Response:** A JSON object containing the file content as a base64-encoded string (`encodedFile`) and its SHA256 hash for verification.

*   **`GET /api/health`**
    *   **Purpose:** A health check endpoint to monitor server status.
    *   **Response:** Returns server status, timestamp, and current queue metrics.

*   **`GET /api/stats`**
    *   **Purpose:** Provides statistics on total jobs processed and current queue state.
    *   **Response:** Returns total jobs processed, current queue size, and capacity.
<center><img src="/oll0maBurpExtension/picts/apiStats.png" width="70%" /></center>


## Installation & Configuration

### 1. Queue Server Setup

1.  **Build the Queue Server:**
    ```bash
    # Modify the prep.sh file with appropriate directories
    ./prep.sh
    ```

2.  **Configure the Server:**
    A `config.json` file will be automatically generated on first run. Modify it to set your preferences:
    ```json
    {
        "queueServerURL": "https://localhost:8443",
        "ollamaURL": "http://localhost:11434/api/chat",
        "apiKey": "your_secure_api_key_here",
        "tlsConfig": "keys/tlsConfig.json",
        "tlsCert": "keys/server.crt",
        "tlsKey": "keys/server.key"
    }
    ```
    *   `queueServerURL`: The address where this server will run (must be HTTPS).
    *   `ollamaURL`: The URL of your Ollama API `chat` endpoint.
    *   `apiKey`: A secret key for authenticating requests from Burp. If empty, one is generated on startup.
    *   The TLS files will be auto-generated if they don't exist.

3.  **Run the Server:**
    ```bash
    ./oll0maQueue.bin
    ```
    Note the generated API key if this is the first run.

### 2. Burp Extension Installation

The Burp extensions (`oll0maExtension.py` and `oll0maView.py`) are written in Python and must be installed within Burp Suite's **Extender** tool.

1.  Open Burp Suite and navigate to the **Extender** tab.
2.  Click on the **Extensions** sub-tab.
3.  Click **Add**.
4.  In the "Add extension" dialog:
    *   Set **Extension Type** to `Python`.
    *   Click **Select file...** and browse to the location of `oll0maExtension.py`.
    *   Click **Next**. The extension should compile and load without errors. An "oll0ma" tab will appear in the UI.
5.  Repeat steps 3-4 for the `oll0maView.py` file. A second "oll0ma View" tab will appear.

### 3. Configuring the Burp Extensions

1.  **Configure the Submit Extension:**
    *   Click on the new **oll0ma** tab in Burp.
    *   In the settings area, enter the **Queue Server URL** (e.g., `https://localhost:8443`).
    *   Enter the **API Key** that was configured in or generated by the queue server.
    *   The extension will use these settings to submit HTTP requests/responses for analysis.

2.  **Using the View Extension:**
    *   Click on the **oll0ma View** tab.
    *   Enter the same **Queue Server URL** and **API Key**.
    *   Click **Load Files** to fetch the list of available analysis results from the server.
    *   Select a file from the list and click **Download File** to fetch and display the LLM's analysis directly in Burp.

## Usage Workflow

1.  **Proxy:** Use Burp Proxy to click on an HTTP request with related response.
2.  **Submit:** Right-click the request in any Burp tool (Proxy, Repeater, etc.) and send it to the oll0ma extension, or use the UI in the **oll0ma** tab.
3.  **Queue:** The extension sends the data to the queue server, which places it in a processing queue.
4.  **Process:** The server sends the request to the configured Ollama instance for analysis based on the system prompt and saves the result.
5.  **Review:** Open the **oll0ma View** tab in Burp, load the list of results, and download the analysis for your request.

## Set Options for more Conservative Answers and Consistency 

Currently the options are hard coded in the main.go file, plan is to move them to the config.json file.

```
modelOptions.NumCTX = 4096 // Context window size - Setting to 4096 for more tokens to be used, Default: 2048

modelOptions.Temperature = 0.1 // I am not looking for creativity - 0.1 is conservative, Default: 0.8

modelOptions.RepeatLastN = 0 // How far back to look - Default: 64, Setting to 0 disabled to disable the look back

modelOptions.RepeatPenalty = 1.1 // At this time leaving this alone

modelOptions.TopK = 10 // Probability of generating non-sense - Lowering this to 10, Default: 40

modelOptions.TopP = 0.5 // Focused and conservative is 0.5, Default: 0.9
```

## Measuring Different Models (Test 1)

Output files for the below are under the test1 folder...

Conducted a measured test using different models and if the information they returned for an HTTP Request Response and then 2 direct submissions for the same pages using HTB: Poison.  Also, testing the model options below.

Model Options: 
NumCTX=4096, Temperature=0.1, RepeatLastN=0, RepeatPenalty=1.1, TopK=10, TopP=0.5
WaitTime = 15 minutes for a Job

Measuring criteria:
- 1 point is awarded for a helpful suggestion provided in the analysis (per suggestion)
- 1 point is awarded for a correct analysis of a security vulnerability
- 1 point if the model did not timeout within 15 minutes

| Model            | r24 | r26 | r27 | r28 | r29 | r37 | d1  | d2  | Total   |
| ---------------- | --- | --- | --- | --- | --- | --- | --- | --- | ------- |
| qwen3:0.6b       | 1   | 0   | 0   | 0   | 0   | 0   | 1   | 1   | 3       |
| llama3.2:3b      | 2   | 1   | 1   | 0   | 3   | 2   | 3   | 2   | 14      |
| deepseek-r1:1.5b | 1   | 0   | 0   | 0   | 1   | 0   | 2   | 0   | 4       |
| gemma3:1b (1st)  | 5   | 0   | 0   | 0   | 0   | 0   | 0   | 0   | 5 (DNF) |
| gemma3:1b (2nd)  | 3   | 3   | 0   | 0   | 0   | 0   | 0   | 0   | 6 (DNF) |
| mistral:7b       | 3   | 0   | 1   | 1   | 2   | 1   | 3   | 1   | 12      |
| deepseek-r1:8b   |     |     |     |     |     |     |     |     | N/A     |

Expectations:
1. (r26-r37) Looking for a suggestion for a local file inclusion file vulnerability
2. (r26-r37) Looking for a suggestion for a remote file inclusion file vulnerability
3. (d1) Looking for a suggestion of how to read the output of nmap
4. (d2) Looking for a suggestion of how to decode a base64 block of text
5. (r37) Analysis of a backup file being stored with the web server files
6. Points for correct analysis of LFI, RFI, output of nmap and the base64 block of text

Model deepseek-r1:8b
Exceeds the waittime due to the hardware that I have for testing

Model mistral:7b
1. (r24) Identified old php version needs to be updated
2. (r24) Evaluated the response to find the file parameter
3. (r24) Suggested Input Validation
4. (r26) Identified that it is evaluating the php.ini file and the various directives
5. (r27) Identified the old php version however point is awarded above with r24
6. (r29) Identified a path traversal vulnerability
7. (r29) Suggests hiding server header information
8. (r37) Identifies that a password is revealed, a little sarcasm with it...
9. (d1) Identified the 2 open ports
10. (d1) Identified that the SSH Version is old
11. (d1) Identified the apache version is older
12. (d2) Identified the string as base64 (Did not decode)

Model qwen3:0.6b
1. (r24) Suggested CSRF Tokens
2. (d1) Identified the 2 open ports
3. (d2) Suggested base64 

Model gemma3:1b (2nd - Did not finish - Not sure why...)
1. (r24) Suggested testing for XSS in Headers
2. (r24) Suggested input validation
3. (r24) Suggested CSRF Protection
4. (r27) Recommended the use of Content Security Policy
5. (r27) Suggested to Update PHP
6. (r27) Suggested to Review Code
7. 

Model gemma3:1b (1st - Did not finish)
1. (r24) Suggested testing the Accept header with XSS
2. (r24) Evaluated the response using the file parameter with input validation
3. (r24) Provided Cookie Management suggestions for CSRF
4. (r24) Recommended the use of Content Security Policy
5. (r24) Recommended a php code review due to the php version
6. Notes: Did not complete processing r26 hung ollama restarting the analysis with this model


Model deepseek-r1:1.5b
1. (r24) Analyzed the HTTP response having browse.php
2. (r29) Suggests to remove the file parameter from the request
3. (d1) Identification of open ports successful
4. (d1) Recommended the analysis of the SSH keys

Model llama3.2:3b Scoring
1. (r24) Identified the Apache version as outdated and provided a severity of Medium
2. (r24) Identified the PHP version as being outdated and provided a severity of Medium
3. (r26) Correct analysis of some of the ini.php settings
4. (r27) Identified the server banner as not necessary to include in responses
5. (r29) Identified vulnerability to path traversal attacks
6. (r29) Provided a payload for path traversal to ../../../../etc/passwd
7. (r29) Provided recommendation to sanitize user input
8. (r37) Identified the block of text is base64
9. (r37) Suggestion to encrypt a password with bcrypt of pbkdf2
10. (d1) Identification of open ports successful
11. (d1) Recommendation to evaluate the service versions
12. (d1) Recommendation to Analyze SSH keys and verify the authenticity
13. (d2) Provided a python script to decode the base64
14. (d2) Identified it as base64 encoding


