package main

import (
	"bytes"
	"context"
	"crypto/sha256"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"text/template"
	"time"
	"unicode"

	"github.com/gorilla/mux"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Feature List
/**
1. Create a configuration file to store the settings (Completed)
2. Organize the functions better (Completed)
3. Create a round robin system to send requests to multiple Ollama instances (Not going to implement)
4. Sanitize inputs to prevent injection attacks on the API
5. Pull the settings for the extension from the server instead of hardcoding them in the extension (Completed)
6. If request or response sent from the extension is smaller than 30 characters it crashes the queue server (Fixed)
7. Query the Ollama models available and populate a dropdown in the extension (Completed the query portion)
8. Query the files available for system prompts and populate the text area (Completed a different way)
9. Fix the Load Config button to not populate the requests text area (Fixed)
10. Fix the Load config button to not send verbose output to the console (Fixed)
11. Create a message when the job processes reaches the 15 minute timeout with the request log file and the model used (Error message is presented)
12. Create a configuration option to expand the 15 minute timeout to wait on processing a job (Completed)
13. Create a configuration where multiple ollama servers can be retrieved and utilized (Not going to implement)
14. Change the text field of the model to a dropdown populated list (Completed)
15. Change the text field of the ollama server to be a dropdown based on configuration in the event multiple ollamas exist (Retrieving files implemented)
16. Add a system prompt text area that can be saved to a file and retrieved from the server (Completed a different way - learned from burpference)
17. Add a file upload option to send files to the server to be utilized as system prompts (Not going to implement)
18. Create a static page where a prompt can be sent to the queue server unrelated to Burp (with API Key Auth) - Completed
19. Pull the system prompt from the queue server (Completed)
20. Some of the responses contain halucinations try and limit that with using options (Completed)
21. Place the options for the Ollama Server in the config.json file (Completed)
22. Had the JSON incorrect for the options for the ollama API (Fixed)

**/

type Configuration struct {
	QueueServerURL   string             `json:"queueServerURL"`
	APIKey           string             `json:"apiKey"`
	OllamaURL        string             `json:"ollamaURL"`
	WaitTime         int                `json:"waitTime"`
	SystemPromptFile string             `json:"systemPromptFile"`
	ModelOptions     ModelOptionsStruct `json:"modelOptions"`
	HTTPSEnabled     bool               `json:"httpsEnabled"`
	TLSConfig        string             `json:"tlsConfig"`
	TLSCert          string             `json:"tlsCert"`
	TLSKey           string             `json:"tlsKey"`
}

type ModelOptionsStruct struct {
	NumCTX        int     `json:"num_ctx"`        // Default: 2048 - Size of the context window used - The model has a max...
	Temperature   float64 `json:"temperature"`    // Default: 0.8 - 1.0 is more creative to 0.1 conservative text generation
	RepeatLastN   int     `json:"repeat_last_n"`  // How far back to look default 64
	RepeatPenalty float64 `json:"repeat_penalty"` // Repeat is more lenient Default: 1.1 0.9 may be better
	TopK          int     `json:"top_k"`          // Reduces the probability of generating non-sense.  Lower value is conservative Default: 40
	TopP          float64 `json:"top_p"`          // Default: 0.9 - 0.5 is more conservative text generation
}

// Model represents the structure of a model from Ollama API
type Model struct {
	Name       string `json:"name"`
	ModifiedAt string `json:"modified_at"`
	Size       int64  `json:"size"`
	Digest     string `json:"digest"`
}

// TagsResponse represents the response structure from /api/tags
type TagsResponse struct {
	Models []Model `json:"models"`
}

// Request represents the incoming POST request
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

type FileRequest struct {
	APIKey string `json:"apiKey,omitempty"`
}

type ChatResponse struct {
	Response string `json:"response"`
	Error    string `json:"error,omitempty"`
}

type FileDownloadRequest struct {
	APIKey   string `json:"apiKey,omitempty"`
	FileName string `json:"fileName,omitempty"`
}

type OllamaRequestStruct struct {
	Stream       bool               `json:"stream"`
	Messages     []OllamaMessages   `json:"messages"`
	Model        string             `json:"model"`
	ModelOptions ModelOptionsStruct `json:"options,omitempty"`
}

type OllamaResponseStruct struct {
	Model              string         `json:"model"`
	CreatedAt          string         `json:"created_at"`
	Message            OllamaMessages `json:"message"`
	DoneReason         string         `json:"done_reason"`
	Done               bool           `json:"done"`
	TotalDuration      float64        `json:"total_duration"`
	LoadDuration       float64        `json:"load_duration"`
	PromptEvalCount    int            `json:"prompt_eval_count"`
	PromptEvalDuration float64        `json:"prompt_eval_duration"`
	EvalCount          int            `json:"eval_count"`
	EvalDuration       float64        `json:"eval_duration"`
}

type OllamaOutputStruct struct {
	Model                string  `json:"model"`
	CreatedAt            string  `json:"created_at"`
	OriginalSystemPrompt string  `json:"original_system_prompt"`
	OriginalRequest      string  `json:"original_request"`
	OriginalResponse     string  `json:"original_response"`
	ResultMessage        string  `json:"result_message"`
	DoneReason           string  `json:"done_reason"`
	Done                 bool    `json:"done"`
	TotalDuration        float64 `json:"total_duration"`
	LoadDuration         float64 `json:"load_duration"`
	PromptEvalCount      int     `json:"prompt_eval_count"`
	PromptEvalDuration   float64 `json:"prompt_eval_duration"`
	EvalCount            int     `json:"eval_count"`
	EvalDuration         float64 `json:"eval_duration"`
}

type OllamaMessages struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// Job represents a queued job
type Job struct {
	Request Request
	Result  chan string
}

// QueueServer manages the HTTP server and job queue
type QueueServer struct {
	router        *mux.Router
	server        *http.Server
	jobQueue      chan Job
	logger        *zap.Logger
	wg            sync.WaitGroup
	ctx           context.Context
	cancel        context.CancelFunc
	jobCounter    int
	mu            sync.Mutex
	requestNumber string
}

func (c *Configuration) CreateConfig(f string) error {
	var modelOptions ModelOptionsStruct
	modelOptions.NumCTX = 4096
	modelOptions.Temperature = 0.1
	modelOptions.RepeatLastN = 0
	modelOptions.RepeatPenalty = 1.1
	modelOptions.TopK = 10
	modelOptions.TopP = 0.5

	c.QueueServerURL = "https://localhost:8443"
	c.APIKey = ""
	c.OllamaURL = "http://localhost:11434"
	c.WaitTime = 15
	c.SystemPromptFile = "prompts/systemPrompt.txt"
	c.ModelOptions = modelOptions
	c.HTTPSEnabled = true
	c.TLSConfig = "keys/tlsConfig.json"
	c.TLSCert = "keys/tls.crt"
	c.TLSKey = "keys/tls.key"
	jsonData, err := json.MarshalIndent(c, "", "    ")
	if err != nil {
		return err
	}

	err = os.WriteFile(f, jsonData, 0644)
	if err != nil {
		return err
	}

	// Create the System Prompt File
	// Create prompts directory if it doesn't exist
	if err := os.MkdirAll("prompts", 0755); err != nil {
		return nil
	}
	systemPromptDefault := "WW91IGFyZSBhIHdlYiBhcHBsaWNhdGlvbiBwZW5ldHJhdGlvbiB0ZXN0ZXIgcmV2aWV3aW5nICoqb25lIEhUVFAgcmVxdWVzdC9yZXNwb25zZSBwYWlyKiogY2FwdHVyZWQgaW4gQnVycCBTdWl0ZSBQcm94eSBIaXN0b3J5LgoKIyBPYmplY3RpdmUKCklkZW50aWZ5IGFuZCBjbGVhcmx5IGNsYXNzaWZ5IHBvdGVudGlhbCB2dWxuZXJhYmlsaXRpZXMgYXMgKipDcml0aWNhbCwgSGlnaCwgTWVkaXVtLCBvciBMb3cqKiBiYXNlZCBvbiBldmlkZW5jZSBwcmVzZW50IGluIHRoZSByZXF1ZXN0L3Jlc3BvbnNlLiBXaGVyZSBhcHBsaWNhYmxlLCBwcm92aWRlIGEgY29uY2lzZSBQb0MvcGF5bG9hZCBhbmQgcmVtZWRpYXRpb24gZ3VpZGFuY2UuCgojIEhvdyB0byBKdWRnZSBTZXZlcml0eQoKU2NvcmUgZWFjaCB1bmlxdWUgaXNzdWUgdXNpbmcgdGhlc2UgZmFjdG9ycyAoZGVjaWRlIGEgZmluYWwgc2V2ZXJpdHkgbGFiZWwpOgoKKiAqKkltcGFjdDoqKiBEYXRhIGV4cG9zdXJlLCBhY2NvdW50IHRha2VvdmVyLCBjb2RlIGV4ZWN1dGlvbiwgcHJpdmlsZWdlIGVzY2FsYXRpb24uCiogKipFeHBsb2l0YWJpbGl0eToqKiBDYW4gYW4gZXh0ZXJuYWwsIHVuYXV0aGVudGljYXRlZCB1c2VyIGV4cGxvaXQgdGhpcyByZWxpYWJseSBmcm9tIGEgc2luZ2xlIHJlcXVlc3Q/CiogKipTZW5zaXRpdmUgRGF0YToqKiBTZWNyZXRzLCBjcmVkZW50aWFscywgUElJLCB0b2tlbnMsIHNlc3Npb24gSURzLgoqICoqQmxhc3QgUmFkaXVzOioqIEFmZmVjdHMgb25lIHVzZXIsIG1hbnkgdXNlcnMsIG9yIHRoZSB3aG9sZSBzeXN0ZW0/CgpVc2UgdGhpcyBxdWljayBydWJyaWM6CgoqICoqQ3JpdGljYWw6KiogRGlyZWN0IFJDRTsgYXV0aCBieXBhc3MvSURPUiB0byBzZW5zaXRpdmUgZGF0YTsgU1NSRiB0byBjbG91ZCBtZXRhZGF0YTsgU1FMaSBkdW1waW5nIGRhdGE7IHRva2VuL3Nlc3Npb24gdGhlZnQgbGVhZGluZyB0byBhY2NvdW50IHRha2VvdmVyOyB3aWxkY2FyZCtjcmVkZW50aWFscyBDT1JTIG9uIHNlbnNpdGl2ZSBlbmRwb2ludHMuCiogKipIaWdoOioqIFN0b3JlZC9yZWZsZWN0ZWQgWFNTIG9uIHByaXZpbGVnZWQgcGFnZXM7IENTUkYgb24gc3RhdGUtY2hhbmdpbmcgYWN0aW9ucyB3aXRoIG5vIHByb3RlY3Rpb247IGluc2VjdXJlIHNlc3Npb24gZmxhZ3MgZW5hYmxpbmcgaGlqYWNrOyBtaXNjb25maWd1cmVkIGF1dGgvbG9naWMgbGV0dGluZyBwcml2aWxlZ2UgZXNjYWxhdGlvbi4KKiAqKk1lZGl1bToqKiBNaXNzaW5nIG9yIHdlYWsgc2VjdXJpdHkgaGVhZGVycyBvbiBzZW5zaXRpdmUgcmVzcG9uc2VzOyB2ZXJib3NlIGVycm9yIG1lc3NhZ2VzOyBvcGVuIHJlZGlyZWN0IGluIGF1dGggZmxvd3M7IHdlYWsgU2FtZVNpdGU7IHByZWRpY3RhYmxlIElEcy4KKiAqKkxvdzoqKiBJbmZvcm1hdGlvbmFsIGxlYWthZ2UgKHNlcnZlciBiYW5uZXIsIGZyYW1ld29yayB2ZXJzaW9uKSB3aXRob3V0IGltbWVkaWF0ZSBleHBsb2l0OyBtaW5vciBoZWFkZXIgaGFyZGVuaW5nIGdhcHMgb24gbm9uLXNlbnNpdGl2ZSBjb250ZW50LgoKIyBXaGF0IHRvIEV4YW1pbmUgKGZyb20ganVzdCB0aGlzIHBhaXIpCgoxLiAqKlJlcXVlc3QgQW5hbHlzaXMqKgoKICAgKiAqKk1ldGhvZCAmIFRhcmdldDoqKiBVbnNhZmUgdXNlIG9mIGBHRVRgIGZvciBzdGF0ZSBjaGFuZ2VzOyB1bnVzdWFsIG1ldGhvZHMgKGBUUkFDRWAsIGBQVVRgLCBgREVMRVRFYCkgZXhwb3NlZC4KICAgKiAqKkhlYWRlcnM6KiogYEhvc3RgIGhlYWRlciB0cnVzdCAoaG9zdC1oZWFkZXIgaW5qZWN0aW9uKSwgYFgtRm9yd2FyZGVkLUZvcmAgY29udHJvbGxhYmlsaXR5LCBvdmVybHkgcGVybWlzc2l2ZSBgT3JpZ2luYC4KICAgKiAqKkF1dGg6KiogVG9rZW5zL0pXVC9BUEkga2V5cy9CYXNpYyBhdXRoIGluIGhlYWRlcnMgb3IgVVJMOyB0b2tlbiBzY29wZS9leHBpcnkvYWxnIChgbm9uZWAvYEhTMjU2YCB3LyBwdWJsaWMga2V5IGNvbmZ1c2lvbik7IGJlYXJlciB0b2tlbnMgaW4gcXVlcnkgc3RyaW5nLgogICAqICoqUGFyYW1zICYgQm9keToqKiBVbnNhbml0aXplZCB1c2VyIGlucHV0IGhpbnRzIChTUUxpLCBjb21tYW5kL3RlbXBsYXRlIGluamVjdGlvbiwgcGF0aCB0cmF2ZXJzYWwgYC4uL2AsIFhNTCBpbiBgQ29udGVudC1UeXBlOiBhcHBsaWNhdGlvbi94bWxgIOKGkiBYWEUpLCBmaWxlIHVwbG9hZHMgd2l0aCB3ZWFrIHZhbGlkYXRpb24uCiAgICogKipDU1JGIFNpZ25hbHM6KiogU3RhdGUtY2hhbmdpbmcgcmVxdWVzdCBsYWNraW5nIENTUkYgdG9rZW4sIG9yIHRva2VuIG5vdCB0aWVkIHRvIHNlc3Npb24uCjIuICoqUmVzcG9uc2UgQW5hbHlzaXMqKgoKICAgKiAqKlN0YXR1cy9Cb2R5OioqIFN0YWNrIHRyYWNlcywgREIgZXJyb3JzLCBkZWJ1ZyBmbGFncywgc2VjcmV0cyAoa2V5cywgVVJMcywgY3JlZGVudGlhbHMpLCB1c2VyIG9iamVjdHMgZnJvbSBvdGhlciBhY2NvdW50cyAoSURPUikuCiAgICogKipDT1JTOioqIGBBY2Nlc3MtQ29udHJvbC1BbGxvdy1PcmlnaW46ICpgIHdpdGggYC4uLi1DcmVkZW50aWFsczogdHJ1ZWAgKG9yIHJlZmxlY3RlZCBvcmlnaW5zIG9uIHNlbnNpdGl2ZSBlbmRwb2ludHMpLgogICAqICoqU2VjdXJpdHkgSGVhZGVyczoqKiBNaXNzaW5nL3dlYWsgYENvbnRlbnQtU2VjdXJpdHktUG9saWN5YCwgYFN0cmljdC1UcmFuc3BvcnQtU2VjdXJpdHlgLCBgWC1Db250ZW50LVR5cGUtT3B0aW9uc2AsIGBYLUZyYW1lLU9wdGlvbnNgL2BQZXJtaXNzaW9ucy1Qb2xpY3lgLCBgUmVmZXJyZXItUG9saWN5YCwgYENyb3NzLU9yaWdpbi1PcGVuZXItUG9saWN5YC4KICAgKiAqKkNhY2hpbmc6KiogU2Vuc2l0aXZlIGNvbnRlbnQgd2l0aG91dCBgQ2FjaGUtQ29udHJvbDogbm8tc3RvcmVgIC8gYFByYWdtYTogbm8tY2FjaGVgLgogICAqICoqQ29udGVudC1UeXBlIHZzIEJvZHk6KiogTWlzbWF0Y2hlcyBlbmFibGluZyBYU1Mgb3IgTUlNRS1zbmlmZmluZzsgSlNPTiBlbmRwb2ludHMgbWlzc2luZyBgYXBwbGljYXRpb24vanNvbmAuCjMuICoqQ29va2llcyAoZnJvbSBgU2V0LUNvb2tpZWApKioKCiAgICogTWlzc2luZyBgU2VjdXJlYCBvbiBIVFRQUzsgbWlzc2luZyBgSHR0cE9ubHlgOyBpbmFwcHJvcHJpYXRlIGBTYW1lU2l0ZWAgKGBOb25lYCB3aXRob3V0IGBTZWN1cmVgLCBvciBsYXggb24gYXV0aCBjb29raWVzKTsgb3Zlcmx5IGJyb2FkIGBEb21haW5gL2BQYXRoYDsgbG9uZy1saXZlZCBzZXNzaW9uIGNvb2tpZXM7IGR1cGxpY2F0ZSBvciBjb25mbGljdGluZyBjb29raWVzIChmaXhhdGlvbiByaXNrcykuCgojIE91dHB1dCBGb3JtYXQgKGJlIGNvbmNpc2UgYW5kIGRldGVybWluaXN0aWMpCgoqKlN1bW1hcnk6KiogMeKAkzIgc2VudGVuY2Ugb3ZlcnZpZXcuCioqT3ZlcmFsbCBTZXZlcml0eToqKiBIaWdoZXN0IHNldmVyaXR5IGZvdW5kIChDcml0aWNhbC9IaWdoL01lZGl1bS9Mb3cvTm9uZSkuCgoqKkZpbmRpbmdzIChvbmUgcGVyIHJvdyk6KioKCiogKipTZXZlcml0eToqKiBDcml0aWNhbCB8IEhpZ2ggfCBNZWRpdW0gfCBMb3cKKiAqKkNhdGVnb3J5OioqIGUuZy4sIEF1dGgvU2Vzc2lvbiwgSW5qZWN0aW9uLCBYU1MsIENTUkYsIENPUlMsIEhlYWRlcnMsIENvb2tpZXMsIExvZ2ljLCBJbmZvTGVhaywgQ2FjaGluZwoqICoqRW5kcG9pbnQgJiBNZXRob2Q6KiogYC9wYXRoYCwgYEdFVC9QT1NUYAoqICoqRXZpZGVuY2U6KiogUXVvdGUgZXhhY3QgaGVhZGVyL2JvZHkgc25pcHBldHMgb3IgcGFyYW1ldGVyIG5hbWVzIChtaW5pbWl6ZTsgcmVkYWN0IG9idmlvdXMgc2VjcmV0cykuCiogKipSaXNrOioqIFNob3J0IGltcGFjdCBzdGF0ZW1lbnQuCiogKipQb0MgLyBQYXlsb2FkOioqIE1pbmltYWwgcmVwcm9kdWNpYmxlIGV4YW1wbGUgKHNlZSB0ZW1wbGF0ZXMgYmVsb3cpLgoqICoqUmVtZWRpYXRpb246KiogQ29uY3JldGUgZml4LgoKKipJZiBObyBJc3N1ZXM6KiogU3RhdGUg4oCcTm8gaXNzdWVzIGV2aWRlbnQgaW4gdGhpcyBwYWlyLOKAnSBhbmQgYnJpZWZseSBub3RlIHdoYXQgeW91IHZlcmlmaWVkLgoKIyBQYXlsb2FkIC8gUG9DIFRlbXBsYXRlcyAodXNlIG9ubHkgaWYgaW5kaWNhdGVkIGJ5IGV2aWRlbmNlKQoKKiAqKlJlZmxlY3RlZCBYU1M6KiogYCI+PHN2ZyBvbmxvYWQ9YWxlcnQoMSk+YCBvciBgPC9zY3JpcHQ+PGltZyBzcmM9eCBvbmVycm9yPWFsZXJ0KDEpPmAgaW4gcmVmbGVjdGVkIHBhcmFtZXRlcjsgdmVyaWZ5IGNvbnRleHQgKEhUTUwsIGF0dHIsIEpTKS4KKiAqKlN0b3JlZCBYU1MgKHByZXZpZXcgaW4gcmVzcG9uc2UpOioqIFBvc3Qgc2FtZSBwYXlsb2FkIHRvIGEgZmllbGQgc2hvd24gaW4gbGF0ZXIgcmVzcG9uc2VzLgoqICoqU1FMaSAoYm9vbGVhbi90aW1lLWJhc2VkKToqKiBgJyBPUiAnMSc9JzFgICB8ICBgaWQ9MSBBTkQgU0xFRVAoNSlgIChsb29rIGZvciBEQiBlcnJvcnMvZGVsYXlzKS4KKiAqKlNTUkY6KiogUG9pbnQgcGFyYW1ldGVycyBsaWtlIGB1cmw9YCB0byBgaHR0cDovLzE2OS4yNTQuMTY5LjI1NC9sYXRlc3QvbWV0YS1kYXRhL2Agb3IgY29sbGFib3JhdG9yIFVSTC4KKiAqKlBhdGggVHJhdmVyc2FsOioqIGAuLi8uLi8uLi8uLi9ldGMvcGFzc3dkYCBvciBgJTJlJTJlL2AgZW5jb2RpbmdzIGluIGZpbGUvcGF0aCBwYXJhbXMuCiogKipPcGVuIFJlZGlyZWN0OioqIGBuZXh0PWh0dHBzOi8vYXR0YWNrZXIudGxkYCBhbmQgY29uZmlybSAzeHggdG8gZXh0ZXJuYWwgZG9tYWluLgoqICoqQ1NSRjoqKiBTdGF0ZS1jaGFuZ2luZyBQT1NUIHdpdGhvdXQgQ1NSRiB0b2tlbjsgZGVtb25zdHJhdGUgY3Jvc3Mtc2l0ZSBmb3JtIGF1dG8tc3VibWl0LgoqICoqSG9zdCBIZWFkZXIgSW5qZWN0aW9uOioqIFNlbmQgY3VzdG9tIGBIb3N0OiBhdHRhY2tlci50bGRgIGFuZCBvYnNlcnZlIGFic29sdXRlIFVSTHMvcGFzc3dvcmQgcmVzZXQgbGlua3MuCgojIEdyYWRpbmcgRXhhbXBsZXMgKG1hcCB0byB0aGUgcnVicmljKQoKKiAqKkNyaXRpY2FsOioqIGBTZXQtQ29va2llOiBzZXNzaW9uPS4uLjsgSHR0cE9ubHlgIG1pc3NpbmcgKyByZWZsZWN0ZWQgYm9keSBwcmludHMgc2Vzc2lvbiDihpIgdGhlZnQg4oaSIGFjY291bnQgdGFrZW92ZXIuCiogKipIaWdoOioqIGBBY2Nlc3MtQ29udHJvbC1BbGxvdy1DcmVkZW50aWFsczogdHJ1ZWAgd2l0aCByZWZsZWN0ZWQgYEFjY2Vzcy1Db250cm9sLUFsbG93LU9yaWdpbmAgb24gYC9hcGkvdXNlcmAgcmV0dXJucyBQSUkuCiogKipNZWRpdW06KiogTG9naW4gcmVzcG9uc2UgbWlzc2luZyBgSFNUU2AgYW5kIGBDU1BgOyB2ZXJib3NlIHN0YWNrIHRyYWNlIHJldmVhbHMgaW50ZXJuYWwgcGF0aHMuCiogKipMb3c6KiogYFNlcnZlcjogbmdpbngvMS4xOC4wYCBvbmx5OyBub24tc2Vuc2l0aXZlIGVuZHBvaW50IG1pc3NpbmcgYFJlZmVycmVyLVBvbGljeWAuCgojIE5vdGVzCgoqIE9ubHkganVkZ2Ugd2hhdCBpcyB2aXNpYmxlIGluIHRoaXMgc2luZ2xlIHBhaXI7IGlmIG1vcmUgZXZpZGVuY2UgaXMgcmVxdWlyZWQsIHNheSDigJxuZWVkcyBtdWx0aS1yZXF1ZXN0IHZlcmlmaWNhdGlvbi7igJ0KKiBLZWVwIGV4cGxhbmF0aW9ucyBjb21wYWN0OyBwcmlvcml0aXplIGNvbmNyZXRlIGV2aWRlbmNlLgoqIFJlZGFjdCBzZWNyZXRzIGluIGV4YW1wbGVzIHdoaWxlIHByb3ZpbmcgdGhlIHBvaW50LgoKLS0tCgoqKlRoZSBIVFRQIHJlcXVlc3QgYW5kIHJlc3BvbnNlIHBhaXIgYXJlIHByb3ZpZGVkIGJlbG93IHRoaXMgbGluZToqKgoK"
	CreateFileFromB64(systemPromptDefault, "prompts/systemPrompt.txt")

	return nil
}

func (c *Configuration) SaveConfig(f string) error {
	jsonData, err := json.MarshalIndent(c, "", "    ")
	if err != nil {
		return err
	}

	err = os.WriteFile(f, jsonData, 0644)
	if err != nil {
		return err
	}

	return nil
}

func (c *Configuration) LoadConfig(cPtr string) error {
	configFile, err := os.Open(cPtr)
	if err != nil {
		return err
	}
	defer configFile.Close()
	decoder := json.NewDecoder(configFile)
	if err := decoder.Decode(&c); err != nil {
		return err
	}

	return nil
}

// NewQueueServer creates a new queue server instance
func NewQueueServer(addr string, queueSize int, requestNumber string) (*QueueServer, error) {
	// Create context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())

	// Initialize logger
	logger, err := initLogger()
	if err != nil {
		cancel()
		return nil, fmt.Errorf("failed to initialize logger: %w", err)
	}

	qs := &QueueServer{
		router:        mux.NewRouter(),
		jobQueue:      make(chan Job, queueSize),
		logger:        logger,
		ctx:           ctx,
		cancel:        cancel,
		requestNumber: requestNumber,
	}

	// Configure routes
	qs.setupRoutes()

	// Configure HTTPS server
	qs.server = &http.Server{
		Addr:    addr,
		Handler: qs.router,
		TLSConfig: &tls.Config{
			MinVersion: tls.VersionTLS12,
		},
	}

	return qs, nil
}

// initLogger sets up structured logging with both console and file output
func initLogger() (*zap.Logger, error) {
	// Create logs directory if it doesn't exist
	if err := os.MkdirAll("logs", 0755); err != nil {
		return nil, err
	}

	// File encoder config
	fileEncoderConfig := zap.NewProductionEncoderConfig()
	fileEncoderConfig.TimeKey = "timestamp"
	fileEncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	// Console encoder config
	consoleEncoderConfig := zap.NewDevelopmentEncoderConfig()
	consoleEncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder

	// File writer
	timestamp := time.Now().Format("2006-01-02")
	fileName := fmt.Sprintf("logs/app_%s.log", timestamp)
	logFile, err := os.OpenFile(fileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, err
	}

	// Create cores
	fileCore := zapcore.NewCore(
		zapcore.NewJSONEncoder(fileEncoderConfig),
		zapcore.AddSync(logFile),
		zap.InfoLevel,
	)

	consoleCore := zapcore.NewCore(
		zapcore.NewConsoleEncoder(consoleEncoderConfig),
		zapcore.AddSync(os.Stdout),
		zap.InfoLevel,
	)

	// Combine cores
	core := zapcore.NewTee(fileCore, consoleCore)

	return zap.New(core), nil
}

// setupRoutes configures the HTTP routes
func (qs *QueueServer) setupRoutes() {
	qs.router.HandleFunc("/api/submit", qs.handlePostJob).Methods("POST")
	qs.router.HandleFunc("/api/health", qs.handleHealthCheck).Methods("GET")
	qs.router.HandleFunc("/api/stats", qs.handleStats).Methods("GET")
	qs.router.HandleFunc("/api/files", qs.handleFileList).Methods("POST")
	qs.router.HandleFunc("/api/file", qs.handleFileDownload).Methods("POST")
	qs.router.HandleFunc("/api/loadconfig", qs.handleLoadConfig).Methods("POST")
	qs.router.HandleFunc("/api/direct.html", qs.handleDirectSubmit).Methods("GET")
	qs.router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if err := os.MkdirAll("static", 0755); err != nil {
			return
		}
		createIndexHTML("/static/index.html")
		http.FileServer(http.Dir("./static")).ServeHTTP(w, r)
	}).Methods("GET")
	//qs.router.HandleFunc("/chat.html", qs.submitChat).Methods("GET")
}

