package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
)

type RandomQuote struct {
	ID string `json:"_id"`
	Content string `json:"content"`
	Author string `json:"author"`
	Tags []string `json:"tags"`
	AuthorSlug string `json:"authorSlug"`
	Length int `json:"length"`
	DateAdded string `json:"dateAdded"`
	DateModified string `json:"dateModified"`
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

type Choices struct {
	Index int `json:"index"`
	Message struct {
				Role string `json:"role"`
				Content string `json:"content"`
			}
	FinishReason string `json:"finish_reason"`
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

func main() {
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// 1 - Get random quote
		resp, err := http.Get("https://api.quotable.io/random")
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to make a request: %s", err.Error()), http.StatusInternalServerError)
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			http.Error(w, "Reponse not OK from Random Quotes API", http.StatusInternalServerError)
			return
		}

		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to read response body: %s", err.Error()), http.StatusInternalServerError)
			return
		}
		
		var quote = &RandomQuote{}
		if err := json.Unmarshal(bodyBytes, &quote); err != nil {
			http.Error(w, fmt.Sprintf("Failed to Unmarshal the repsonse body: %s", err.Error()), http.StatusInternalServerError)
			return
		}
		
		// 1.1 Format hashtags
		hashtags := []string{}
		replacer := strings.NewReplacer(" ", "")

		for _, word := range quote.Tags {
			hashtag := "#" + replacer.Replace(word)
        	hashtags = append(hashtags, hashtag)
		}

		// 2 - Use chatGPT for something
		instructions := PromptBuilder(fmt.Sprintf("%s - %s", quote.Content, quote.Author))
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
		fmt.Println(chatResp.Body)
		
		if chatResp.StatusCode != http.StatusOK {
			http.Error(w, fmt.Sprintf("Request failed with status code: %v", chatResp.StatusCode), http.StatusInternalServerError)
			return
		}
		
		chatBodyBytes, err := io.ReadAll(chatResp.Body)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to read response body: %s", err.Error()), http.StatusInternalServerError)
			return
		}
		
		var chatBody = &AiResponse{}
		if err := json.Unmarshal(chatBodyBytes, &chatBody); err != nil {
			http.Error(w, fmt.Sprintf("Failed to Unmarshal the repsonse body: %s", err.Error()), http.StatusInternalServerError)
			return
		}

		// 3 - Build the tweet
		// generatedHashtags := strings.Split(chatBody.Choices[0].Message.Content, " ")
		// concatHashtags := removeDuplicate(append(generatedHashtags, hashtags...))
		formattedHashtags := strings.Join(hashtags, " ")

		var result string
		result = fmt.Sprintf("%s - %s \n%s #DailyQuotes", quote.Content, quote.Author, formattedHashtags)

		// 4 - Return a formated text for the tweet
		w.Header().Set("Content-Type", "text/plain")
		w.Write([]byte(result))
    })

	 port := os.Getenv("PORT")

    log.Printf("Listening on PORT %s", port)
    log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", port), nil))
}

func PromptBuilder(prompt string) ChatBody {
	chat := ChatBody{
		Model: "gpt-3.5-turbo-16k",
		Messages: []MessagesType{
			{
				Role:    "system",
				Content: "The text below is quote from famous people; Take the quote and generate a list of 2 or 3 hashtags; The hashtags are positive and help with to get a good mindset; Each hashtag in the list is seperated by one space; Only return the hashtags; Text:",
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

func removeDuplicate[T string | int](sliceList []T) []T {
	allKeys := make(map[T]bool)
	list := []T{}
	for _, item := range sliceList {
		 if _, value := allKeys[item]; !value {
			  allKeys[item] = true
			  list = append(list, item)
		 }
	}
	return list
}