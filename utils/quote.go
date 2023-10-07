package utils

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
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

func GetRandomQuote(w http.ResponseWriter) (quote RandomQuote, hashtags []string) {
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

	if err := json.Unmarshal(bodyBytes, &quote); err != nil {
		http.Error(w, fmt.Sprintf("Failed to Unmarshal the repsonse body: %s", err.Error()), http.StatusInternalServerError)
		return
	}

	// 1.1 Format hashtags
	replacer := strings.NewReplacer(" ", "")

	for _, word := range quote.Tags {
		hashtag := "#" + replacer.Replace(word)
		hashtags = append(hashtags, hashtag)
	}

	return quote, hashtags
}