func (qs *QueueServer) handleDirectSubmit(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		tmpl := template.Must(template.New("form").Parse(htmlForm))
		tmpl.Execute(w, nil)
		return
	}
}

func (qs *QueueServer) handleLoadConfig(w http.ResponseWriter, r *http.Request) {
	var req FileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		qs.logger.Error("Failed to decode request", zap.Error(err))
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if req.APIKey != config.APIKey {
		qs.logger.Warn("Unauthorized access attempt (Invalid API Key)")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Fetch model names from Ollama server
	modelNames, err := GetModelNames()
	if err != nil {
		qs.logger.Error("Failed to get model names", zap.Error(err))
		http.Error(w, "Failed to get model names", http.StatusInternalServerError)
		return
	}

	// Read the file systemPrompt.txt from the directory
	var systemPrompt string
	systemPrompt, err = ReadFile(config.SystemPromptFile)
	if err != nil {
		qs.logger.Error("Failed to read system prompt file: %s", zap.Error(err))
		// If an error exists reading the systemPrompt File then the below is default
		systemPrompt = `You are a web application penetration tester conducting a comprehensive assessment to find logic flaws and vulnerabilities then to leverage the security flaws to demonstrate risk.
        
Your objective is to examine the HTTP requests and responses that are available through the burp suite proxy history from the web application.
        
This analysis will focus on:
- Request and Response Evaluation: Scrutinizing HTTP requests and responses for security misconfigurations, sensitive data exposure, and other vulnerabilities.
- Authentication and Session Management: Assessing the effectiveness of authentication mechanisms and session handling practices.
- Input Validation and Output Encoding: Identifying weaknesses related to input validation that may lead to injection attacks or cross-site scripting (XSS).
- Cookie Evaluation: Identify the attributes of HTTP cookies and the attributes that are missing for security
        
Use reasoning and context to find potential flaws in the HTTP request and response provided. Provide example payloads and PoCs that could lead to a demonstration of the vulnerbaility.

Not every request and response will have vulnerabilities, be concise yet deterministic.

The HTTP request and and response pair are provided below this line:
`
	}

	//fmt.Println(modelNames)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"ollamaURL":    config.OllamaURL,
		"models":       modelNames,
		"systemPrompt": systemPrompt,
	})
	qs.logger.Info("Configuration Retrieved")

}

