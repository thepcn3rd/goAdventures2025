package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"time"

	cf "github.com/thepcn3rd/goAdvsCommonFunctions"
)

// Setup the environment variables for go
/*

go env -w GOROOT="/usr/lib/go"
go env -w GOPATH="/home/thepcn3rd/go/workspaces/nistCVE"

*/

// References: https://github.com/dasrick/go-teams-notify
// References: https://learn.microsoft.com/en-us/microsoftteams/platform/webhooks-and-connectors/how-to/add-incoming-webhook?tabs=dotnet
// References (How to format the text)
//		https://learn.microsoft.com/en-us/microsoftteams/platform/task-modules-and-cards/cards/cards-format?tabs=adaptive-md%2Cdesktop%2Cconnector-html

type adaptiveCardJSON struct {
	Type        string              `json:"type"`
	Attachments []attachmentsStruct `json:"attachments"`
}

type attachmentsStruct struct {
	ContentType string        `json:"contentType"`
	Content     contentStruct `json:"content"`
}

type contentStruct struct {
	Type    string       `json:"type"`
	Body    []bodyStruct `json:"body"`
	Schema  string       `json:"$schema"`
	Version string       `json:"version"`
}

type bodyStruct struct {
	Type string `json:"type"`
	Text string `json:"text"`
	Wrap bool   `json:"wrap"`
}

// Noticed based on the reference above that it is a POST to the teams channel
func SendTeamsMessage(message string, teamsWebhookURL string) {
	teamsClient := &http.Client{Timeout: 15 * time.Second}
	var teamsCard adaptiveCardJSON

	teamsCard.Type = "message"
	teamsCard.Attachments = make([]attachmentsStruct, 1)
	teamsCard.Attachments[0].ContentType = "application/vnd.microsoft.card.adaptive"
	teamsCard.Attachments[0].Content.Type = "AdaptiveCard"
	teamsCard.Attachments[0].Content.Body = make([]bodyStruct, 1)
	teamsCard.Attachments[0].Content.Body[0].Type = "TextBlock"
	teamsCard.Attachments[0].Content.Body[0].Text = message
	teamsCard.Attachments[0].Content.Body[0].Wrap = true
	teamsCard.Attachments[0].Content.Schema = "http://adaptivecards.io/schemas/adaptive-card.json"

	webhookMessageByte, _ := json.Marshal(teamsCard)
	webhookMessageBuffer := bytes.NewBuffer(webhookMessageByte)

	teamsRequest, err := http.NewRequest(http.MethodPost, teamsWebhookURL, webhookMessageBuffer)
	cf.CheckError("Unable to build new teams HTTP request", err, true)
	teamsRequest.Header.Add("Content-Type", "application/json;charset=utf-8")

	// Used for debugging
	//teamsResponse, err := teamsClient.Do(teamsRequest)
	//checkError("Unable to receive teams HTTP response", err)
	//if teamsResponse.StatusCode == 200 {
	//	fmt.Println("\nTeams message sent successfully!")
	//}

	_, err = teamsClient.Do(teamsRequest)
	cf.CheckError("Unable to receive teams HTTP response", err, true)

}
