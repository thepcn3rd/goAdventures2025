# Oll0ma Queue Server: Asynchronous LLM Analysis for Burp Suite

**Credit: burpference and Verizon Red Team AI for some of the core concepts that I used to build this Queue Server.**  

## Overview
The oll0ma queue server is built to work with "oll0ma Submit to API" burp extension (oll0maExtension.py).  The extension will prepare information provided to it of an HTTP Request and HTTP Response from Burp to the Queue Server.  The Queue Server will then manage the queue and processing of the information going to a LLM.  The LLM based on the system prompt is supposed to identify weaknesses and vulnerabilities. After the information is processed it is saved on the queue server and then can be viewed in Burp with the "oll0ma View" burp extension (oll0maView.py)

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
![API Stats](/oll0maBurpExtension/apiStats.png)


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