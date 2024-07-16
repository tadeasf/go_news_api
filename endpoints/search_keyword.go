package endpoints

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"go_news_api/utils"

	"github.com/gin-gonic/gin"
)

// GetNewsByKeyword handles the request for news articles by keyword
func GetNewsByKeyword(c *gin.Context) {
	keyword := c.Query("keyword")
	page, perPage := GetPaginationParams(c)

	if keyword == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Keyword is required"})
		return
	}

	searchQuery := PrepareSearchQuery(keyword)
	articles, total, err := SearchArticles(searchQuery, page, perPage)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	apiResponse := CreateAPIResponse(articles, total, page, perPage)
	c.JSON(http.StatusOK, apiResponse)
}

func GetPaginationParams(c *gin.Context) (int, int) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "20"))

	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 20
	}

	return page, perPage
}

func PrepareSearchQuery(keyword string) string {
	return strings.Join(strings.Fields(keyword), " & ")
}

func SearchArticles(searchQuery string, page, perPage int) ([]utils.Article, int64, error) {
	offset := (page - 1) * perPage
	var articles []utils.Article
	var total int64

	query := utils.DB.Model(&utils.Article{}).
		Joins("LEFT JOIN sources ON articles.source_id = sources.id").
		Where("to_tsvector('english', articles.author || ' ' || articles.title || ' ' || articles.description || ' ' || articles.content) @@ to_tsquery('english', ?)", searchQuery).
		Order(fmt.Sprintf("ts_rank(to_tsvector('english', articles.author || ' ' || articles.title || ' ' || articles.description || ' ' || articles.content), to_tsquery('english', '%s')) DESC", searchQuery))

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	result := query.Preload("Source").
		Offset(offset).
		Limit(perPage).
		Find(&articles)

	if result.Error != nil {
		return nil, 0, result.Error
	}

	return articles, total, nil
}

func CreateAPIResponse(articles []utils.Article, total int64, page, perPage int) *utils.APIResponse {
	return &utils.APIResponse{
		Status:        "ok",
		TotalResults:  int(total),
		TotalArticles: len(articles),
		Articles:      articles,
		Page:          page,
		PerPage:       perPage,
	}
}
