package main

import (
	"context"
	"fmt"
	"time"

	"google.golang.org/adk/agent"
	"google.golang.org/adk/agent/llmagent"
	"google.golang.org/adk/agent/workflowagents/parallelagent"
	"google.golang.org/adk/agent/workflowagents/sequentialagent"
	"google.golang.org/adk/model/gemini"
	"google.golang.org/adk/runner"
	"google.golang.org/adk/session"
	"google.golang.org/genai"
)

func CreateEmail(c Configuration, i InformationStruct) (InformationStruct, error) {
	ctx := context.Background()
	/**
		geminiFlash_2_5, err := gemini.NewModel(ctx, "gemini-2.5-flash", &genai.ClientConfig{
			APIKey: config.APIKey,
		})
		if err != nil {
			log.Fatalf("Failed to create model: %v", err)
		}
	**/

	geminiFlash_2_5_Lite, err := gemini.NewModel(ctx, "gemini-2.5-flash-lite", &genai.ClientConfig{
		APIKey: c.APIKey,
	})

	if err != nil {
		//log.Fatalf("Failed to create model: %v", err)
		return InformationStruct{}, fmt.Errorf("failed to create model: %v", err)
	}

	promptEmailCreation := fmt.Sprintf(`You are an analyst that is evaluating the Product URLs provided to create an advertisement for %s.
	
	**1st Product**
	Product URL: %s
	Image URL: %s
	
	**2nd Product**
	Product URL: %s
	Image URL: %s

	**3rd Product**
	Product URL: %s
	Image URL: %s
	
	**4th Product**
	Product URL: %s
	Image URL: %s

	Category URL: %s

	Only visit the URL provided and then, generate an HTML-formatted email advertisement based on this item with a red color appeal. The email should include:

	An engaging subject line.
	A main heading (e.g., <h1> ) for the items.
	
	Then in 2 columns and 2 rows list the items with the following:
	A brief introductory paragraph highlighting the appeal of the item, styled with <p> tags.
	A sub-heading (e.g., <h2> or <h3> ) for the item's name.
	A paragraph with its brief description.
	Provide a picture of the product with a clickable link to the image URL provided.
	A clickable link ( <a> tag) using the item's direct product URL, with descriptive link text like 'View Item on Store'.
	A strong call to action at the end, styled with a button-like <a> tag, encouraging recipients to 'Shop All Home Decor' on <Store>, linked to the main home decor Category URL.
	Ensure the HTML is well-structured and easy to read. The tone should be inviting and enthusiastic.

	Only use valid links derived from the URL provided.
	`, i.DemographicInfo, i.URL1, i.URLImage1, i.URL2, i.URLImage2, i.URL3, i.URLImage3, i.URL4, i.URLImage4, i.CategoryURL)

	// Creation Agent
	creationAgent, err := llmagent.New(llmagent.Config{
		Name:        "CreationAgent",
		Model:       geminiFlash_2_5_Lite,
		Instruction: promptEmailCreation,
		OutputKey:   "creationOutput",
	})
	if err != nil {
		//log.Fatalf("Failed to create the creation agent: %v", err)
		return InformationStruct{}, fmt.Errorf("failed to create the creation agent: %v", err)
	}

	promptVerificationAgent := fmt.Sprintf(`You are a college student between the ages of 18-25, evaluate the HTML formatted email.  The criteria to evaluate the email is below:

	Criteria to Evaluate Email:
	1. Check the spelling of the descriptions provided
	2. Verify that no strange characters exist in the descriptions
	3. Analyze the links and verify they resolve

	Provide feedback for what could be improved with the email.

	**HTML Formatted Email**
	{creationOutput}
	`)

	// Verification Agent
	verificationAgent, err := llmagent.New(llmagent.Config{
		Name:        "VerificationAgent",
		Model:       geminiFlash_2_5_Lite,
		Instruction: promptVerificationAgent,
		OutputKey:   "verificationOutput",
	})
	if err != nil {
		//log.Fatalf("Failed to create the learning agent: %v", err)
		return InformationStruct{}, fmt.Errorf("failed to create the learning agent: %v", err)
	}

	// Run the translation agents in parallel
	promptFrenchAgent := fmt.Sprintf(`You are a translator that will take the content and translate the HTML formatted email to French.
	
	**HTML Formatted Email**
	{creationOutput}`)

	// French Agent
	frenchAgent, err := llmagent.New(llmagent.Config{
		Name:        "FrenchAgent",
		Model:       geminiFlash_2_5_Lite,
		Instruction: promptFrenchAgent,
		OutputKey:   "frenchOutput",
	})
	if err != nil {
		//log.Fatalf("Failed to create the learning agent: %v", err)
		return InformationStruct{}, fmt.Errorf("failed to create the learning agent: %v", err)
	}

	// Run the translation agents in parallel
	promptGermanAgent := fmt.Sprintf(`You are a translator that will take the content and translate the HTML formatted email to French.
	
	**HTML Formatted Email**
	{creationOutput}`)

	// German Agent
	germanAgent, err := llmagent.New(llmagent.Config{
		Name:        "GermanAgent",
		Model:       geminiFlash_2_5_Lite,
		Instruction: promptGermanAgent,
		OutputKey:   "germanOutput",
	})
	if err != nil {
		//log.Fatalf("Failed to create the learning agent: %v", err)
		return InformationStruct{}, fmt.Errorf("failed to create the learning agent: %v", err)
	}

	// Parallel Agents to Translate the Information
	parallelTranslationAgent, err := parallelagent.New(parallelagent.Config{
		AgentConfig: agent.Config{
			Name:        "ParallelTranslationAgent",
			Description: "Runs multiple agents that conducts the translation of the information",
			SubAgents:   []agent.Agent{frenchAgent, germanAgent},
		},
	})
	if err != nil {
		//log.Fatalf("Failed to create the learning agent: %v", err)
		return InformationStruct{}, fmt.Errorf("failed to create the parallel agent: %v", err)
	}

	// Sequential Agent Example
	rootAgent, err := sequentialagent.New(sequentialagent.Config{
		AgentConfig: agent.Config{
			Name:        "CreateEmail",
			Description: "Executes a sequence of agents to create an HTML email.",
			SubAgents:   []agent.Agent{creationAgent, verificationAgent, parallelTranslationAgent},
		},
	})
	if err != nil {
		//log.Fatalf("Failed to create the sequential agent: %v", err)
		return InformationStruct{}, fmt.Errorf("failed to create the sequential agent: %v", err)
	}

	sessionService := session.InMemoryService()
	initialState := map[string]any{
		"topic": "Email Builder",
	}
	sessionInstance, err := sessionService.Create(ctx, &session.CreateRequest{
		AppName: rootAgent.Name(),
		UserID:  "generation",
		State:   initialState,
	})
	if err != nil {
		//log.Fatalf("Failed to create session: %v", err)
		return InformationStruct{}, fmt.Errorf("failed to create a session: %v", err)
	}

	//
	//userTopic := "Browse the site and identify home decor items that would be appealing to 18-25 year olds."

	r, err := runner.New(runner.Config{
		AppName:        rootAgent.Name(),
		Agent:          rootAgent,
		SessionService: sessionService,
	})
	if err != nil {
		//log.Fatalf("Failed to create runner: %v", err)
		return InformationStruct{}, fmt.Errorf("failed to create the runner: %v", err)
	}

	input := genai.NewContentFromText("Generate an Email", genai.RoleUser)
	events := r.Run(ctx, "generation", sessionInstance.Session.ID(), input, agent.RunConfig{
		StreamingMode: agent.StreamingModeSSE,
	})

	var finalResponse string
	previousAgentAuthor := ""
	eventSectionResponse := ""
	var lastEventTimestamp time.Time
	for event, err := range events {
		if err != nil {
			//log.Fatalf("An error occurred during agent execution: %v", err)
			return InformationStruct{}, fmt.Errorf("an error occurred during agent execution: %v", err)
		}

		//previousAgentAuthor = event.Author

		// Print each event as it arrives.
		if event.Author != previousAgentAuthor {
			//finalResponse += "\n\n--\n\n"

			switch previousAgentAuthor {
			case "CreationAgent":
				i.HTMLEmail = eventSectionResponse
			case "VerificationAgent":
				i.VerificationInformation = eventSectionResponse
			case "FrenchAgent":
				i.FrenchInformation = eventSectionResponse
			case "GermanAgent":
				i.GermanInformation = eventSectionResponse
			}
			//fmt.Printf("[A2] %s - Response from LLM: %s\n", previousAgentAuthor, eventSectionResponse)
			eventSectionResponse = ""
			if previousAgentAuthor != "" {
				//fmt.Println("\n----- Agent Response Completed -----")
				fmt.Printf("[A] %s - Response Completed ", previousAgentAuthor)
				//fmt.Printf("Event Author: %s\n", event.Author)
				//fmt.Printf("Event ID: %s\n", event.ID)
				//fmt.Printf("Event Branch: %s\n", event.Branch)
				//fmt.Printf("Event Invocation ID: %s\n", event.InvocationID)
				fmt.Printf("- Timestamp: %s\n", event.Timestamp)
				lastEventTimestamp = event.Timestamp
			}
			//fmt.Printf("Event Content Role: %s\n", event.Content.Role)

		}
		previousAgentAuthor = event.Author

		countParts := 0
		// Sometimes this crashes with an integer overflow, not sure at the moment what causes that...
		if len(event.Content.Parts) > 0 {
			for i, part := range event.Content.Parts {
				finalResponse += part.Text
				//eventSectionResponse += part.Text
				//fmt.Println(part.Text)
				countParts = i
			}
			// At the moment this works to capture the final parts...
			eventSectionResponse = event.Content.Parts[countParts].Text
		}

	}

	fmt.Printf("[A] %s - Response Completed ", previousAgentAuthor)
	fmt.Printf("- Timestamp: %s\n", lastEventTimestamp)

	switch previousAgentAuthor {
	case "CreationAgent":
		i.HTMLEmail = eventSectionResponse // This section is not necessary of the switch at this location
	case "VerificationAgent":
		i.VerificationInformation = eventSectionResponse // Capture the Verification Information as it Exits the For Loop
	case "FrenchAgent":
		i.FrenchInformation = eventSectionResponse
	case "GermanAgent":
		i.GermanInformation = eventSectionResponse
	}
	//fmt.Println("\n--- Agent Interaction Result ---")
	//fmt.Println("Agent Final Response: " + finalResponse)

	_, err = sessionService.Get(ctx, &session.GetRequest{
		UserID:    "generation",
		AppName:   rootAgent.Name(),
		SessionID: sessionInstance.Session.ID(),
	})

	if err != nil {
		//log.Fatalf("Failed to retrieve final session: %v", err)
		return InformationStruct{}, fmt.Errorf("failed to retrieve final session information: %v", err)
	}

	//fmt.Println("Final Session State:", finalSession.Session.State())
	//fmt.Printf("[F] Final Verdict of All Agents: %s\n", strings.ReplaceAll(finalResponse, "\n", ""))
	i.FinalResponse = finalResponse

	return i, nil
}
