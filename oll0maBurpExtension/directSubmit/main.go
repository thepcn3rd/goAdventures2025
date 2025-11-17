package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"time"
)

type Request struct {
	Model         string    `json:"model"`
	Request       string    `json:"request"`
	Response      string    `json:"response"`
	SystemPrompt  string    `json:"systemPrompt"`
	Timestamp     time.Time `json:"timestamp,omitempty"`
	ID            string    `json:"id,omitempty"`
	APIKey        string    `json:"apiKey,omitempty"`
	RequestNumber string    `json:"requestNumber,omitempty"`
}

type ChatResponse struct {
	Response string `json:"response"`
	Error    string `json:"error,omitempty"`
}

func main() {
	http.HandleFunc("/", homeHandler)
	http.HandleFunc("/api/chat", chatHandler)

	fmt.Println("Server started on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func directSubmitHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		tmpl := template.Must(template.New("form").Parse(htmlForm))
		tmpl.Execute(w, nil)
		return
	}
}

func chatHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodPost {
		http.Error(w, `{"error": "Method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	var req Request
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error": "Invalid JSON"}`, http.StatusBadRequest)
		return
	}

	// Validate required fields
	if req.Model == "" || req.SystemPrompt == "" {
		http.Error(w, `{"error": "Server model and prompt are required"}`, http.StatusBadRequest)
		return
	}

	// Create response struct
	response := ChatResponse{}

	response.Response = "Request processed successfully and saved to output.txt"
	json.NewEncoder(w).Encode(response)
}

const htmlForm = `
<!DOCTYPE html>
<html>
<head>
    <title>Chat API Form</title>
</head>
<body>
    <h1>Oll0ma Direct Submission</h1>
    
    <div>
        Allows a direct submission to Oll0ma without using Burp Suite. The response will be saved to the log files and can be retrieved through Burp.
    </div>
    <hr />
    <br>

    <form id="chatForm">
        <div>
            <label for="server_url">Model:</label><br>
            <input type="text" id="txtModel" name="txtModel" required style="width: 400px;">
        </div>
        <br>
        
        <div>
            <label for="api_key">API Key:</label><br>
            <input type="text" id="txtAPIKey" name="txtAPIKey" style="width: 400px;">
        </div>
        <br>

        <div>
            <label for="lblRequestNumber">Request Number:</label><br>
            <input type="text" id="txtRequestNumber" name="txtRequestNumber" style="width: 400px;">
        </div>
        <br>
        
        <div>
            <label for="system_prompt">System Prompt:</label><br>
            <textarea id="txtSystemPrompt" name="txtSystemPrompt" rows="8" style="width: 600px;"></textarea>
        </div>
        <br>
        
        <div>
            <label for="prompt">Prompt:</label><br>
            <textarea id="txtPrompt" name="txtPrompt" rows="8" style="width: 600px;" required></textarea>
        </div>
        <br>
        
        <button type="submit">Send Request</button>
    </form>
    
    <div id="result" style="margin-top: 20px;"></div>
    
    <script>
        document.getElementById('chatForm').addEventListener('submit', async function(e) {
            e.preventDefault();
            
            const formData = {
                server_url: document.getElementById('server_url').value,
                api_key: document.getElementById('api_key').value,
                system_prompt: document.getElementById('system_prompt').value,
                prompt: document.getElementById('prompt').value
            };
            
            const resultDiv = document.getElementById('result');
            resultDiv.innerHTML = 'Sending request...';
            
            try {
                const response = await fetch('/api/submit', {
                    method: 'POST',
                    headers: {
                        'Content-Type': 'application/json',
                    },
                    body: JSON.stringify(formData)
                });
                
                const data = await response.json();
                
                if (data.error) {
                    resultDiv.innerHTML = '<strong>Error:</strong> ' + data.error;
                } else {
                    resultDiv.innerHTML = '<strong>Response:</strong> ' + data.response + 
                                         '<br><br><strong>Response saved to log files and can be retrieved through Burp</strong>';
                }
            } catch (error) {
                resultDiv.innerHTML = '<strong>Error:</strong> ' + error.message;
            }
        });
    </script>
</body>
</html>
`
