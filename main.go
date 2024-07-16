package main

// TODO: add endpoints for fetching news by search query
// labels: endpoint, feature, enhancement

// TODO: add endpoints for fetching news by trending topics
// labels: endpoint, feature, enhancement

// TODO: add endpoints for fetching news by trending categories
// labels: endpoint, feature, enhancement

// TODO: add scheduling via background cronjob or something like that to continuously pull new articles
// labels: enhancement, feature

// TODO: sentiment analysis, topics and keywords extraction, and other NLP features
// labels: feature, enhancement

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"gorm.io/gorm"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	_ "go_news_api/docs"
	"go_news_api/endpoints"
	"go_news_api/utils"
)

// @title News API
// @version 1.0
// @description A simple news API using Gin and external news services
// @host news.tadeasfort.cz
// @BasePath /api/v1
// @schemes https
func main() {

	// Load .env file
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: Error loading .env file:", err)
	}

	// Initialize database connection
	utils.InitDB()

	// Perform automatic migration
	if err := utils.MigrateDB(); err != nil {
		log.Fatalf("Failed to perform database migration: %v", err)
	} else {
		log.Println("Database migration successful")
	}

	r := gin.Default()

	// Add this new route handler for the root path
	r.GET("/", func(c *gin.Context) {
		c.Redirect(http.StatusMovedPermanently, "/docs/index.html")
	})

	// API v1 group
	v1 := r.Group("/api/v1")
	{
		v1.GET("/health", healthCheck)
		v1.GET("/test-postgresql", testPostgreSQL)
		v1.POST("/init-db", initializeDatabase)
		v1.GET("/migrate", migrateDatabase)
		v1.GET("/top-headlines", getTopHeadlines)
		v1.GET("/trending-topics", getTrendingTopicsNews)
		v1.GET("/fetch-trending-categories", fetchTrendingCategories)
		v1.GET("/news-by-keyword", getNewsByKeyword)
	}

	// Modify the Swagger documentation route
	r.GET("/docs/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Set Gin mode and port based on GIN_MODE
	ginMode := os.Getenv("GIN_MODE")
	port := ":8824"

	if ginMode == "release" {
		gin.SetMode(gin.ReleaseMode)
		port = ":8823"
	} else {
		gin.SetMode(gin.DebugMode)
	}

	log.Printf("Running in %s mode on port %s", ginMode, port)

	if err := r.Run(port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

// @Summary Health check
// @Description Check if the API is up and running
// @Produce json
// @Success 200 {object} map[string]string
// @Router /health [get]
func healthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "UP",
	})
}

// @Summary Test PostgreSQL connection
// @Description Test if the connection to PostgreSQL is working
// @Produce json
// @Success 200 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /test-postgresql [get]
func testPostgreSQL(c *gin.Context) {
	sqlDB, err := utils.DB.DB()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Failed to get database instance",
			"error":   err.Error(),
		})
		return
	}

	if err := sqlDB.Ping(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Failed to ping PostgreSQL",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Successfully connected to PostgreSQL",
	})
}

// @Summary Initialize database
// @Description Create tables in the database
// @Produce json
// @Success 200 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /init-db [post]
func initializeDatabase(c *gin.Context) {
	// Drop existing tables
	err := utils.DB.Migrator().DropTable(&utils.APIResponse{}, &utils.Article{}, &utils.Source{}, &utils.Keyword{}, &utils.SearchQuery{}, &utils.TrendingTopic{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Failed to drop existing tables",
			"error":   err.Error(),
		})
		return
	}

	// Create new tables
	err = utils.DB.AutoMigrate(&utils.APIResponse{}, &utils.Article{}, &utils.Source{}, &utils.Keyword{}, &utils.SearchQuery{}, &utils.TrendingTopic{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Failed to create new tables",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Database initialized successfully",
	})
}

// @Summary Migrate database
// @Description Run database migrations
// @Produce json
// @Success 200 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /migrate [get]
func migrateDatabase(c *gin.Context) {
	err := utils.DB.AutoMigrate(&utils.APIResponse{}, &utils.Article{}, &utils.Source{}, &utils.Keyword{}, &utils.SearchQuery{}, &utils.TrendingTopic{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Failed to run migrations",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Migrations completed successfully",
	})
}

