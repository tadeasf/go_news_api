# Go News API

A robust and scalable news aggregation API built with Go, leveraging multiple news sources to provide trending topics, top headlines, and more.

## News && TODO

### Latest Changes

* 🔧 (automerge.yml): update automerge workflow to trigger on more events and improve merge commit message. PR [#16](https://github.com/tadeasf/go_news_api/pull/16) by [@tadeasf](https://github.com/tadeasf).

#### Security Fixes

* 🔧 (workflows): add GitHub Actions workflows for build, release, and changelog generation. PR [#17](https://github.com/tadeasf/go_news_api/pull/17) by [@tadeasf](https://github.com/tadeasf).

#### Fixes

* Fix-todo-to-issue-again. PR [#9](https://github.com/tadeasf/go_news_api/pull/9) by [@tadeasf](https://github.com/tadeasf).
* 🔧 (todo-to-issue.yml): remove language-specific configuration for Go. PR [#8](https://github.com/tadeasf/go_news_api/pull/8) by [@tadeasf](https://github.com/tadeasf).
* 🔧 (todo-to-issue.yml): fix escape sequences in language block comment patterns. PR [#7](https://github.com/tadeasf/go_news_api/pull/7) by [@tadeasf](https://github.com/tadeasf).
* 🔧 (latest-changes.yml): ensure "Latest Changes" section exists in README. PR [#6](https://github.com/tadeasf/go_news_api/pull/6) by [@tadeasf](https://github.com/tadeasf).

### TODO

- Add endpoints for fetching news by keyword
- Add endpoints for fetching news by search query
- Add endpoints for fetching news by trending categories
- Implement scheduling via background cronjob for continuous article pulling
- Implement sentiment analysis, topics, and keywords extraction

## Features

- Fetch top headlines from multiple news sources (NewsAPI and GNews)
- Get news articles for trending topics
- Fetch trending categories from Exploding Topics
- Database integration with PostgreSQL for caching and data persistence
- API request limiting to comply with external API usage restrictions
- Swagger documentation for easy API exploration

## Endpoints

1. `GET /api/v1/health`: Health check endpoint
2. `GET /api/v1/test-postgresql`: Test PostgreSQL connection
3. `POST /api/v1/init-db`: Initialize database tables
4. `GET /api/v1/migrate`: Run database migrations
5. `GET /api/v1/top-headlines`: Get top headlines from NewsAPI or GNews
6. `GET /api/v1/trending-topics`: Get news articles for trending topics
7. `GET /api/v1/fetch-trending-categories`: Fetch top 10 trending categories

For detailed API documentation, visit the Swagger UI at `/docs/index.html` when running the server.

## How to develop

1. Clone the repository
2. Install dependencies: `go mod download`
3. Create a `.env` file in the root directory with the following variables:

   ```sh
   PG_HOST=your_postgres_host
   PG_USER=your_postgres_user
   PG_PASSWORD=your_postgres_password
   PG_DB=your_postgres_database
   PG_PORT=your_postgres_port
   NEWS_API_KEY=your_newsapi_key
   GNEWS_API_KEY=your_gnews_key
   GIN_MODE=debug
   ```

4. Run the server: `go run main.go`

## How to test

Currently, there are no automated tests implemented. This is an area for future improvement.

## How to build

To build the project, run the following command in the root directory:

```sh
go build -o go-news-api
```

This will create an executable named `go-news-api` in the current directory.

## Contributing

Contributions are welcome! Please follow these steps to contribute:

1. Fork the repository
2. Create a new branch for your feature or fix
3. Make changes and commit them with a descriptive commit message
4. Push your changes to your forked repository
5. Open a pull request to the main repository
