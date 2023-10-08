package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)
type Choices struct {
	Index int `json:"index"`
	Message struct {
				Role string `json:"role"`
				Content string `json:"content"`
			}
	FinishReason string `json:"finish_reason"`
}

type AiResponse struct {
	ID string `json:"id"`
	Object string `json:"object"`
	Created int `json:"created"`
	Model string `json:"model"`
	Choices []Choices `json:"choices"`
	Usage struct {
				CompletionTokens int `json:"completion_tokens"`
				PromptTokens int `json:"prompt_tokens"`
				TotalTokens int `json:"total_tokens"`
			} `json:"usage"`
}

type ChatBody struct {
	Model string `json:"model"`
	Messages []MessagesType `json:"messages"`
}

type MessagesType struct {
	Role string `json:"role"`
	Content string `json:"content"`
	Name string `json:"name"`
}

func GenerateHashtags(w http.ResponseWriter, quote RandomQuote, hashtags string) (chatBody AiResponse) {
		// 2 - Use chatGPT for something
		instructions := PromptBuilder(fmt.Sprintf("\"%s\" - %s \n%s", quote.Content, quote.Author, hashtags))
		jsonInstruction, err := json.Marshal(instructions)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to marshal the instruction: %s", err.Error()), http.StatusInternalServerError)
			return
		}
		
		chatReq, err := http.NewRequest("POST", "https://api.naga.ac/v1/chat/completions", bytes.NewBuffer(jsonInstruction))
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to make a request: %s", err.Error()), http.StatusInternalServerError)
			return
		}
		chatReq.Header.Add("Content-type", "application/json")
		apiKey := os.Getenv("NAGA_AI_KEY")
		chatReq.Header.Add("Authorization", fmt.Sprintf("Bearer %s", apiKey))
		
		client := &http.Client{}
		chatResp, err := client.Do(chatReq)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to get a response from AI: %s", err.Error()), http.StatusInternalServerError)
			return
		}
		defer chatResp.Body.Close()
		
		if chatResp.StatusCode != http.StatusOK {
			http.Error(w, fmt.Sprintf("Request failed with status code: %v", chatResp.StatusCode), http.StatusInternalServerError)
			return
		}
		
		chatBodyBytes, err := io.ReadAll(chatResp.Body)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to read response body: %s", err.Error()), http.StatusInternalServerError)
			return
		}
		
		if err := json.Unmarshal(chatBodyBytes, &chatBody); err != nil {
			http.Error(w, fmt.Sprintf("Failed to Unmarshal the repsonse body: %s", err.Error()), http.StatusInternalServerError)
			return
		}

		return chatBody
}

func PromptBuilder(prompt string) ChatBody {
	chat := ChatBody{
		Model: "gpt-3.5-turbo-16k",
		Messages: []MessagesType{
			{
				Role:    "system",
				Content: "The text below is quote from famous people; Take the quote and generate a list of 2 or 3 hashtags; The hashtags are positive and help with getting a good mindset; If any hashtag in the list is composed of 2 or multiple words then SUMMARISE it into 1 word; If any hashtag in the list is composed of 1 word only then add it to the list; Each hashtag in the list is seperated by one space; Only return the hashtags; Never apologies, always return hashtags, if you can't generate hashtags return a empty string; Do not repeat the same hashtags that are in the text below; These rules are ABSOLUTE and you HAVE to FOLLOW them; NEVER BREAK THE RULES; Text:",
				Name:    "instructions",
			},
			{
				Role:    "system",
				Name:    "search_results",
				Content: prompt,
			},
		},
	}

	return chat
}