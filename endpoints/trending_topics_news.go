package endpoints

import (
	"fmt"
	"go_news_api/utils"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func GetTopicsCount(c *gin.Context) int {
	topicsCount, err := strconv.Atoi(c.DefaultQuery("topics", "1"))
	if err != nil || topicsCount < 1 || topicsCount > 10 {
		return 1
	}
	return topicsCount
}

func HandleTransactionError(tx *gorm.DB, c *gin.Context) {
	if r := recover(); r != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Panic occurred: %v", r)})
	}
}

func GetSelectedTopics(tx *gorm.DB, topicsCount int) ([]utils.TrendingTopic, error) {
	allTrendingTopics, err := FetchTrendingTopics()
	if err != nil {
		return nil, fmt.Errorf("Failed to fetch trending topics: %v", err)
	}
	return utils.GetRandomTopics(allTrendingTopics, topicsCount), nil
}

func GetOrFetchAPIResponse(tx *gorm.DB, source string, selectedTopics []utils.TrendingTopic) (*utils.APIResponse, error) {
	today := time.Now().Format("2006-01-02")
	existingSearches, err := CheckExistingSearches(tx, selectedTopics, today)
	if err != nil {
		return nil, err
	}

	if len(existingSearches) == len(selectedTopics) {
		return GetExistingAPIResponse(tx, source, today)
	}

	return FetchNewAPIResponse(tx, source, selectedTopics)
}

func CheckExistingSearches(tx *gorm.DB, selectedTopics []utils.TrendingTopic, today string) ([]utils.SearchQuery, error) {
	var existingSearches []utils.SearchQuery
	err := tx.Where("query IN (?) AND DATE(searched_at) = ?", GetTopicNames(selectedTopics), today).Find(&existingSearches).Error
	if err != nil {
		return nil, fmt.Errorf("Failed to check existing searches: %v", err)
	}
	return existingSearches, nil
}

func GetExistingAPIResponse(tx *gorm.DB, source, today string) (*utils.APIResponse, error) {
	var apiResponse utils.APIResponse
	err := tx.Where("api_source = ? AND type = ? AND DATE(created_at) = ?", source, "topic", today).First(&apiResponse).Error
	if err != nil {
		return nil, fmt.Errorf("Failed to retrieve existing results: %v", err)
	}
	return &apiResponse, nil
}

func FetchNewAPIResponse(tx *gorm.DB, source string, selectedTopics []utils.TrendingTopic) (*utils.APIResponse, error) {
	apiResponse, err := FetchAPIResponse(source, selectedTopics)
	if err != nil {
		return nil, err
	}

	err = SaveAPIResponse(tx, apiResponse, selectedTopics)
	if err != nil {
		return nil, err
	}

	return apiResponse, nil
}

func FetchAPIResponse(source string, selectedTopics []utils.TrendingTopic) (*utils.APIResponse, error) {
	var apiResponse *utils.APIResponse
	var err error

	switch source {
	case "newsapi":
		apiResponse, err = GetNewsAPITrendingTopicsNews(selectedTopics)
	case "gnews":
		apiResponse, err = GetGNewsTrendingTopicsNews(selectedTopics)
	default:
		return nil, fmt.Errorf("Invalid source")
	}

	if err != nil {
		return nil, fmt.Errorf("API request failed: %v", err)
	}

	apiResponse.Type = "topic"
	apiResponse.Topic = strings.Join(GetTopicNames(selectedTopics), ", ")

	return apiResponse, nil
}

func SaveAPIResponse(tx *gorm.DB, apiResponse *utils.APIResponse, selectedTopics []utils.TrendingTopic) error {
	var existingResponse utils.APIResponse
	if err := tx.Where(utils.APIResponse{APISource: apiResponse.APISource, Type: apiResponse.Type}).
		Attrs(utils.APIResponse{
			Status:        apiResponse.Status,
			TotalResults:  apiResponse.TotalResults,
			TotalArticles: apiResponse.TotalArticles,
			Topic:         apiResponse.Topic,
		}).FirstOrCreate(&existingResponse).Error; err != nil {
		return fmt.Errorf("Failed to save or update API response: %v", err)
	}

	existingResponse.Articles = apiResponse.Articles
	if err := tx.Save(&existingResponse).Error; err != nil {
		return fmt.Errorf("Failed to update API response articles: %v", err)
	}

	*apiResponse = existingResponse

	if err := SaveSearchQueries(tx, selectedTopics, apiResponse); err != nil {
		return err
	}

	if err := SaveArticles(tx, apiResponse); err != nil {
		return err
	}

	return nil
}

func SaveSearchQueries(tx *gorm.DB, selectedTopics []utils.TrendingTopic, apiResponse *utils.APIResponse) error {
	for _, topic := range selectedTopics {
		searchQuery := utils.SearchQuery{
			Query:       topic.Topic,
			SearchedAt:  time.Now(),
			ResultCount: len(apiResponse.Articles),
		}
		if err := tx.Create(&searchQuery).Error; err != nil {
			return fmt.Errorf("Failed to save search query: %v", err)
		}
	}
	return nil
}

func SaveArticles(tx *gorm.DB, apiResponse *utils.APIResponse) error {
	for _, article := range apiResponse.Articles {
		if err := SaveSource(tx, &article); err != nil {
			return err
		}

		if err := SaveOrUpdateArticle(tx, apiResponse, &article); err != nil {
			return err
		}

		if err := SaveKeywords(tx, &article); err != nil {
			return err
		}
	}
	return nil
}

func SaveSource(tx *gorm.DB, article *utils.Article) error {
	if err := tx.Where(utils.Source{Name: article.Source.Name}).FirstOrCreate(&article.Source).Error; err != nil {
		return fmt.Errorf("Failed to save source: %v", err)
	}
	return nil
}

func SaveOrUpdateArticle(tx *gorm.DB, apiResponse *utils.APIResponse, article *utils.Article) error {
	var existingArticle utils.Article
	result := tx.Where("url = ?", article.URL).First(&existingArticle)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			// Article doesn't exist, create a new one
			article.APIResponseID = apiResponse.ID
			if err := tx.Create(article).Error; err != nil {
				return fmt.Errorf("Failed to create new article: %v", err)
			}
		} else {
			// Some other error occurred
			return fmt.Errorf("Error checking for existing article: %v", result.Error)
		}
	} else {
		// Article exists, update it
		existingArticle.APIResponseID = apiResponse.ID
		existingArticle.Author = article.Author
		existingArticle.Title = article.Title
		existingArticle.Description = article.Description
		existingArticle.PublishedAt = article.PublishedAt
		existingArticle.Content = article.Content
		existingArticle.URLToImage = article.URLToImage
		if err := tx.Save(&existingArticle).Error; err != nil {
			return fmt.Errorf("Failed to update existing article: %v", err)
		}
		*article = existingArticle
	}
	return nil
}

func SaveKeywords(tx *gorm.DB, article *utils.Article) error {
	keywords := ExtractKeywords(article.Title + " " + article.Description)
	for _, word := range keywords {
		var keyword utils.Keyword
		if err := tx.Where(utils.Keyword{Word: word}).FirstOrCreate(&keyword).Error; err != nil {
			return fmt.Errorf("Failed to save keyword: %v", err)
		}
		if err := tx.Model(article).Association("Keywords").Append(&keyword); err != nil {
			return fmt.Errorf("Failed to associate keyword with article: %v", err)
		}
	}
	return nil
}
