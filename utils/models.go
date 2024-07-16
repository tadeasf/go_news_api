package utils

import (
	"encoding/json"
	"time"

	"gorm.io/gorm"
)

type APIResponse struct {
	gorm.Model
	Status        string    `json:"status"`
	TotalResults  int       `json:"totalResults,omitempty"`
	TotalArticles int       `json:"totalArticles,omitempty"`
	Articles      []Article `gorm:"foreignKey:APIResponseID"`
	APISource     string    `json:"api_source"` // "gnews" or "newsapi"
	Type          string    `json:"type"`       // "category" or "topic"
	Topic         string    `json:"topic,omitempty"`
	Page          int       `json:"page,omitempty"`
	PerPage       int       `json:"per_page,omitempty"`
}
type Article struct {
	gorm.Model
	Source        Source `json:"source" gorm:"foreignKey:SourceID"`
	SourceID      uint
	Author        string `json:"author"`
	Title         string `json:"title"`
	Description   string `json:"description"`
	URL           string `json:"url" gorm:"index:idx_articles_url,priority:1"`
	URLToImage    string `json:"urlToImage"`
	PublishedAt   string `json:"publishedAt"`
	Content       string `json:"content"`
	APIResponseID uint
	Keywords      []Keyword `gorm:"many2many:article_keywords;"`
	Language      string    `json:"language,omitempty"`
}

func (a *Article) UnmarshalJSON(data []byte) error {
	type Alias Article
	aux := &struct {
		*Alias
		Source struct {
			ID   interface{} `json:"id"`
			Name string      `json:"name"`
			URL  string      `json:"url"`
		} `json:"source"`
	}{
		Alias: (*Alias)(a),
	}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	a.Source.ExternalID = aux.Source.ID
	a.Source.Name = aux.Source.Name
	a.Source.URL = aux.Source.URL
	return nil
}

type Source struct {
	gorm.Model
	ExternalID interface{} `json:"id" gorm:"-"` // Use interface{} to accept both string and int
	Name       string      `json:"name"`
	URL        string      `json:"url"`
}

type Keyword struct {
	gorm.Model
	Word string `json:"word"`
}

type SearchQuery struct {
	gorm.Model
	Query       string    `json:"query"`
	SearchedAt  time.Time `json:"searched_at"`
	ResultCount int       `json:"result_count"`
}

type NewsAPIResponse struct {
	Status       string `json:"status"`
	TotalResults int    `json:"totalResults"`
	Articles     []struct {
		Source      Source `json:"source"`
		Author      string `json:"author"`
		Title       string `json:"title"`
		Description string `json:"description"`
		URL         string `json:"url"`
		URLToImage  string `json:"urlToImage"`
		PublishedAt string `json:"publishedAt"`
		Content     string `json:"content"`
	} `json:"articles"`
}

type GNewsResponse struct {
	TotalArticles int       `json:"totalArticles"`
	Articles      []Article `json:"articles"`
}

type TrendingTopic struct {
	gorm.Model
	Topic        string  `json:"topic"`
	SearchGrowth string  `json:"search_growth"`
	GrowthValue  float64 `json:"growth_value"`
}

type NewsAPIRequest struct {
	gorm.Model
	Topic       string `gorm:"index"`
	Source      string `gorm:"index"`
	RequestedAt time.Time
}

// @model SwaggerAPIResponse
type SwaggerAPIResponse struct {
	Status        string    `json:"status"`
	TotalResults  int       `json:"totalResults"`
	TotalArticles int       `json:"totalArticles"`
	Articles      []Article `json:"articles"`
	APISource     string    `json:"apiSource"`
	Page          int       `json:"page,omitempty"`
	PerPage       int       `json:"per_page,omitempty"`
}
