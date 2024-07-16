package endpoints

import (
	"fmt"
	"go_news_api/utils"
	"strings"
	"time"
)

// requestCount keeps track of the number of requests made
var requestCount int

// maxRequestsPerDay is the maximum number of requests allowed per day
const maxRequestsPerDay = 100

func CheckRequestLimit() error {
	// Reset requestCount at midnight
	now := time.Now()
	if now.Hour() == 0 && now.Minute() == 0 {
		requestCount = 0
	}

	// Check if the request limit has been reached
	if requestCount >= maxRequestsPerDay {
		return fmt.Errorf("request limit reached for today")
	}

	// Increment requestCount
	requestCount++
	return nil
}

// Helper function to get topic names from TrendingTopic slice
func GetTopicNames(topics []utils.TrendingTopic) []string {
	var names []string
	for _, topic := range topics {
		names = append(names, topic.Topic)
	}
	return names
}

func CombineAPIResponses(responses []utils.APIResponse) utils.APIResponse {
	var combinedArticles []utils.Article
	var totalResults int

	for _, response := range responses {
		combinedArticles = append(combinedArticles, response.Articles...)
		totalResults += response.TotalResults
	}

	return utils.APIResponse{
		Articles:  combinedArticles,
		APISource: "combined",
	}
}

// Helper function to extract keywords from text
func ExtractKeywords(text string) []string {
	// This is a simple implementation. You might want to use a more sophisticated method or library.
	words := strings.Fields(strings.ToLower(text))
	uniqueWords := make(map[string]bool)
	for _, word := range words {
		if len(word) > 3 { // Only consider words longer than 3 characters
			uniqueWords[word] = true
		}
	}
	var keywords []string
	for word := range uniqueWords {
		keywords = append(keywords, word)
	}
	return keywords
}