// GetOllamaModels fetches the list of available models from Ollama server
func GetOllamaModels() ([]Model, error) {
	//url := fmt.Sprintf("http://%s:%d/api/tags", host, port)
	url := config.OllamaURL + "/api/tags"
	// Create HTTP client with timeout
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	// Create GET request
	resp, err := client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Ollama server: %w", err)
	}
	defer resp.Body.Close()

	// Check status code
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("server returned non-200 status: %s", resp.Status)
	}

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	//fmt.Println("Response from Ollama /api/tags endpoint:")
	//fmt.Println(string(body))

	// Parse JSON response
	var tagsResponse TagsResponse
	err = json.Unmarshal(body, &tagsResponse)
	if err != nil {
		return nil, fmt.Errorf("failed to parse JSON response: %w", err)
	}

	return tagsResponse.Models, nil
}

// GetModelNames returns just the model names as a string slice
func GetModelNames() ([]string, error) {
	models, err := GetOllamaModels()
	if err != nil {
		return nil, err
	}

	names := make([]string, len(models))
	for i, model := range models {
		names[i] = model.Name
	}

	return names, nil
}

func (qs *QueueServer) handleFileDownload(w http.ResponseWriter, r *http.Request) {
	var req FileDownloadRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		qs.logger.Error("Failed to decode request", zap.Error(err))
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	if req.APIKey != config.APIKey {
		qs.logger.Warn("Unauthorized access attempt (Invalid API Key)")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	filePath := "output/" + req.FileName
	file, err := os.Open(filePath)
	if err != nil {
		qs.logger.Error("Failed to open file", zap.String("file", req.FileName), zap.Error(err))
		http.Error(w, "File not found", http.StatusNotFound)
		return
	}
	defer file.Close()

	// Read the entire file content
	data, err := io.ReadAll(file)
	if err != nil {
		panic(err)
	}

	// Encode to base64 string
	encodedFile := base64.StdEncoding.EncodeToString(data)
	// Create a sha256 hash of the encoded filename
	hash := sha256.Sum256([]byte(encodedFile))

	// return the file in json format
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"encodedFile": encodedFile,
		"hash":        fmt.Sprintf("%x", hash),
	})
	qs.logger.Info("File retrieved", zap.String("hash", fmt.Sprintf("%x", hash)))

}

