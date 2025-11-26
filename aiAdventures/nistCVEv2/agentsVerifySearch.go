package main

import (
	"context"
	"fmt"
	"log"
	"nistCVEv2/ollama"
	"os"
	"strings"

	"google.golang.org/adk/agent"
	"google.golang.org/adk/agent/llmagent"
	"google.golang.org/adk/agent/workflowagents/sequentialagent"
	"google.golang.org/adk/model"
	"google.golang.org/adk/runner"
	"google.golang.org/adk/session"
	"google.golang.org/genai"
)

type AgenticAgentConfiguration struct {
	VerificationModel       string
	VerificationInstruction string
	VerificationDescription string
	CriticModel             string
	CriticInstruction       string
	CriticDescription       string
	RefinerModel            string
	RefinerInstruction      string
	RefinerDescription      string
	NVD                     NVDOutputStruct
	Iteration               int
}

func AgenticAgentCall(aaConfig AgenticAgentConfiguration) (string, error) {
	ctx := context.Background()
	// Send the results to Ollama for Verification that the results relate to the CVE
	verificationModel, err := ollama.OllamaNewModel(ctx, aaConfig.VerificationModel, config.OllamaURL, config.OllamaWaitTime)
	if err != nil {
		log.Printf("Ollama model creation failed: %v", err)
		os.Exit(0)
	}

	criticModel, err := ollama.OllamaNewModel(ctx, aaConfig.CriticModel, config.OllamaURL, config.OllamaWaitTime)
	if err != nil {
		log.Printf("Ollama model creation failed: %v", err)
		os.Exit(0)
	}

	refinerModel, err := ollama.OllamaNewModel(ctx, aaConfig.RefinerModel, config.OllamaURL, config.OllamaWaitTime)
	if err != nil {
		log.Printf("Ollama model creation failed: %v", err)
		os.Exit(0)
	}

	// Verification Agent
	//var outputResult string
	verificationAgent, err := llmagent.New(llmagent.Config{
		Name:        "VerificationAgent",
		Model:       verificationModel,
		Description: aaConfig.VerificationDescription,
		Instruction: aaConfig.VerificationInstruction,
		OutputKey:   "outputResult",
	})
	if err != nil {
		log.Fatalf("Failed to create the outline agent: %v", err)
	}

	//aaConfig.CriticInstruction = strings.Replace(aaConfig.VerificationInstruction, "VERIFICATION_PLACEHOLDER", "outputResult", -1)

	// Critic Agent
	//var outputCritic string
	criticAgent, err := llmagent.New(llmagent.Config{
		Name:        "CriticAgent",
		Model:       criticModel,
		Description: aaConfig.CriticDescription,
		Instruction: aaConfig.CriticInstruction,
		OutputKey:   "outputCritic",
	})
	if err != nil {
		log.Fatalf("Failed to create the outline agent: %v", err)
	}

	/**
	exitLoopTool, err := functiontool.New(
		functiontool.Config{
			Name:        "exitLoop",
			Description: "Call this function ONLY when the critique indicates no further changes are needed.",
		},
		ExitLoop,
	)
	if err != nil {
		return "", fmt.Errorf("failed to create exit loop tool: %v", err)
	}
	**/

	//aaConfig.RefinerInstruction = strings.Replace(aaConfig.RefinerInstruction, "VERIFICATION_PLACEHOLDER", "outputResult", -1)
	//aaConfig.RefinerInstruction = strings.Replace(aaConfig.RefinerInstruction, "REFINER_PLACEHOLDER", "outputCritic", -1)

	// Refiner Agent
	//var outputRefiner string
	refinerAgent, err := llmagent.New(llmagent.Config{
		Name:        "RefinerAgent",
		Model:       refinerModel,
		Description: aaConfig.RefinerDescription,
		Instruction: aaConfig.RefinerInstruction,
		OutputKey:   "outputRefiner",
	})
	if err != nil {
		log.Fatalf("Failed to create the outline agent: %v", err)
	}

	/**
	refinementLoop, err := loopagent.New(loopagent.Config{
		AgentConfig: agent.Config{
			Name:      "RefinementLoop",
			SubAgents: []agent.Agent{criticAgent, refinerAgent},
		},
		MaxIterations: 2,
	})
	**/

	// Sequential Agent
	rootAgent, err := sequentialagent.New(sequentialagent.Config{
		AgentConfig: agent.Config{
			Name:        "RootAgent",
			Description: "You are a CVE analyst.  You evaluate search results and verify they match a given CVE.",
			SubAgents:   []agent.Agent{verificationAgent, criticAgent, refinerAgent},
		},
	})
	if err != nil {
		log.Fatalf("Failed to create the writer agent: %v", err)
	}

	userTopic := aaConfig.NVD.CVEID

	// Modified to use session and a runner instead of using the command line launcher
	sessionService := session.InMemoryService()
	initialState := map[string]any{
		"topic": userTopic,
	}

	sessionInstance, err := sessionService.Create(ctx, &session.CreateRequest{
		AppName: "VerificationAgent",
		UserID:  "thepcn3rd",
		State:   initialState,
	})
	if err != nil {
		log.Fatalf("Failed to create session: %v", err)
	}

	r, err := runner.New(runner.Config{
		AppName:        "VerificationAgent",
		Agent:          rootAgent,
		SessionService: sessionService,
	})
	if err != nil {
		log.Fatalf("Failed to create runner: %v", err)
	}

	input := genai.NewContentFromText("Summarize the search results for "+userTopic, genai.RoleUser)
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

		if eventSectionResponse != "" {
			fmt.Printf("[A] %s - Response from LLM: %s\n", previousAgentAuthor, eventSectionResponse)
		}

		// Print each event as it arrives.
		if event.Author != previousAgentAuthor {
			//finalResponse += "\n\n--\n\n"
			if previousAgentAuthor != "" {
				eventSectionResponse = ""
				//fmt.Println("\n----- Agent Response Completed -----")
				fmt.Printf("[A] %s - Response Completed ", previousAgentAuthor)
				//fmt.Printf("Event Author: %s\n", event.Author)
				//fmt.Printf("Event ID: %s\n", event.ID)
				//fmt.Printf("Event Branch: %s\n", event.Branch)
				//fmt.Printf("Event Invocation ID: %s\n", event.InvocationID)
				fmt.Printf("Timestamp: %s\n", event.Timestamp)
			}

			//fmt.Printf("Event Content Role: %s\n", event.Content.Role)
			previousAgentAuthor = event.Author
		}

		//for _, part := range event.Content.Parts {
		for _, part := range event.LLMResponse.Content.Parts {
			finalResponse += part.Text + " "
			eventSectionResponse += part.Text + " "
		}

	}
	fmt.Printf("[A] %s - Response from LLM: %s\n", previousAgentAuthor, eventSectionResponse)
	//fmt.Println("\n--- Agent Interaction Result ---")
	fmt.Printf("[F] Final Verdict of All Agents: %s\n", strings.ReplaceAll(finalResponse, "\n", ""))

	//finalSession, err := sessionService.Get(ctx, &session.GetRequest{
	_, err = sessionService.Get(ctx, &session.GetRequest{
		UserID:    "thepcn3rd",
		AppName:   "VerificationAgent",
		SessionID: sessionInstance.Session.ID(),
	})
	if err != nil {
		log.Fatalf("Failed to retrieve final session: %v", err)
	}

	//fmt.Println("Final Session State:", finalSession.Session.State())

	// Send the results to Ollama for Summarization
	// Verify the summary relates to the CVE and the description provided by NVD
	return finalResponse, nil
}

