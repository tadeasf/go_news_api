package endpoints

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"go_news_api/utils"
)

// GetNewsAPITopHeadlinesByCategory fetches top headlines from News API
func GetNewsAPITopHeadlinesByCategory(country, category string) (*utils.APIResponse, error) {
	// Check and reset request count if necessary
	if err := CheckRequestLimit(); err != nil {
		return nil, err
	}

	url := fmt.Sprintf("https://newsapi.org/v2/top-headlines?country=%s&category=%s&apiKey=%s", country, category, os.Getenv("NEWS_API_KEY"))
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// Check the response content type
	contentType := resp.Header.Get("Content-Type")
	if strings.Contains(contentType, "application/xml") {
		// Handle XML response
		return nil, fmt.Errorf("received XML response from NewsAPI: %s", string(body))
	}

	var newsAPIResponse utils.NewsAPIResponse
	err = json.Unmarshal(body, &newsAPIResponse)
	if err != nil {
		return nil, err
	}

	var articles []utils.Article
	for _, article := range newsAPIResponse.Articles {
		articles = append(articles, utils.Article{
			Source:      article.Source,
			Author:      article.Author,
			Title:       article.Title,
			Description: article.Description,
			URL:         article.URL,
			PublishedAt: article.PublishedAt,
			Content:     article.Content,
		})
	}

	apiResponse := &utils.APIResponse{
		Status:       newsAPIResponse.Status,
		TotalResults: newsAPIResponse.TotalResults,
		Articles:     articles,
		APISource:    "newsapi",
	}

	return apiResponse, nil
}

// GetNewsAPITrendingTopicsNews fetches news for trending topics from News API
func GetNewsAPITrendingTopicsNews(topics []utils.TrendingTopic) (*utils.APIResponse, error) {
	// Check and reset request count if necessary
	if err := CheckRequestLimit(); err != nil {
		return nil, err
	}

	var newsAPIResponses []utils.APIResponse
	for _, topic := range topics {
		// Check if we've already made this request recently
		var existingRequest utils.NewsAPIRequest
		if err := utils.DB.Where("topic = ? AND source = ? AND requested_at > ?", topic.Topic, "newsapi", time.Now().AddDate(0, 0, -7)).First(&existingRequest).Error; err == nil {
			// We've already made this request in the last week, skip it
			continue
		}

		// Prepare the search query
		query := url.QueryEscape(strings.Join(strings.Fields(topic.Topic), " OR "))

		// Make API call for each trending topic
		newsAPIResponse, err := GetNewsAPIEverythingByTopic(query)
		if err != nil {
			return nil, err
		}

		if len(newsAPIResponse.Articles) > 0 {
			newsAPIResponses = append(newsAPIResponses, *newsAPIResponse)
		}

		// Store the request
		utils.DB.Create(&utils.NewsAPIRequest{
			Topic:       topic.Topic,
			Source:      "newsapi",
			RequestedAt: time.Now(),
		})
	}

	// If no articles were found for any topic, return an empty response
	if len(newsAPIResponses) == 0 {
		return &utils.APIResponse{
			Status:       "ok",
			TotalResults: 0,
			Articles:     []utils.Article{},
			APISource:    "newsapi",
			Type:         "topic",
			Topic:        strings.Join(GetTopicNames(topics), ", "),
		}, nil
	}

	// Combine all responses into a single APIResponse
	combinedResponse := CombineAPIResponses(newsAPIResponses)

	// Create a new APIResponse
	apiResponse := &utils.APIResponse{
		Status:       combinedResponse.Status,
		TotalResults: combinedResponse.TotalResults,
		Articles:     combinedResponse.Articles,
		APISource:    "newsapi",
		Type:         "topic",
		Topic:        strings.Join(GetTopicNames(topics), ", "),
	}

	return apiResponse, nil
}

// getNewsAPIEverythingByTopic fetches everything from News API for a given topic
func GetNewsAPIEverythingByTopic(topic string) (*utils.APIResponse, error) {
	apiKey := os.Getenv("NEWS_API_KEY")
	url := fmt.Sprintf("https://newsapi.org/v2/everything?q=%s&from=%s&to=%s&sortBy=popularity&apiKey=%s&language=en",
		topic, utils.GetLastWeekDate(), utils.GetTodayDate(), apiKey)

	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// Log the response for debugging
	log.Printf("NewsAPI response for topic '%s': %s", topic, string(body))

	var newsAPIResponse utils.NewsAPIResponse
	err = json.Unmarshal(body, &newsAPIResponse)
	if err != nil {
		return nil, err
	}

	var articles []utils.Article
	for _, article := range newsAPIResponse.Articles {
		articles = append(articles, utils.Article{
			Source:      article.Source,
			Author:      article.Author,
			Title:       article.Title,
			Description: article.Description,
			URL:         article.URL,
			PublishedAt: article.PublishedAt,
			Content:     article.Content,
			URLToImage:  article.URLToImage,
		})
	}

	apiResponse := &utils.APIResponse{
		Status:       newsAPIResponse.Status,
		TotalResults: newsAPIResponse.TotalResults,
		Articles:     articles,
		APISource:    "newsapi",
	}

	return apiResponse, nil
}
