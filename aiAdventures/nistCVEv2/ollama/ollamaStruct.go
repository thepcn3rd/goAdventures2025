package ollama

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"iter"
	"net/http"
	"time"

	"google.golang.org/adk/model"
	"google.golang.org/genai"
)

type OllamaConnector struct {
	OllamaURL string
	APIKey    string
	WaitTime  int
}

type OllamaRequestStruct struct {
	Stream   bool             `json:"stream"`
	Think    bool             `json:"think"`
	Messages []OllamaMessages `json:"messages"`
	Model    string           `json:"model"`
	//ModelOptions ModelOptionsStruct `json:"options,omitempty"`
}

type OllamaMessages struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ModelOptionsStruct struct {
	NumCTX        int     `json:"num_ctx"`        // Default: 2048 - Size of the context window used - The model has a max...
	Temperature   float64 `json:"temperature"`    // Default: 0.8 - 1.0 is more creative to 0.1 conservative text generation
	RepeatLastN   int     `json:"repeat_last_n"`  // How far back to look default 64
	RepeatPenalty float64 `json:"repeat_penalty"` // Repeat is more lenient Default: 1.1 0.9 may be better
	TopK          int     `json:"top_k"`          // Reduces the probability of generating non-sense.  Lower value is conservative Default: 40
	TopP          float64 `json:"top_p"`          // Default: 0.9 - 0.5 is more conservative text generation
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

func (o *OllamaRequestStruct) AddMessage(role string, content string) {
	message := OllamaMessages{
		Role:    role,
		Content: content,
	}
	o.Messages = append(o.Messages, message)
}

func (o *OllamaRequestStruct) ClearMessages() {
	o.Messages = []OllamaMessages{}
}

/**
func (o *OllamaRequestStruct) SetModelOptions(numCTX int, temperature float64, repeatLastN int, repeatPenalty float64, topK int, topP float64) {
	o.ModelOptions = ModelOptionsStruct{
		NumCTX:        numCTX,
		Temperature:   temperature,
		RepeatLastN:   repeatLastN,
		RepeatPenalty: repeatPenalty,
		TopK:          topK,
		TopP:          topP,
	}
}
**/

func (o *OllamaRequestStruct) SetModel(model string) {
	o.Model = model
}

func (o *OllamaRequestStruct) SetStream(stream bool) {
	o.Stream = stream
}

func (o *OllamaRequestStruct) SetThink(think bool) {
	o.Think = think
}

func (o *OllamaRequestStruct) SubmitRequest(connector OllamaConnector) (string, error) {
	var ollamaResponse OllamaResponseStruct
	jsonData, err := json.Marshal(o)
	if err != nil {
		return "", err
	}

	//log.Printf("Ollama URL: %s\n", connector.OllamaURL)
	//log.Printf("Ollama Request JSON: %s\n\n", jsonData)

	req, err := http.NewRequest("POST", connector.OllamaURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{
		Timeout: time.Duration(connector.WaitTime) * time.Minute,
	}

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("received non-OK HTTP status: %s", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("error reading response body: %w", err)
	}

	//log.Printf("%s\n\n", body)

	err = json.Unmarshal(body, &ollamaResponse)
	if err != nil {
		return "", fmt.Errorf("error unmarshaling ollama response: %w", err)
	}

	return ollamaResponse.Message.Content, nil
}

type OllamaModel struct {
	ModelName  string
	OllamaURL  string
	WaitTime   int
	SequenceID int
}

func OllamaNewModel(ctx context.Context, modelName string, ollamaURL string, ollamaWaitTime int) (model.LLM, error) {
	ctx = context.Background()

	return &OllamaModel{
		ModelName:  modelName,
		OllamaURL:  ollamaURL,
		WaitTime:   ollamaWaitTime,
		SequenceID: 0,
	}, nil

}

func (om *OllamaModel) IncrementSeqID() {
	om.SequenceID += 1
}

func (om *OllamaModel) Name() string {
	return om.ModelName
}

func (om *OllamaModel) GenerateContent(ctx context.Context, req *model.LLMRequest, stream bool) iter.Seq2[*model.LLMResponse, error] {
	if req.Config == nil {
		req.Config = &genai.GenerateContentConfig{}
	}
	if req.Config.HTTPOptions == nil {
		req.Config.HTTPOptions = &genai.HTTPOptions{}
	}

	return func(yield func(*model.LLMResponse, error) bool) {
		resp, err := om.generate(ctx, req)
		yield(resp, err)
	}

}

func (om *OllamaModel) generate(ctx context.Context, req *model.LLMRequest) (*model.LLMResponse, error) {
	ctx = context.Background()

	var connector OllamaConnector
	connector.OllamaURL = om.OllamaURL
	connector.WaitTime = om.WaitTime // in minutes

	var ollamaRequest OllamaRequestStruct
	ollamaRequest.SetModel(om.ModelName)
	//ollamaRequest.SetModelOptions(2048, 0.7, 64, 1.1, 40, 0.9)
	ollamaRequest.SetStream(false)
	ollamaRequest.SetThink(false) // Diabling the output of the think stage to clear up the output.
	ollamaRequest.ClearMessages()
	ollamaRequest.AddMessage("system", req.Contents[0].Parts[0].Text)
	ollamaRequest.AddMessage("user", req.Config.SystemInstruction.Parts[0].Text)

	/**
	// Debugging ----------------------------------------------

	for contentID, content := range req.Contents {
		//ollamaRequest.AddMessage(content.Role, content.Parts[0])
		log.Printf("ID: %d Added message - Role: %s\n", contentID, content.Role)
		log.Printf("ID: %d Content has %d parts\n", contentID, len(content.Parts))
		//log.Printf("First part content text: %v\n", content.Parts[0].Text)
		for partID, part := range content.Parts {
			//ollamaRequest.AddMessage(content.Role, part)
			if len(part.Text) > 50 {
				log.Printf("PartID: %d Added part to message: %v...\n", partID, part.Text[0:50])
			} else {
				log.Printf("PartID: %d Added part to message: %v\n", partID, part.Text)
			}
			//log.Printf("PartID: %d Text: %v\n", partID, part.Text[0:len(part.Text)])
			log.Printf("PartID: %d InlineData: %v\n", partID, part.InlineData)
		}
	}

	log.Println("\nRequest Config System Instruction Parts:")
	for _, part := range req.Config.SystemInstruction.Parts { // Here is the instruction that is sent in...
		log.Printf("Config System Part Text: %s\n", part.Text)
	}
	//log.Print(req.Config.SystemInstruction.Parts)
	log.Printf("--\n\n")

	// Debugging ----------------------------------------------
	**/

	responseText, err := ollamaRequest.SubmitRequest(connector)
	if err != nil {
		return nil, fmt.Errorf("submitrequest of the ollamarequest failed%v", err)
	}

	om.IncrementSeqID()

	return &model.LLMResponse{
		// https://pkg.go.dev/google.golang.org/genai#Content
		Content:        genai.NewContentFromText(responseText, "model"),
		CustomMetadata: map[string]any{},
		Partial:        false,
		TurnComplete:   true,
		Interrupted:    false,
	}, nil
}
