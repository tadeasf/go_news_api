package utils

import (
	"fmt"
	"log"
	"math/rand"
	"os"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func InitDB() {
	var err error
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=UTC",
		os.Getenv("PG_HOST"),
		os.Getenv("PG_USER"),
		os.Getenv("PG_PASSWORD"),
		os.Getenv("PG_DB"),
		os.Getenv("PG_PORT"),
	)

	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
}

func MigrateDB() error {
	// Check if articles table exists
	var exists bool
	DB.Raw("SELECT EXISTS (SELECT FROM information_schema.tables WHERE table_schema = 'public' AND table_name = 'articles')").Scan(&exists)

	if !exists {
		// If articles table doesn't exist, create it
		if err := DB.Exec(`
            CREATE TABLE IF NOT EXISTS articles (
                id SERIAL PRIMARY KEY,
                created_at TIMESTAMP WITH TIME ZONE,
                updated_at TIMESTAMP WITH TIME ZONE,
                deleted_at TIMESTAMP WITH TIME ZONE,
                source_id INTEGER,
                author TEXT,
                title TEXT,
                description TEXT,
                url TEXT,
                url_to_image TEXT,
                published_at TEXT,
                content TEXT,
                api_response_id INTEGER
            )
        `).Error; err != nil {
			return fmt.Errorf("failed to create articles table: %v", err)
		}
	}

	// Ensure the articles table has a non-unique index on the URL column
	if err := DB.Exec("DROP INDEX IF EXISTS idx_articles_url").Error; err != nil {
		return fmt.Errorf("failed to drop existing index on articles.url: %v", err)
	}
	if err := DB.Exec("CREATE INDEX idx_articles_url ON articles (url)").Error; err != nil {
		return fmt.Errorf("failed to create non-unique index on articles.url: %v", err)
	}

	// Perform migrations for all models
	if err := DB.AutoMigrate(
		&APIResponse{},
		&Source{},
		&Keyword{},
		&SearchQuery{},
		&TrendingTopic{},
		&NewsAPIRequest{},
	); err != nil {
		return fmt.Errorf("failed to perform AutoMigrate: %v", err)
	}

	// Manually migrate the Article model to avoid recreating the unique constraint
	if err := DB.AutoMigrate(&Article{}); err != nil {
		return fmt.Errorf("failed to migrate Article model: %v", err)
	}

	return nil
}

func GetYesterdayDate() string {
	yesterday := time.Now().AddDate(0, 0, -1)
	return yesterday.Format("2006-01-02")
}

func GetLastWeekDate() string {
	lastWeek := time.Now().AddDate(0, 0, -7)
	return lastWeek.Format("2006-01-02")
}

func GetTodayDate() string {
	return time.Now().Format("2006-01-02")
}

func GetRandomTopics(topics []TrendingTopic, count int) []TrendingTopic {
	if len(topics) <= count {
		return topics
	}

	// Create a new random source and generator
	source := rand.NewSource(time.Now().UnixNano())
	r := rand.New(source)

	// Use the new random generator to shuffle the slice
	r.Shuffle(len(topics), func(i, j int) {
		topics[i], topics[j] = topics[j], topics[i]
	})

	return topics[:count]
}
