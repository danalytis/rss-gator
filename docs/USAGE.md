# Usage Guide

This guide covers all the commands and features available in the Gator RSS Aggregator.

## Quick Start

Before using Gator, ensure you've completed the [Installation Guide](INSTALLATION.md).

## Command Reference

All commands follow the pattern: `./gator <command> [arguments]`

### User Management Commands

#### `register <username>`
Register a new user and automatically log in as that user.

```bash
./gator register alice
# Output: successfully registered user: alice
```

**Notes:**
- Username must be unique
- User is automatically logged in after registration
- Configuration file is updated with the new user

#### `login <username>`
Switch to an existing user.

```bash
./gator login alice
# Output: User set to 'alice'
```

**Notes:**
- User must exist in the database
- Updates configuration file with current user
- Required before using user-specific commands

#### `users`
List all registered users, highlighting the current user.

```bash
./gator users
# Output:
# * alice (current)
# * bob
# * charlie
```

#### `reset`
Remove test users from the database (removes users: 'kahya', 'holgith', 'ballan').

```bash
./gator reset
# Output: database reset
```

**Warning:** This is primarily for development/testing purposes.

### Feed Management Commands

#### `addfeed <name> <url>`
Add a new RSS feed and automatically follow it.

```bash
./gator addfeed "Hacker News" "https://feeds.ycombinator.com/news.rss"
```

**Notes:**
- Requires active user login
- Creates feed if it doesn't exist
- Automatically creates a follow relationship
- Handles duplicate URLs gracefully

#### `feeds`
List all feeds in the system with their creators.

```bash
./gator feeds
# Output:
# List of Feeds:
# Feed Name: Hacker News
# Feed URL: https://feeds.ycombinator.com/news.rss
# Created by: alice
# 
# Feed Name: Go Blog
# Feed URL: https://go.dev/blog/feed.atom
# Created by: bob
```

#### `follow <url>`
Follow an existing feed.

```bash
./gator follow "https://go.dev/blog/feed.atom"
# Output: You (alice) are now following Go Blog
```

**Notes:**
- Requires active user login
- Feed must already exist in the system
- Use `addfeed` to add new feeds

#### `following`
List all feeds that the current user is following.

```bash
./gator following
# Output:
# Feeds followed by user alice
# - Hacker News
# - Go Blog
```

**Notes:**
- Requires active user login
- Shows only feeds followed by the current user

#### `unfollow <url>`
Stop following a specific feed.

```bash
./gator unfollow "https://go.dev/blog/feed.atom"
# Output: https://go.dev/blog/feed.atom - Feed unfollowed
```

**Notes:**
- Requires active user login
- Does not delete the feed, only removes the follow relationship
- Other users can still follow the same feed

### Post Browsing Commands

#### `browse [limit]`
View recent posts from followed feeds.

```bash
# Browse default number of posts (2)
./gator browse

# Browse specific number of posts
./gator browse 5
```

**Example Output:**
```
Title: New Go Release Available
Published: 2025-08-06 10:30:00 +0000 UTC
Description: Go 1.21.5 has been released with important security updates...
URL: https://go.dev/blog/go1.21.5
---
Title: Building Scalable Web Services
Published: 2025-08-05 15:45:00 +0000 UTC  
Description: Learn how to build web services that can handle millions of requests...
URL: https://example.com/scalable-services
---
```

**Notes:**
- Requires active user login
- Shows posts from all followed feeds
- Default limit is 2 posts if not specified
- Posts are deduplicated by URL
- Ordered by publication date (newest first)

### Feed Aggregation Commands

#### `agg <time_interval>`
Start automatic feed aggregation at specified intervals.

```bash
# Aggregate every 5 minutes
./gator agg 5m

# Aggregate every 30 seconds
./gator agg 30s

# Aggregate every hour
./gator agg 1h
```