// handleFileList returns a list of files in the output directory
func (qs *QueueServer) handleFileList(w http.ResponseWriter, r *http.Request) {
	var req FileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		qs.logger.Error("Failed to decode request", zap.Error(err))
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if req.APIKey != config.APIKey {
		qs.logger.Warn("Unauthorized access attempt (Invalid API Key)")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	files, err := os.ReadDir("output")
	if err != nil {
		qs.logger.Error("Failed to read output directory", zap.Error(err))
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	var fileNames []string
	for _, file := range files {
		if !file.IsDir() {
			fileNames = append(fileNames, file.Name())
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"files": fileNames,
	})
	qs.logger.Info("File list retrieved", zap.Int("file_count", len(fileNames)))
}

// handlePostJob processes incoming POST requests
func (qs *QueueServer) handlePostJob(w http.ResponseWriter, r *http.Request) {
	var req Request
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		qs.logger.Error("Failed to decode request", zap.Error(err))
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	//fmt.Println("Received API Key: " + req.APIKey)
	//fmt.Println("Expected API Key: " + config.APIKey)
	if req.APIKey != config.APIKey {
		qs.logger.Warn("Unauthorized access attempt (Invalid API Key)")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Generate ID for job
	req.ID = generateID()
	req.Timestamp = time.Now()

	// Create job with result channel
	resultChan := make(chan string, 1)
	job := Job{
		Request: req,
		Result:  resultChan,
	}

	// Try to add to queue with timeout
	select {
	case qs.jobQueue <- job:
		var ollamaRequest []byte
		var err error
		if job.Request.Response == "333350adF" { // Mobile POST
			ollamaRequest = []byte(job.Request.Request)
		} else {
			ollamaRequest, err = base64.StdEncoding.DecodeString(job.Request.Request)
			if err != nil {
				qs.logger.Error("Failed to decode request", zap.Error(err))
				job.Result <- fmt.Sprintf("Error decoding request: %v", err)
				return
			}
		}
		ollamaRequestString := FilterAlphanumericWithSpaces(string(ollamaRequest))
		if len(ollamaRequestString) <= 30 {
			ollamaRequestString = ollamaRequestString + "_" + GenerateRandomString(30-len(ollamaRequestString))
		}
		qs.logger.Info("Job queued successfully",
			zap.String("job_id", req.ID),
			zap.Any("data", ollamaRequestString[:30]),
			zap.String("model", req.Model),
		)
		// Wait for processing result with timeout
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusAccepted)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"id":        req.ID,
			"status":    "queued",
			"message":   "Job accepted",
			"timestamp": time.Now(),
		})

	case <-time.After(5 * time.Second):
		qs.logger.Error("Queue full", zap.String("job_id", req.ID))
		http.Error(w, "Queue full", http.StatusServiceUnavailable)
	}
}

