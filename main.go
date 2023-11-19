package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/Ofsen/random-quotes-api/utils"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	router := chi.NewRouter()
	router.Use(middleware.Logger)

	router.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	router.Get("/", func(w http.ResponseWriter, r *http.Request) {
		// 1 - Get random quote
		quote, hashtags := utils.GetRandomQuote(w)

		// 2 - Use AI to generate more hashtags
		chatBody := utils.GenerateHashtags(w, quote, strings.Join(hashtags, " "))

		// 3 - Build the tweet
		generatedHashtags := strings.Split(chatBody.Choices[0].Message.Content, " ")
		concatHashtags := removeDuplicate(append(generatedHashtags, hashtags...))
		formattedHashtags := strings.Join(concatHashtags, " ")

		var result string
		result = fmt.Sprintf("%s - %s \n%s #DailyQuotes", quote.Content, quote.Author, formattedHashtags)

		// 4 - Return a formated text for the tweet
		w.Header().Set("Content-Type", "text/plain")
		w.Write([]byte(result))
    })

	 port := os.Getenv("PORT")

    log.Printf("Listening on PORT %s", port)
    log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", port), router))
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