**Time Format Examples:**
- `30s` - 30 seconds
- `1m` - 1 minute  
- `5m` - 5 minutes
- `1h` - 1 hour
- `2h30m` - 2 hours and 30 minutes

**Behavior:**
- Runs continuously until stopped (Ctrl+C)
- Fetches feeds in round-robin fashion
- Updates `last_fetched_at` timestamp
- Stores new posts in database
- Skips duplicate posts (based on URL)
- Handles parsing errors gracefully

## Practical Workflows

### Setting Up a New User

```bash
# 1. Register and login
./gator register myusername

# 2. Add some feeds
./gator addfeed "TechCrunch" "https://techcrunch.com/feed/"
./gator addfeed "Dev.to" "https://dev.to/feed"
./gator addfeed "GitHub Blog" "https://github.blog/feed/"

# 3. Verify feeds are being followed
./gator following

# 4. Start aggregation in the background
./gator agg 10m &

# 5. Browse recent posts
./gator browse 10
```

### Following Existing Feeds

```bash
# 1. Login as existing user
./gator login alice

# 2. See what feeds are available
./gator feeds

# 3. Follow interesting feeds
./gator follow "https://feeds.ycombinator.com/news.rss"
./gator follow "https://go.dev/blog/feed.atom"

# 4. Check your subscriptions
./gator following
```

### Managing Multiple Users

```bash
# Switch between users to manage different feed collections
./gator login work_user
./gator following  # See work-related feeds

./gator login personal_user  
./gator following  # See personal feeds

# Each user has independent feed subscriptions
```

### Running Background Aggregation

```bash
# Start aggregation in background
./gator agg 5m > aggregation.log 2>&1 &

# Check process
ps aux | grep gator

# View logs
tail -f aggregation.log

# Stop background process
pkill gator
```

## Advanced Usage Tips

### Optimal Aggregation Intervals
- **High-frequency feeds** (news sites): 5-15 minutes
- **Blog feeds**: 30 minutes to 1 hour  
- **Low-frequency feeds** (weekly blogs): 1-4 hours
- **Development/testing**: 30 seconds to 1 minute

### Feed URL Discovery
Common RSS feed patterns:
- `/feed`
- `/rss`
- `/feed.xml`
- `/rss.xml`
- `/feed.atom`

### Batch Operations
You can create scripts to manage multiple feeds:

```bash
#!/bin/bash
# setup-tech-feeds.sh

./gator login tech_user

# Add multiple tech feeds
feeds=(
  "Ars Technica,https://feeds.arstechnica.com/arstechnica/index"
  "The Verge,https://www.theverge.com/rss/index.xml"
  "Wired,https://www.wired.com/feed/rss"
)

for feed in "${feeds[@]}"; do
  IFS=',' read -r name url <<< "$feed"
  ./gator addfeed "$name" "$url"
done

echo "Tech feeds setup complete!"
./gator following
```

## Configuration Details

The configuration file (`~/.gatorconfig.json`) contains:

```json
{
  "db_url": "postgres://username:password@localhost:5432/gator?sslmode=disable",
  "current_user_name": "active_username"
}
```

- `db_url`: PostgreSQL connection string
- `current_user_name`: Currently logged-in user (updated by login/register commands)

## Error Handling

Gator handles various error conditions gracefully:

- **Invalid RSS feeds**: Logs parsing errors but continues
- **Network timeouts**: Retries on next aggregation cycle
- **Duplicate posts**: Silently skips (based on URL uniqueness)
- **Database conflicts**: Handles user/feed duplicates appropriately
- **Invalid commands**: Provides usage hints

## Performance Considerations

- **Database size**: Posts accumulate over time; consider periodic cleanup
- **Memory usage**: Minimal during normal operation
- **Network usage**: Depends on feed frequency and content size
- **Concurrent users**: Database handles multiple users efficiently

## Next Steps

- Check [Troubleshooting Guide](TROUBLESHOOTING.md) for common issues
- Explore the source code to understand the internals
- Consider contributing new features or improvements