// handleHealthCheck returns server status
func (qs *QueueServer) handleHealthCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":     "healthy",
		"timestamp":  time.Now(),
		"queue_size": len(qs.jobQueue),
		"capacity":   cap(qs.jobQueue),
	})
}

// handleStats returns queue statistics
func (qs *QueueServer) handleStats(w http.ResponseWriter, r *http.Request) {
	qs.mu.Lock()
	defer qs.mu.Unlock()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"total_jobs_processed": qs.jobCounter,
		"current_queue_size":   len(qs.jobQueue),
		"queue_capacity":       cap(qs.jobQueue),
		"timestamp":            time.Now(),
	})
}

// processJobs handles job processing in the background
func (qs *QueueServer) processJobs() {
	qs.wg.Add(1)
	defer qs.wg.Done()

	qs.logger.Info("Starting job processor")

	for {
		select {
		case job := <-qs.jobQueue:
			qs.processJob(job)
		case <-qs.ctx.Done():
			qs.logger.Info("Job processor shutting down")
			return
		}
	}
}

// processJob simulates job processing
func (qs *QueueServer) processJob(job Job) {
	var err error
	startTime := time.Now()

	if config.APIKey != job.Request.APIKey {
		qs.logger.Warn("Invalid API Key with Submission", zap.String("job_id", job.Request.ID))
		job.Result <- "Unauthorized"
		return
	}

	var ollamaRequest []byte
	if job.Request.Response == "333350adF" { // Mobile POST
		ollamaRequest = []byte(job.Request.Request)
	} else {
		ollamaRequest, err = base64.StdEncoding.DecodeString(job.Request.Request)
		if err != nil {
			qs.logger.Error("Failed to decode request", zap.Error(err))
			job.Result <- fmt.Sprintf("Error decoding request: %v", err)
			return
		}
	}

	ollamaRequestString := FilterAlphanumericWithSpaces(string(ollamaRequest))
	if len(ollamaRequestString) <= 30 {
		ollamaRequestString = ollamaRequestString + "_" + GenerateRandomString(30-len(ollamaRequestString))
	}

	qs.logger.Info("Processing job started",
		zap.String("job_id", job.Request.ID),
		zap.Any("data", ollamaRequestString[:30]))

	//fmt.Println(job.Request.Model)
	var oReq OllamaRequestStruct
	oReq.Stream = false

	var ollamaModel []byte
	if job.Request.Response == "333350adF" { // Mobile POST
		ollamaModel = []byte(job.Request.Model)
	} else {
		ollamaModel, err = base64.StdEncoding.DecodeString(job.Request.Model)
		if err != nil {
			qs.logger.Error("Failed to decode model", zap.Error(err))
			job.Result <- fmt.Sprintf("Error decoding model: %v", err)
			return
		}
	}
	oReq.Model = string(ollamaModel)

	var ollamaSystemPrompt = []byte{}
	if job.Request.Response == "333350adF" { // Mobile POST
		ollamaSystemPrompt = []byte(job.Request.SystemPrompt)
	} else {
		ollamaSystemPrompt, err = base64.StdEncoding.DecodeString(job.Request.SystemPrompt)
		if err != nil {
			qs.logger.Error("Failed to decode system prompt", zap.Error(err))
			job.Result <- fmt.Sprintf("Error decoding system prompt: %v", err)
			return
		}
	}

	var ollamaResponse = []byte{}
	if job.Request.Response == "333350adF" { // Mobile POST
		ollamaResponse = []byte("22222")
		job.Request.RequestNumber = "_mobile_" + job.Request.RequestNumber
	} else {
		ollamaResponse, err = base64.StdEncoding.DecodeString(job.Request.Response)
		if err != nil {
			qs.logger.Error("Failed to decode response", zap.Error(err))
			job.Result <- fmt.Sprintf("Error decoding response: %v", err)
			return
		}
	}
	if string(ollamaResponse) == "22222" {
		ollamaResponse = []byte("")
	}

	var roleSystem OllamaMessages
	roleSystem.Role = "system"
	roleSystem.Content = string(ollamaSystemPrompt)

	var roleUser OllamaMessages
	roleUser.Role = "user"
	roleUser.Content = fmt.Sprintf("%s\n\n%s", string(ollamaRequest), string(ollamaResponse))

	oReq.Messages = append(oReq.Messages, roleSystem)
	oReq.Messages = append(oReq.Messages, roleUser)

	// Setup the Model Options
	/** Moved to the config
	var modelOptions ModelStruct
	modelOptions.NumCTX = config.ModelOptions.NumCTX               // Context window size - Setting to 4096 for more tokens to be used, Default: 2048
	modelOptions.Temperature = config.ModelOptions.Temperature     // I am not looking for creativity - 0.1 is conservative, Default: 0.8
	modelOptions.RepeatLastN = config.ModelOptions.RepeatLastN     // How far back to look - Default: 64, Setting to 0 disabled to disable the look back
	modelOptions.RepeatPenalty = config.ModelOptions.RepeatPenalty // At this time leaving this alone
	modelOptions.TopK = config.ModelOptions.TopK                   // Probability of generating non-sense - Lowering this to 10, Default: 40
	modelOptions.TopP = config.ModelOptions.TopP                   // Focused and conservative is 0.5, Default: 0.9
	**/
	oReq.ModelOptions = config.ModelOptions

	jsonData, err := json.Marshal(oReq)
	if err != nil {
		qs.logger.Error("Failed to marshal Ollama request", zap.Error(err))
		job.Result <- fmt.Sprintf("Error marshalling request: %v", err)
		return
	}

	response, err := sendToOllama([]byte(jsonData), qs, job)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	var ollamaResp OllamaResponseStruct
	err = json.Unmarshal(response, &ollamaResp)
	if err != nil {
		qs.logger.Error("Failed to unmarshal Ollama response", zap.Error(err))
		job.Result <- fmt.Sprintf("Error unmarshalling response: %v", err)
		return
	}

	var ollamaOutput OllamaOutputStruct
	ollamaOutput.Model = ollamaResp.Model
	ollamaOutput.CreatedAt = ollamaResp.CreatedAt
	ollamaOutput.OriginalSystemPrompt = string(ollamaSystemPrompt)
	ollamaOutput.OriginalRequest = string(ollamaRequest)
	ollamaOutput.OriginalResponse = string(ollamaResponse)
	ollamaOutput.ResultMessage = ollamaResp.Message.Content
	ollamaOutput.DoneReason = ollamaResp.DoneReason
	ollamaOutput.Done = ollamaResp.Done
	ollamaOutput.TotalDuration = ollamaResp.TotalDuration
	ollamaOutput.LoadDuration = ollamaResp.LoadDuration
	ollamaOutput.PromptEvalCount = ollamaResp.PromptEvalCount
	ollamaOutput.PromptEvalDuration = ollamaResp.PromptEvalDuration
	ollamaOutput.EvalCount = ollamaResp.EvalCount
	ollamaOutput.EvalDuration = ollamaResp.EvalDuration

	ollamaResponseOutput, err := json.MarshalIndent(ollamaOutput, "", "    ")
	if err != nil {
		qs.logger.Error("Failed to marshal Ollama output", zap.Error(err))
		job.Result <- fmt.Sprintf("Error marshalling output: %v", err)
		return
	}

	//fmt.Println("Response from Ollama:", string(response))
	// Create output directory if it doesn't exist
	if err := os.MkdirAll("output", 0755); err != nil {
		return
	}
	// Save the response to a file
	timestamp := time.Now().Format("2006-01-02_150405")
	outputFileName := fmt.Sprintf("output/results_req%s_%s_%s", job.Request.RequestNumber, timestamp, job.Request.ID)
	SaveOutputFile(string(ollamaResponseOutput), outputFileName)

	// Send the jsonData to the Ollama API
	qs.logger.Info("Sending request to Ollama API",
		zap.String("job_id", job.Request.ID),
		zap.String("ollama_request", ollamaRequestString[:30]))

	// Simulate a delay for processing
	time.Sleep(1 * time.Second)

	// Update statistics
	qs.mu.Lock()
	qs.jobCounter++
	qs.mu.Unlock()

	processingDuration := time.Since(startTime)

	result := fmt.Sprintf("Processed in %v ID: %v Request: %s", processingDuration, job.Request.ID, ollamaRequestString[:30])

	qs.logger.Info("Job processing completed",
		zap.String("job_id", job.Request.ID),
		zap.Duration("processing_time", processingDuration),
		zap.String("result", result))

	// Send result back to HTTP handler
	select {
	case job.Result <- result:
	case <-time.After(1 * time.Second):
		qs.logger.Warn("Failed to send result - channel timeout",
			zap.String("job_id", job.Request.ID))
	}

	// Wait for 1 minute between jobs for server cool down...
	time.Sleep(1 * time.Minute)
}

