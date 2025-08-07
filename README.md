# Gator RSS Aggregator

A command-line RSS feed aggregator written in Go that allows users to manage RSS feeds, follow feeds, and browse posts from subscribed feeds.

## Features

- User management with registration and login
- Add and manage RSS feeds
- Follow/unfollow specific feeds per user
- Browse recent posts from subscribed feeds
- Automatic feed aggregation at configurable intervals
- Multi-user support with independent subscriptions

## Requirements

- Go 1.19 or higher
- PostgreSQL 12 or higher
- sqlc (SQL code generator)
- goose (database migration tool)

## Documentation

- [Installation Guide](docs/INSTALLATION.md) - Complete setup instructions
- [Usage Guide](docs/USAGE.md) - Commands and examples  
- [Troubleshooting Guide](docs/TROUBLESHOOTING.md) - Common issues and solutions

## Quick Example

```bash
# Register a new user
./gator register alice

# Add and follow an RSS feed
./gator addfeed "Hacker News" "https://feeds.ycombinator.com/news.rss"

# Start background aggregation
./gator agg 5m &

# Browse recent posts
./gator browse 10
```

## Development

### Database Schema Changes

1. Create a new migration file in `sql/schema/`
2. Run the migration with goose
3. Update SQL queries in `sql/queries/`
4. Regenerate Go code with `sqlc generate`

### Adding Commands

1. Add SQL queries to appropriate files in `sql/queries/`
2. Run `sqlc generate` to update database code
3. Implement handler function in `main.go`
4. Register the command in the `main()` function

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Test thoroughly and run `sqlc generate` if needed
5. Submit a pull request

## 📚 About This Project

This project was built as part of the **backend development curriculum** on [Boot.dev](https://boot.dev) - an online platform focused on teaching practical backend skills through hands-on coding projects.

### What I learned:
- Go programming fundamentals
- PostgreSQL database design
- SQL query optimization with sqlc
- RSS feed parsing and aggregation
- CLI application development

## License

This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for details.
