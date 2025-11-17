package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"

	"google.golang.org/adk/agent"
	"google.golang.org/adk/agent/llmagent"
	"google.golang.org/adk/agent/workflowagents/sequentialagent"
	"google.golang.org/adk/model"

	"google.golang.org/adk/runner"
	"google.golang.org/adk/session"
	"google.golang.org/genai"
)

/**
References:
https://google.github.io/adk-docs/agents/multi-agents/#workflow-agents-as-orchestrators
https://www.kaggle.com/kaggle5daysofai/code

https://google.github.io/adk-docs/agents/workflow-agents/sequential-agents/


**/

type Configuration struct {
	Topic          string `json:"topic"`
	GeminiAPIKey   string `json:"geminiAPIKey"`
	OllamaURL      string `json:"ollamaURL"`
	OllamaWaitTime int    `json:"ollamaWaitTime"`
}

func (c *Configuration) CreateConfig(f string) error {
	c.Topic = "How to Research Threat Actors in 2025"
	c.GeminiAPIKey = ""
	c.OllamaURL = "http://localhost:11434/api/chat"
	c.OllamaWaitTime = 10 // HTTP Waittime for a response from ollama in minutes
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

func ollamaNewModel(ctx context.Context, modelName string, c Configuration) (model.LLM, error) {
	ctx = context.Background()
	return &ollamaModel{
		name:       modelName,
		ollamaURL:  c.OllamaURL,
		waitTime:   c.OllamaWaitTime,
		SequenceID: 0,
	}, nil

}

func main() {
	ConfigPtr := flag.String("config", "config.json", "Path to configuration file")
	flag.Parse()

	// Load the Configuration file
	var config Configuration
	configFile := *ConfigPtr
	log.Println("Loading the following config file: " + configFile + "\n")
	if err := config.LoadConfig(configFile); err != nil {
		//fmt.Println("Could not load the configuration file, creating a new default config.json")
		config.CreateConfig(configFile)
		log.Fatalf("Modify the %s file to customize how the tool functions: %v\n", configFile, err)
	}

	ctx := context.Background()

	/**
	model, err := gemini.NewModel(ctx, "gemini-2.5-flash", &genai.ClientConfig{
		APIKey: config.GeminiAPIKey,
	})
	if err != nil {
		log.Fatalf("Failed to create the Gemini model: %v", err)
	}
	**/

	llamaModel, err := ollamaNewModel(ctx, "llama3.2:3b", config)
	if err != nil {
		log.Printf("Ollama model creation failed: %v", err)
		os.Exit(0)
	}

	deepseekModel, err := ollamaNewModel(ctx, "deepseek-r1:1.5b", config)
	if err != nil {
		log.Printf("Ollama model creation failed: %v", err)
		os.Exit(0)
	}

	/** Hit 10 minute timeout for ollama responses
	qwenModel, err := ollamaNewModel(ctx, "qwen3:8b", config)
	if err != nil {
		log.Printf("Ollama model creation failed: %v", err)
		os.Exit(0)
	}
	**/

	/** Hit 10 minute timeout for ollama responses
	gptossModel, err := ollamaNewModel(ctx, "gpt-oss:20b", config)
	if err != nil {
		log.Printf("Ollama model creation failed: %v", err)
		os.Exit(0)
	}
	**/

	// Outline Agent
	outlineAgent, err := llmagent.New(llmagent.Config{
		Name:        "OutlineAgent",
		Model:       deepseekModel,
		Instruction: "Create a blog outline for the given topic with:\n1. A catchy headline\n2. An introduction hook\n3. 3-5 main sections with 2-3 bullet points for each\n4. A concluding thought",
		OutputKey:   "outlineOutput",
	})
	if err != nil {
		log.Fatalf("Failed to create the outline agent: %v", err)
	}

	// Writer Agent
	writerAgent, err := llmagent.New(llmagent.Config{
		Name:        "WriterAgent",
		Model:       deepseekModel,
		Instruction: "Using the provided outline from {outlineOutput}, write a detailed blog post expanding on each section and bullet point. Ensure the content is engaging, informative, and flows well from one section to the next.",
		OutputKey:   "outlineDraft",
	})
	if err != nil {
		log.Fatalf("Failed to create the writer agent: %v", err)
	}

	// Editor Agent
	editorAgent, err := llmagent.New(llmagent.Config{
		Name:        "EditorAgent",
		Model:       llamaModel,
		Instruction: "Edit this draft: {outlineDraft}.  Your task is to poslish the text by fixing any grammatical errors, improving the flow and sentence structure, and enhancing overall clarity.",
		OutputKey:   "editorOutput",
	})
	if err != nil {
		log.Fatalf("Failed to create the writer agent: %v", err)
	}

	// Sequential Agent Example
	rootAgent, err := sequentialagent.New(sequentialagent.Config{
		AgentConfig: agent.Config{
			Name:        "BlogPostGenerator",
			Description: "Executes a sequence of agents to generate a blog post based on a given topic. ",
			SubAgents:   []agent.Agent{outlineAgent, writerAgent, editorAgent},
		},
	})
	if err != nil {
		log.Fatalf("Failed to create the writer agent: %v", err)
	}

	userTopic := config.Topic

	// Modified to use session and a runner instead of using the command line launcher
	sessionService := session.InMemoryService()
	initialState := map[string]any{
		"topic": userTopic,
	}

	sessionInstance, err := sessionService.Create(ctx, &session.CreateRequest{
		AppName: rootAgent.Name(),
		UserID:  "thepcn3rd",
		State:   initialState,
	})
	if err != nil {
		log.Fatalf("Failed to create session: %v", err)
	}

	r, err := runner.New(runner.Config{
		AppName:        rootAgent.Name(),
		Agent:          rootAgent,
		SessionService: sessionService,
	})
	if err != nil {
		log.Fatalf("Failed to create runner: %v", err)
	}

	input := genai.NewContentFromText("Generate a blogpost about: "+userTopic, genai.RoleUser)
	events := r.Run(ctx, "thepcn3rd", sessionInstance.Session.ID(), input, agent.RunConfig{
		StreamingMode: agent.StreamingModeSSE,
	})

	var finalResponse string
	previousAgentAuthor := ""
	eventSectionResponse := ""
	for event, err := range events {
		if err != nil {
			log.Fatalf("An error occurred during agent execution: %v", err)
		}

		// Print each event as it arrives.
		if event.Author != previousAgentAuthor {
			//fmt.Printf("\nEvent Section Response:\n%s\n", eventSectionResponse)
			finalResponse += "\n\n--\n\n"
			eventSectionResponse = ""
			fmt.Println("\n----- Agent Response -----")
			fmt.Printf("Agent Name: %s\n", event.Author)
			//fmt.Printf("Event Author: %s\n", event.Author)
			//fmt.Printf("Event ID: %s\n", event.ID)
			//fmt.Printf("Event Branch: %s\n", event.Branch)
			//fmt.Printf("Event Invocation ID: %s\n", event.InvocationID)
			fmt.Printf("Event Timestamp: %s\n", event.Timestamp)
			//fmt.Printf("Event Content Role: %s\n", event.Content.Role)
			previousAgentAuthor = event.Author
		}

		for _, part := range event.Content.Parts {
			finalResponse += part.Text
			eventSectionResponse += part.Text
		}

	}

	//fmt.Println("\n--- Agent Interaction Result ---")
	fmt.Printf("\nAgent Final Response:\n%s\n\n", finalResponse)

	//finalSession, err := sessionService.Get(ctx, &session.GetRequest{
	_, err = sessionService.Get(ctx, &session.GetRequest{
		UserID:    "testing",
		AppName:   rootAgent.Name(),
		SessionID: sessionInstance.Session.ID(),
	})
	if err != nil {
		log.Fatalf("Failed to retrieve final session: %v", err)
	}

	//fmt.Println("Final Session State:", finalSession.Session.State())

}