// Start begins the HTTP server and job processing
func (qs *QueueServer) Start() error {
	// Start job processor
	go qs.processJobs()

	if config.HTTPSEnabled {
		qs.logger.Info("Starting HTTPS server", zap.String("address", qs.server.Addr))
		if err := qs.server.ListenAndServeTLS(config.TLSCert, config.TLSKey); err != nil && err != http.ErrServerClosed {
			return fmt.Errorf("failed to start server: %w", err)
		}
	} else {
		qs.logger.Info("Starting HTTP server", zap.String("address", qs.server.Addr))
		if err := qs.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			return fmt.Errorf("failed to start server: %w", err)
		}
	}

	return nil
}

// Stop gracefully shuts down the server
func (qs *QueueServer) Stop() error {
	qs.logger.Info("Initiating graceful shutdown")

	qs.cancel() // Signal job processor to stop

	// Shutdown HTTP server with timeout
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := qs.server.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("server shutdown failed: %w", err)
	}

	// Wait for job processor to finish
	qs.wg.Wait()
	qs.logger.Info("Server shutdown completed")

	return qs.logger.Sync()
}

// generateID creates a unique job ID
func generateID() string {
	return fmt.Sprintf("job_%d_%d", time.Now().Unix(), time.Now().UnixNano()%1000)
}