func OllamaVerifySearchResults(nvdOutputAll []NVDOutputStruct) ([]NVDOutputStruct, error) {
	for nvdCount, nvd := range nvdOutputAll {
		// If 2 or less Brave Search Results, skip the verification generation
		if len(nvd.BraveSearchResults) <= 2 {
			fmt.Printf("[V] Skipping Verification for CVE: %s - Only %d Brave Search Results Found\n\n", nvd.CVEID, len(nvd.BraveSearchResults))
			continue
		}
		for iteration := range nvd.BraveSearchResults {
			commonDescription := "Research Assistant - Are the search results similar to the CVE description."
			commonIntruction := fmt.Sprintf(`You are a research assistant validating CVE data search results.

I will provide:
- CVE Information
- Search Result

Your task:
1. Compare the Search Result to the CVE Information.
2. Determine if the CVE-xxxx-xxxxx number matches in both the Search Result and the CVE Information
3. Determine whether the Search Result is similar with the CVE Information.

Response Rules:
- If the CVE-xxxx-xxxxxx number is the same in the Search Result and the CVE Information, respond: YES!
- If the Search Result is similar or aligns with the CVE Information, respond: YES!
- If it contradicts, misstates, or conflicts with the CVE, respond: NO!

You must respond with **ONLY** “YES!” or “NO!”
No explanation, no reasoning, no qualifiers, no other words.

---

CVE Information:
%s - %s

Search Result:
%s - %s
`, nvd.CVEID, nvd.Description, nvd.BraveSearchResults[iteration].Title, nvd.BraveSearchResults[iteration].Description)

			// Used deepseek-r1:1.5b for the critic however its reasoning was not great for CVE analysis
			aaConfig := AgenticAgentConfiguration{
				VerificationModel:       "llama3.2:3b",
				VerificationDescription: commonDescription,
				VerificationInstruction: commonIntruction,
				CriticModel:             "qwen3:1.7b",
				CriticDescription:       commonDescription,
				CriticInstruction:       commonIntruction,
				RefinerModel:            "gemma3:1b",
				RefinerDescription:      commonDescription,
				RefinerInstruction:      commonIntruction,
				NVD:                     nvd,
				Iteration:               iteration,
			}

			//fmt.Printf("Verification Agent Prompt\n%s\n\n", aaConfig.VerificationInstruction)
			//fmt.Printf("Critic Agent Prompt\n%s\n\n", aaConfig.CriticInstruction)
			//fmt.Printf("Refiner Agent Prompt\n%s\n\n", aaConfig.RefinerInstruction)

			finalResponse, err := AgenticAgentCall(aaConfig)
			if err != nil {
				log.Printf("unable to determine if the topic matches the search results%v\n", err)
			}

			fmt.Printf("[I] Search Result #%d - CVE: %s\n", iteration, nvd.CVEID)
			fmt.Printf("[I] CVE Information: %s\n%s\n", nvd.CVEID, nvd.Description)
			fmt.Printf("[I] Search Result:\n%s\n", nvdOutputAll[nvdCount].BraveSearchResults[iteration].Description)
			//fmt.Println(finalResponse)

			if strings.Contains(finalResponse, "YES! YES! YES!") ||
				strings.Contains(finalResponse, "YES! NO! YES!") ||
				strings.Contains(finalResponse, "NO! YES! YES!") ||
				strings.Contains(finalResponse, "YES! YES! NO!") {
				fmt.Printf("[M] Search Result Matches CVE Topic\n\n")
				nvdOutputAll[nvdCount].BraveSearchResults[iteration].MatchesCVETopic = true
			} else {
				fmt.Println("-")
			}
		}
	}
	return nvdOutputAll, nil
}