// @Summary Get top headlines
// @Description Get top headlines from News API and GNews
// @Produce json
// @Param source query string false "Source of news (newsapi or gnews)"
// @Param country query string false "Country code for headlines"
// @Param category query string false "Category of news"
// @Success 200 {object} utils.SwaggerAPIResponse
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /top-headlines [get]
func getTopHeadlines(c *gin.Context) {
	source := c.DefaultQuery("source", "newsapi")
	country := c.DefaultQuery("country", "us")
	category := c.DefaultQuery("category", "general")

	var apiResponse *utils.APIResponse
	var err error

	switch source {
	case "newsapi":
		apiResponse, err = endpoints.GetNewsAPITopHeadlinesByCategory(country, category)
		// Handle error
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	case "gnews":
		apiResponse, err = endpoints.GetGNewsTopHeadlines(country, category)
		// Handle error
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid source"})
		return
	}

	// Save the API response to the database
	var existingResponse utils.APIResponse
	result := utils.DB.Where("api_source = ?", apiResponse.APISource).First(&existingResponse)
	if result.Error == nil {
		// Update existing record
		existingResponse.Status = apiResponse.Status
		existingResponse.TotalResults = apiResponse.TotalResults
		existingResponse.TotalArticles = apiResponse.TotalArticles
		existingResponse.Articles = apiResponse.Articles
		err = utils.DB.Save(&existingResponse).Error
	} else if result.Error == gorm.ErrRecordNotFound {
		// Create new record
		err = utils.DB.Create(apiResponse).Error
	} else {
		err = result.Error
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, apiResponse)
}

// @Summary Get trending topics news
// @Description Get news articles for trending topics from News API and GNews
// @Produce json
// @Param source query string false "Source of news (newsapi or gnews)"
// @Param topics query int false "Number of random topics to pick (1-10, default 1)"
// @Success 200 {object} utils.SwaggerAPIResponse
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /trending-topics [get]
func getTrendingTopicsNews(c *gin.Context) {
	source := c.DefaultQuery("source", "newsapi")
	topicsCount := endpoints.GetTopicsCount(c)

	tx := utils.DB.Begin()
	defer endpoints.HandleTransactionError(tx, c)

	selectedTopics, err := endpoints.GetSelectedTopics(tx, topicsCount)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	apiResponse, err := endpoints.GetOrFetchAPIResponse(tx, source, selectedTopics)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if err := tx.Commit().Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to commit transaction: %v", err)})
		return
	}

	c.JSON(http.StatusOK, apiResponse)
}

// @Summary Fetch trending categories
// @Description Fetch top 10 trending categories from Exploding Topics
// @Produce json
// @Success 200 {array} utils.TrendingTopic
// @Failure 500 {object} map[string]string
// @Router /fetch-trending-categories [get]
func fetchTrendingCategories(c *gin.Context) {
	trendingTopics, err := endpoints.FetchTrendingTopics()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Begin a transaction
	tx := utils.DB.Begin()

	// Clear existing trending topics
	if err := tx.Where("1 = 1").Delete(&utils.TrendingTopic{}).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to clear existing trending topics"})
		return
	}

	// Save new trending topics
	for _, topic := range trendingTopics {
		if err := tx.Create(&topic).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save trending topics"})
			return
		}
	}

	// Commit the transaction
	if err := tx.Commit().Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to commit transaction"})
		return
	}

	c.JSON(http.StatusOK, trendingTopics)
}

// @Summary Get news by keyword
// @Description Get news articles for a specific keyword from News API and GNews
// @Produce json
// @Param source query string false "Source of news (newsapi or gnews)"
// @Param keyword query string true "Keyword to search for"
// @Success 200 {object} utils.SwaggerAPIResponse
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /news-by-keyword [get]
func getNewsByKeyword(c *gin.Context) {
	endpoints.GetNewsByKeyword(c)
}