func FilterAlphanumericWithSpaces(input string) string {
	var result []rune

	for _, char := range input {
		if unicode.IsLetter(char) || unicode.IsDigit(char) || unicode.IsSpace(char) {
			result = append(result, char)
		}
	}

	return string(result)
}

func sendToOllama(jsonData []byte, qs *QueueServer, j Job) ([]byte, error) {
	// Create HTTP request
	ollamaChatURL := config.OllamaURL + "/api/chat"
	req, err := http.NewRequest("POST", ollamaChatURL,
		bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{
		Timeout: time.Duration(config.WaitTime) * time.Minute,
	}

	// Send request
	resp, err := client.Do(req)
	if err != nil {
		qs.logger.Error("error sending request (waiting the max wait time)",
			zap.String("job_id", j.Request.ID),
			zap.Error(err),
		)
		return nil, fmt.Errorf("error sending request (waited %d minutes): %w", config.WaitTime, err)
	}
	defer resp.Body.Close()

	// Check status code
	var body []byte
	if resp.StatusCode != http.StatusOK {
		body, _ = io.ReadAll(resp.Body)
		return nil, fmt.Errorf("HTTP error %d: %s", resp.StatusCode, string(body))
	} else {
		body, err = io.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("error reading response body: %w", err)
		}
	}

	return body, nil
}