type AgenticAgentConfig struct {
	NVD         NVDOutputStruct
	Iteration   int
	ModelConfig []ModelConfigStruct
}

type ModelConfigStruct struct {
	Name         string
	ModelName    string
	Instruction  string
	Description  string
	OutputKey    string
	ModelCreated model.LLM
}

func (m *ModelConfigStruct) CreateModel(ctx context.Context) error {
	var err error
	m.ModelCreated, err = ollama.OllamaNewModel(ctx, m.ModelName, config.OllamaURL, config.OllamaWaitTime)
	if err != nil {
		return fmt.Errorf("unable to create model: %s - %v", m.Name, err)
	}
	return nil
}

func OllamaCreateSummaryResults(nvdOutputAll []NVDOutputStruct) ([]NVDOutputStruct, error) {
	for nvdCount, nvd := range nvdOutputAll {
		// If only 2 or less Brave Search Results matched the CVE Topic, skip the summary generation
		if nvd.BraveSearchMatches <= 2 {
			fmt.Printf("[S] Skipping Summary Generation for CVE: %s - Only %d Brave Search Results Matched the CVE Topic\n\n", nvd.CVEID, nvd.BraveSearchMatches)
			continue
		}
		summaryInstruction := (`You are a research assistant writing a summary of search results related to a CVE.  Evaluate the following search results and create a summary that is 1 paragraph in length.  The summary should only be based on the search results provided below.
		
		I will provide:
		- Search Results related to a CVE.
		
		Your task:
		1. Review the search results.
		2. Write a concise summary that captures the key points for all of the search results.
		3. Ensure the summary is clear, easy to read and does not repeat information.

		You must response with **ONLY** the summary.
		No explanation, no reasoning, no qualifiers, no other words.

		---
		
		`)
		summaryInstructionResults := ""
		for iteration, result := range nvd.BraveSearchResults {
			if result.MatchesCVETopic == true {
				summaryInstructionResults += fmt.Sprintf("Search Result #%d:\n%s\n\n", iteration, result.Description)
			}
		}
		summaryInstruction += " " + summaryInstructionResults

		selectionInstuction := (`You are a research assistant selecting the best summary of search results related to a CVE.

		I will provide:
		- 2 summaries of search results related to a CVE.
		
		Your task:
		1. Review both summaries.
		2. Select the best summary that provides clarity and is easy to read.

		You must respond with **ONLY** the summary you select.
		No explanation, no reasoning, no qualifiers, no other words.

		---
		
		Summary A:
		{agentA_output}
		
		Summary B:
		{agentB_output}`)

		// Create the instruction for the writers
		aaConfig := AgenticAgentConfig{
			NVD: nvd,
			//Iteration:   iteration,
			ModelConfig: []ModelConfigStruct{},
		}
		Model_A := ModelConfigStruct{
			Name:        "Model_A",
			ModelName:   "llama3.2:3b",
			Instruction: summaryInstruction,
			Description: "Write a Summary of Search Results for a CVE.",
		}
		Model_B := ModelConfigStruct{
			Name:        "Model_B",
			ModelName:   "qwen3:1.7b",
			Instruction: summaryInstruction,
			Description: "Write a Summary of Search Results for a CVE.",
		}
		Model_C := ModelConfigStruct{
			Name:        "Model_C",
			ModelName:   "gemma3:1b",
			Instruction: selectionInstuction,
			Description: "Select the Best Summary of Search Results for a CVE.",
		}
		aaConfig.ModelConfig = append(aaConfig.ModelConfig, Model_A)
		aaConfig.ModelConfig = append(aaConfig.ModelConfig, Model_B)
		aaConfig.ModelConfig = append(aaConfig.ModelConfig, Model_C)

		fmt.Println("[A2] Starting Agentic AI Summary Generation...")

		finalResponseAgent, err := AgenticAgentCallv2(aaConfig)
		if err != nil {
			return nil, err
		}

		//fmt.Printf("\n\nDEBUG INFORMATION:\n%s\n\n", finalResponseAgent)

		fmt.Printf("[I2] CVE Count #%d - CVE: %s\n", nvdCount, nvd.CVEID)
		fmt.Printf("[I2] CVE Information: %s\n%s\n", nvd.CVEID, nvd.Description)
		//fmt.Printf("[I] Search Result:\n%s\n", nvdOutputAll[nvdCount].BraveSearchResults[iteration].Description)
		fmt.Printf("Final Response:\n%s\n\n", finalResponseAgent)

		if finalResponseAgent == "" {
			finalResponseAgent = "No Summary Generated"
		} else {
			nvdOutputAll[nvdCount].BraveSearchSummary = finalResponseAgent
		}

		fmt.Print(nvdOutputAll[nvdCount].BraveSearchSummary)

	}

	return nvdOutputAll, nil
}

