package main

import (
	"context"
	"encoding/json"
	"flag"
	"log"
	"os"

	"google.golang.org/adk/agent"
	"google.golang.org/adk/agent/llmagent"
	"google.golang.org/adk/agent/workflowagents/sequentialagent"
	"google.golang.org/adk/cmd/launcher/adk"
	"google.golang.org/adk/cmd/launcher/full"
	"google.golang.org/adk/model/gemini"
	"google.golang.org/adk/server/restapi/services"
	"google.golang.org/genai"
)

/**
References:
https://google.github.io/adk-docs/agents/multi-agents/#workflow-agents-as-orchestrators
https://www.kaggle.com/kaggle5daysofai/code

https://google.github.io/adk-docs/agents/workflow-agents/sequential-agents/


**/

type Configuration struct {
	APIKey string `json:"apikey"`
}

func (c *Configuration) CreateConfig(f string) error {
	c.APIKey = "NOTHING"
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

	model, err := gemini.NewModel(ctx, "gemini-2.5-flash", &genai.ClientConfig{
		APIKey: config.APIKey,
	})

	if err != nil {
		log.Fatalf("Failed to create model: %v", err)
	}

	// Outline Agent
	outlineAgent, err := llmagent.New(llmagent.Config{
		Name:        "OutlineAgent",
		Model:       model,
		Instruction: "Create a blog outline for the given topic with:\n1. A catchy headline\n2. An introduction hook\n3. 3-5 main sections with 2-3 bullet points for each\n4. A concluding thought",
		OutputKey:   "outlineOutput",
	})
	if err != nil {
		log.Fatalf("Failed to create the outline agent: %v", err)
	}

	// Writer Agent
	writerAgent, err := llmagent.New(llmagent.Config{
		Name:        "WriterAgent",
		Model:       model,
		Instruction: "Using the provided outline from {outlineOutput}, write a detailed blog post expanding on each section and bullet point. Ensure the content is engaging, informative, and flows well from one section to the next.",
		OutputKey:   "outlineDraft",
	})
	if err != nil {
		log.Fatalf("Failed to create the writer agent: %v", err)
	}

	// Editor Agent
	editorAgent, err := llmagent.New(llmagent.Config{
		Name:        "EditorAgent",
		Model:       model,
		Instruction: "Edit this draft: {outlineDraft}.  Your task is to poslish the text by fixing any grammatical errors, improving the flow and sentence structure, and enhancing overall clarity.",
		OutputKey:   "outlineDraft",
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

	configADK := &adk.Config{
		AgentLoader: services.NewSingleAgentLoader(rootAgent),
	}

	l := full.NewLauncher()
	// Run the launcher with the provided command-line arguments.
	//err = l.Execute(ctx, configADK, os.Args[1:])
	err = l.Execute(ctx, configADK, []string{"console"})
	//err = l.Execute(ctx, configADK, []string{"web", "api", "webui"}) // This does not quite work with the webui - need to research further
	if err != nil {
		log.Fatalf("run failed: %v\n\n%s", err, l.CommandLineSyntax())
	}
}