func createIndexHTML(folderDir string) {
	currentDir, _ := os.Getwd()
	newDir := currentDir + folderDir
	//cf.CheckError("Unable to get the working directory", err, true)
	if _, err := os.Stat(newDir); errors.Is(err, os.ErrNotExist) {
		// Output to File - Overwrites if file exists...
		f, err := os.Create(newDir)
		if err != nil {
			fmt.Println("Error creating index.html file:", err)
			return
		}
		defer f.Close()
		html := "<h2>Ollama Queue Server</h2><br />\n"
		html += "<h3>Server is Running...</h3>\n"
		html += "<p>To use the server, operate it through the Burp Suite Extension. You can check the details of its request processing in the logs folder<br />and view the final results in the output folder. If you need to change how the server works, modify the config.json file and then restart<br />the server for your changes to take effect. This tool was developed by thepcn3rd.</p>"
		html += "<p>Developed by thepcn3rd</p><br /><br />"
		html += "<strong>Direct Submission Page:</strong> <a href='/api/direct.html' target='_blank'>/api/direct.html</a><br /><br />"
		html += "<table cellpadding='10'>"
		html += "<tr valign='top'><td><strong>/api/health</strong> - GET - Health Check Endpoint<br />"
		html += "<strong>/api/stats</strong> - GET - Queue Statistics Endpoint<br />"
		html += "<strong>/api/loadconfig</strong> - POST - Load Configuration for the Extension</td>"
		html += "<td><strong>/api/submit</strong> - POST - Submit a Job to the Queue<br />"
		html += "<strong>/api/files</strong> - POST - List Files in the Output Directory<br />\n"
		html += "<strong>/api/file</strong> - POST - Download a File from the Output Directory</td></tr>\n"
		html += "</table>"
		f.Write([]byte(headerHTML()))
		f.Write([]byte(html))
		f.Write([]byte(tailHTML()))
		f.Close()
	}
}

func headerHTML() string {
	hHTML := `<!DOCTYPE html>
			  <html lang="en">
  			  <head>
    			<meta charset="UTF-8" />
    			<meta name="viewport" content="width=device-width, initial-scale=1.0" />
    			<meta http-equiv="X-UA-Compatible" content="ie=edge" />
  			  </head>
  			  <body>`
	return hHTML
}

func tailHTML() string {
	tHTML := "</body></html>"
	return tHTML
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
            <textarea id="txtSystemPrompt" name="txtSystemPrompt" rows="8" style="width: 600px;">You are a security analyst providing answers to questions in simple summaries but with enough detail to answer the question.</textarea>
        </div>
        <br>
        
        <div>
            <label for="prompt">Prompt:</label><br>
            <textarea id="txtPrompt" name="txtPrompt" rows="8" style="width: 600px;" required>What are the different types of XSS that I should test for?</textarea>
        </div>
        <br>
        
        <button type="submit">Send Request</button>
    </form>
    
    <div id="result" style="margin-top: 20px;"></div>
    
    <script>
        document.getElementById('chatForm').addEventListener('submit', async function(e) {
            e.preventDefault();
            
            const formData = {
                model: btoa(document.getElementById('txtModel').value),
                apiKey: document.getElementById('txtAPIKey').value,
                systemPrompt: btoa(document.getElementById('txtSystemPrompt').value),
                request: btoa(document.getElementById('txtPrompt').value),
				response: btoa("22222"),
				requestNumber: "_direct_" + document.getElementById('txtRequestNumber').value
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
                    resultDiv.innerHTML = '<br><strong>The Response</strong> will be saved and can be retrieved through the Burp extension Oll0maView.<br />The file will contain _direct_ with the Request Number specified above.<br><br>';
                }
            } catch (error) {
                resultDiv.innerHTML = '<strong>Error:</strong> ' + error.message;
            }
        });
    </script>
</body>
</html>
`

var config Configuration

func main() {
	//config.OllamaURL = "http://10.27.20.160:11434/api/chat"
	//config.OllamaModel = "llama3.2:3b"

	ConfigPtr := flag.String("config", "config.json", "Configuration file to load for the proxy")
	flag.Parse()

	configFile := *ConfigPtr
	log.Println("Loading the following config file: " + configFile + "\n")
	if err := config.LoadConfig(configFile); err != nil {
		//fmt.Println("Could not load the configuration file, creating a new default config.json")
		config.CreateConfig(configFile)
		log.Fatalf("Modify the %s file to customize how the tool functions: %v\n", configFile, err)
	}

	// Verify the TLS Certificate and Key files exist for the https server
	// Create the location of the keys folder
	dirPathTLS := filepath.Dir(config.TLSConfig)
	err := os.MkdirAll(dirPathTLS, os.ModePerm)
	if err != nil {
		log.Fatalf("Failed to create directories for TLS: %v", err)
	}
	// Does the tlsConfig.json  file exist in the keys folder
	TLSConfigFileExists := FileExists("/" + config.TLSConfig)
	//fmt.Println(configFileExists)
	if !TLSConfigFileExists {
		CreateCertConfigFile(config.TLSConfig)
		log.Fatalf("Created %s, modify the values to create the self-signed cert utilized", config.TLSConfig)
	}

	// Does the server.crt and server.key files exist in the keys folder
	crtFileExists := FileExists("/" + config.TLSCert)
	keyFileExists := FileExists("/" + config.TLSKey)
	if !crtFileExists || !keyFileExists {
		CreateCerts(config.TLSConfig, config.TLSCert, config.TLSKey)
		crtFileExists := FileExists("/" + config.TLSCert)
		keyFileExists := FileExists("/" + config.TLSKey)
		if !crtFileExists || !keyFileExists {
			fmt.Printf("Failed to create %s and %s files\n", config.TLSCert, config.TLSKey)
			os.Exit(0)
		}
	}

	if len(config.APIKey) > 0 {
		fmt.Printf("Using Existing API Key from the config loaded: %s\n\n", config.APIKey)
	} else {
		fmt.Printf("No API Key declared in the config file, generating a one-time use Key\n")
		// Create Random API Key to be Used in the Burp Extension
		config.APIKey = GenerateRandomString(64)
		fmt.Printf("\nGenerated API Key: %s\n\n", config.APIKey)
	}

	// Create queue server with 100 job capacity
	serverIPPort := config.QueueServerURL
	serverIPPort = strings.TrimPrefix(serverIPPort, "https://")
	serverIPPort = strings.TrimPrefix(serverIPPort, "http://")
	server, err := NewQueueServer(serverIPPort, 100, "0") // The request number is 0 until populated
	if err != nil {
		log.Fatalf("Failed to create server: %v", err)
	}

	// Handle graceful shutdown
	go func() {
		sigChan := make(chan os.Signal, 1)
		<-sigChan

		if err := server.Stop(); err != nil {
			log.Printf("Error during shutdown: %v", err)
		}
		os.Exit(0)
	}()

	// Start the server
	if err := server.Start(); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
