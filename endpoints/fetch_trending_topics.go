package endpoints

import (
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"strings"

	"go_news_api/utils"

	"github.com/PuerkitoBio/goquery"
)

func ParseGrowth(growth string) float64 {
	growth = strings.TrimSpace(growth)
	if strings.HasSuffix(growth, "x+") {
		value, _ := strconv.ParseFloat(strings.TrimSuffix(growth, "x+"), 64)
		return value * 100 // Treat "99x+" as 9900%
	}
	value, _ := strconv.ParseFloat(strings.TrimSuffix(growth, "%"), 64)
	return value
}

func FetchTrendingTopics() ([]utils.TrendingTopic, error) {
	resp, err := http.Get("https://explodingtopics.com/blog/trending-topics")
	if err != nil {
		return nil, fmt.Errorf("failed to fetch trending topics: %v", err)
	}
	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to parse HTML: %v", err)
	}

	var trendingTopics []utils.TrendingTopic
	doc.Find("table.tableVariant1 tr").Each(func(i int, s *goquery.Selection) {
		if i == 0 { // Skip header row
			return
		}
		cells := s.Find("td")
		if cells.Length() >= 3 {
			topic := strings.TrimSpace(cells.Eq(1).Text())
			growth := strings.TrimSpace(cells.Eq(2).Text())
			if topic != "" {
				growthValue := ParseGrowth(growth)
				trendingTopics = append(trendingTopics, utils.TrendingTopic{
					Topic:        topic,
					SearchGrowth: growth,
					GrowthValue:  growthValue,
				})
			}
		}
	})

	if len(trendingTopics) == 0 {
		return nil, fmt.Errorf("no trending topics found")
	}

	// Sort topics by growth value in descending order
	sort.Slice(trendingTopics, func(i, j int) bool {
		return trendingTopics[i].GrowthValue > trendingTopics[j].GrowthValue
	})

	// Return up to 100 topics
	if len(trendingTopics) > 100 {
		return trendingTopics[:100], nil
	}
	return trendingTopics, nil
}