func AgenticAgentCallv2(aaConfig AgenticAgentConfig) (string, error) {
	ctx := context.Background()
	// Create the Respective number of Models
	for modelCount, modelConfig := range aaConfig.ModelConfig {
		err := modelConfig.CreateModel(ctx)
		if err != nil {
			return "", fmt.Errorf("unable to create model: %s - %v", modelConfig.Name, err)
		}
		aaConfig.ModelConfig[modelCount] = modelConfig
	}

	// Create the Number of Agents
	Agent_A_Writer, err := llmagent.New(llmagent.Config{
		Name:        "Agent_A",
		Model:       aaConfig.ModelConfig[0].ModelCreated,
		Description: aaConfig.ModelConfig[0].Description,
		Instruction: aaConfig.ModelConfig[0].Instruction,
		OutputKey:   "agentA_output",
	})
	if err != nil {
		log.Fatalf("Failed to create the outline agent: %v", err)
	}

	Agent_B_Writer, err := llmagent.New(llmagent.Config{
		Name:        "Agent_B",
		Model:       aaConfig.ModelConfig[1].ModelCreated,
		Description: aaConfig.ModelConfig[1].Description,
		Instruction: aaConfig.ModelConfig[1].Instruction,
		OutputKey:   "agentB_output",
	})
	if err != nil {
		log.Fatalf("Failed to create the outline agent: %v", err)
	}

	Agent_C_Selector, err := llmagent.New(llmagent.Config{
		Name:        "Agent_C",
		Model:       aaConfig.ModelConfig[2].ModelCreated,
		Description: aaConfig.ModelConfig[2].Description,
		Instruction: aaConfig.ModelConfig[2].Instruction,
		OutputKey:   "agentC_output",
	})
	if err != nil {
		log.Fatalf("Failed to create the outline agent: %v", err)
	}

	// Sequential Agent
	rootAgent, err := sequentialagent.New(sequentialagent.Config{
		AgentConfig: agent.Config{
			Name:        "RootAgent",
			Description: "You are writing and selecting the best summary for a CVE.",
			SubAgents:   []agent.Agent{Agent_A_Writer, Agent_B_Writer, Agent_C_Selector},
		},
	})
	if err != nil {
		log.Fatalf("Failed to create the writer agent: %v", err)
	}

	userTopic := aaConfig.NVD.CVEID

	// Modified to use session and a runner instead of using the command line launcher
	sessionService := session.InMemoryService()
	initialState := map[string]any{
		"topic": userTopic,
	}

	sessionInstance, err := sessionService.Create(ctx, &session.CreateRequest{
		AppName: "SummaryAgent",
		UserID:  "thepcn3rd",
		State:   initialState,
	})
	if err != nil {
		log.Fatalf("Failed to create session: %v", err)
	}

	r, err := runner.New(runner.Config{
		AppName:        "SummaryAgent",
		Agent:          rootAgent,
		SessionService: sessionService,
	})
	if err != nil {
		log.Fatalf("Failed to create runner: %v", err)
	}

	input := genai.NewContentFromText("Verify information about this CVE is accurate "+userTopic, genai.RoleUser)
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

		if eventSectionResponse != "" {
			fmt.Printf("[A2] %s - Response from LLM: %s\n", previousAgentAuthor, eventSectionResponse)
		}

		// Print each event as it arrives.
		if event.Author != previousAgentAuthor {
			//finalResponse += "\n\n--\n\n"
			if previousAgentAuthor != "" {
				eventSectionResponse = ""
				//fmt.Println("\n----- Agent Response Completed -----")
				fmt.Printf("[A2] %s - Response Completed ", previousAgentAuthor)
				//fmt.Printf("Event Author: %s\n", event.Author)
				//fmt.Printf("Event ID: %s\n", event.ID)
				//fmt.Printf("Event Branch: %s\n", event.Branch)
				//fmt.Printf("Event Invocation ID: %s\n", event.InvocationID)
				fmt.Printf("Timestamp: %s\n", event.Timestamp)
			}

			//fmt.Printf("Event Content Role: %s\n", event.Content.Role)
			previousAgentAuthor = event.Author
		}

		//for _, part := range event.Content.Parts {
		for _, part := range event.LLMResponse.Content.Parts {
			finalResponse += part.Text + " "
			eventSectionResponse += part.Text + " "
		}

	}
	fmt.Printf("[A2] %s - Response from LLM: %s\n", previousAgentAuthor, eventSectionResponse)
	//fmt.Println("\n--- Agent Interaction Result ---")
	fmt.Printf("[F2] Final Verdict of All Agents: %s\n", strings.ReplaceAll(finalResponse, "\n", ""))
	finalResponse = strings.ReplaceAll(finalResponse, "Summary A", "")
	finalResponse = strings.ReplaceAll(finalResponse, "Summary B", "")

	//finalSession, err := sessionService.Get(ctx, &session.GetRequest{
	_, err = sessionService.Get(ctx, &session.GetRequest{
		UserID:    "thepcn3rd",
		AppName:   "SummaryAgent",
		SessionID: sessionInstance.Session.ID(),
	})
	if err != nil {
		log.Fatalf("Failed to retrieve final session: %v", err)
	}

	//fmt.Println("Final Session State:", finalSession.Session.State())

	// Send the results to Ollama for Summarization
	// Verify the summary relates to the CVE and the description provided by NVD

	//fmt.Printf("[F] Final Verdict of All Agents: %s\n", strings.ReplaceAll(finalResponse, "\n", ""))

	return finalResponse, nil
}
