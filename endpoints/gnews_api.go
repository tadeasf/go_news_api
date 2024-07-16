package endpoints

import (
	"encoding/json"
	"fmt"
	"go_news_api/utils"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
)

func GetGNewsTopHeadlines(country, category string) (*utils.APIResponse, error) {
	apikey := os.Getenv("GNEWS_API_KEY")
	url := fmt.Sprintf("https://gnews.io/api/v4/top-headlines?category=%s&lang=en&country=%s&max=10&apikey=%s", category, country, apikey)
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var gNewsResponse utils.GNewsResponse
	err = json.Unmarshal(body, &gNewsResponse)
	if err != nil {
		return nil, err
	}

	apiResponse := &utils.APIResponse{
		TotalArticles: gNewsResponse.TotalArticles,
		Articles:      gNewsResponse.Articles,
		APISource:     "gnews",
	}

	return apiResponse, nil
}

func GetGNewsTrendingTopicsNews(topics []utils.TrendingTopic) (*utils.APIResponse, error) {
	var gNewsResponses []utils.APIResponse
	for _, topic := range topics {
		// Make API call for each trending topic
		gNewsResponse, err := GetGNewsSearchByTopic(topic.Topic)
		if err != nil {
			return nil, err
		}
		gNewsResponses = append(gNewsResponses, *gNewsResponse)
	}

	// Combine all responses into a single APIResponse
	combinedResponse := CombineAPIResponses(gNewsResponses)

	// Create a new APIResponse
	apiResponse := &utils.APIResponse{
		TotalArticles: combinedResponse.TotalArticles,
		Articles:      combinedResponse.Articles,
		APISource:     "gnews",
	}

	return apiResponse, nil
}

func GetGNewsSearchByTopic(topic string) (*utils.APIResponse, error) {
	apikey := os.Getenv("GNEWS_API_KEY")

	// Create a url.Values to hold the query parameters
	params := url.Values{}
	params.Add("q", topic)
	params.Add("lang", "en")
	params.Add("country", "us")
	params.Add("max", "10")
	params.Add("apikey", apikey)
	params.Add("from", utils.GetYesterdayDate())
	params.Add("to", utils.GetTodayDate())

	// Construct the URL with properly encoded query parameters
	baseURL := "https://gnews.io/api/v4/search"
	fullURL := baseURL + "?" + params.Encode()

	// Log the URL (remove sensitive information like API key in production)
	logURL := strings.Replace(fullURL, apikey, "REDACTED", 1)
	log.Printf("GNews API request URL: %s", logURL)

	resp, err := http.Get(fullURL)
	if err != nil {
		return nil, fmt.Errorf("HTTP request failed: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %v", err)
	}

	// Log the response status code and headers
	log.Printf("GNews API response status: %s", resp.Status)
	log.Printf("GNews API response headers: %v", resp.Header)

	// Check if the response is HTML (error page)
	if strings.Contains(string(body), "<html") {
		return nil, fmt.Errorf("received HTML response instead of JSON. Status: %s, Response: %s", resp.Status, string(body))
	}

	var gNewsResponse utils.GNewsResponse
	err = json.Unmarshal(body, &gNewsResponse)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON: %v. Response body: %s", err, string(body))
	}

	apiResponse := &utils.APIResponse{
		TotalArticles: gNewsResponse.TotalArticles,
		Articles:      gNewsResponse.Articles,
		APISource:     "gnews",
	}

	return apiResponse, nil
}
