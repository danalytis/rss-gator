# Troubleshooting Guide

This guide covers common issues you might encounter while using Gator RSS Aggregator and their solutions.

## Database Issues

### "could not connect to database"

**Symptoms:**
```
could not connect to database
```

**Possible Causes & Solutions:**

1. **PostgreSQL not running**
   ```bash
   # Check if PostgreSQL is running
   # macOS
   brew services list | grep postgresql
   
   # Ubuntu/Debian
   sudo systemctl status postgresql
   
   # Windows
   net start postgresql-x64-15  # version may vary
   ```

2. **Incorrect database URL**
   - Check `~/.gatorconfig.json` for typos
   - Verify username, password, host, and port
   - Test connection manually:
   ```bash
   psql "postgres://username:password@localhost:5432/gator"
   ```

3. **Database doesn't exist**
   ```bash
   # Create the database
   psql -U postgres -c "CREATE DATABASE gator;"
   ```

4. **Permission issues**
   ```bash
   # Grant permissions to your user
   psql -U postgres -c "GRANT ALL PRIVILEGES ON DATABASE gator TO your_username;"
   ```

### "relation does not exist" or "table not found"

**Symptoms:**
```
pq: relation "users" does not exist
```

**Solution:**
Run database migrations to create the required tables:

```bash
cd sql/schema
goose postgres "your_database_url" up
```

If migrations fail, check that:
- You're in the correct directory (`sql/schema`)
- Database URL is correct
- PostgreSQL user has CREATE permissions

### "duplicate key value violates unique constraint"

**Symptoms:**
```
pq: duplicate key value violates unique constraint "users_name_key"
```

**Solutions:**

1. **For users**: Username already exists
   ```bash
   # Try a different username
   ./gator register different_username
   
   # Or login as existing user
   ./gator login existing_username
   ```

2. **For feeds**: Feed URL already exists
   ```bash
   # Follow the existing feed instead
   ./gator follow "https://existing-feed-url.com/rss"
   ```

## Configuration Issues

### "error: could not read config file"

**Symptoms:**
- Commands fail with configuration-related errors
- No config file found

**Solutions:**

1. **Create config file**
   ```bash
   # Linux/macOS
   cat > ~/.gatorconfig.json << EOF
   {
     "db_url": "postgres://postgres:@localhost:5432/gator?sslmode=disable",
     "current_user_name": ""
   }
   EOF
   
   # Windows PowerShell
   @"
   {
     "db_url": "postgres://postgres:@localhost:5432/gator?sslmode=disable",
     "current_user_name": ""
   }
   "@ | Out-File -FilePath "$env:USERPROFILE\.gatorconfig.json" -Encoding UTF8
   ```

2. **Check file location**
   ```bash
   # Should be in your home directory
   ls -la ~/.gatorconfig.json
   
   # Windows
   dir %USERPROFILE%\.gatorconfig.json
   ```

3. **Fix JSON syntax**
   - Ensure valid JSON format
   - Check for missing commas or quotes
   - Use a JSON validator online

### "user not found" after login

**Symptoms:**
```
user 'username' not found
```

**Solutions:**

1. **Register the user first**
   ```bash
   ./gator register username
   ```

2. **Check existing users**
   ```bash
   ./gator users
   ```

3. **Check current user in config**
   ```bash
   cat ~/.gatorconfig.json
   # Ensure current_user_name matches an existing user
   ```

## Feed and Network Issues

### "Feed parsing errors" or XML parsing failures

**Symptoms:**
- Feeds fail to fetch during aggregation
- "error parsing feeds" messages
- XML unmarshaling errors

**Solutions:**

1. **Verify feed URL manually**
   ```bash
   curl -I "https://your-feed-url.com/rss"
   # Should return HTTP 200 and Content-Type: application/rss+xml or similar
   ```

2. **Test feed content**
   ```bash
   curl "https://your-feed-url.com/rss" | head -20
   # Should show XML content starting with <?xml or <rss>
   ```

3. **Common feed issues:**
   - **Invalid XML**: Some feeds have malformed XML
   - **Wrong Content-Type**: Server returns HTML instead of XML
   - **Authentication required**: Feed requires login
   - **Geographic restrictions**: Feed blocked in your region

4. **Alternative feed URLs to try:**
   - `/feed` instead of `/rss`
   - `/feed.xml` instead of `/feed`
   - `/atom.xml` for Atom feeds
   - Check website for official RSS links

### Network timeout or connection issues

**Symptoms:**
- Feeds fail to fetch
- "connection timeout" errors
- "no such host" errors

**Solutions:**

1. **Check internet connectivity**
   ```bash
   ping google.com
   ```

2. **Test specific feed URL**
   ```bash
   curl -v "https://problematic-feed-url.com/rss"
   ```

3. **Check firewall/proxy settings**
   - Corporate networks may block RSS feeds
   - Try from a different network

4. **Feed server issues**
   - Try again later
   - Check if the website is down

## Build and Dependency Issues

### "sqlc command not found"

**Symptoms:**
```
bash: sqlc: command not found
```

**Solutions:**

1. **Install sqlc**
   ```bash
   go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest
   ```

2. **Check Go bin in PATH**
   ```bash
   echo $PATH | grep -o '[^:]*go[^:]*'
   
   # Add to PATH if missing (Linux/macOS)
   export PATH=$PATH:$(go env GOPATH)/bin
   
   # Make permanent by adding to ~/.bashrc or ~/.zshrc
   echo 'export PATH=$PATH:$(go env GOPATH)/bin' >> ~/.bashrc
   ```

3. **Windows PATH issues**
   - Add `%USERPROFILE%\go\bin` to system PATH
   - Restart Command Prompt/PowerShell

### "goose command not found"

**Symptoms:**
```
bash: goose: command not found
```

**Solutions:**

1. **Install goose**
   ```bash
   go install github.com/pressly/goose/v3/cmd/goose@latest
   ```

2. **Same PATH issues as sqlc above**

### Go build failures

**Symptoms:**
```
go build: module not found
cannot find package
```

**Solutions:**

1. **Initialize/update modules**
   ```bash
   go mod init gator  # if go.mod missing
   go mod tidy
   ```

2. **Clear module cache**
   ```bash
   go clean -modcache
   go mod download
   ```

3. **Check Go version**
   ```bash
   go version
   # Should be 1.19 or higher
   ```

## Runtime Issues

### Application crashes during aggregation

**Symptoms:**
- `./gator agg` stops unexpectedly
- Memory-related crashes
- Database connection errors during long runs

**Solutions:**

1. **Check system resources**
   ```bash
   # Monitor memory and CPU usage
   top
   htop  # if available
   ```

2. **Run with more verbose logging**
   ```bash
   ./gator agg 5m 2>&1 | tee aggregation.log
   ```

3. **Check for specific feeds causing issues**
   - Run aggregation with shorter intervals
   - Monitor which feeds are processed before crashes
   - Test problematic feeds manually

4. **Database connection pooling**
   - Restart PostgreSQL service
   - Check PostgreSQL logs for connection issues

### Posts not appearing in browse

**Symptoms:**
- `./gator browse` shows no posts
- Recent posts missing

**Solutions:**

1. **Verify feed subscriptions**
   ```bash
   ./gator following
   # Ensure you're following feeds
   ```

2. **Check if aggregation is running**
   ```bash
   # Run aggregation manually to test
   ./gator agg 30s
   # Let it run for a few cycles, then Ctrl+C
   ```

3. **Verify feeds have recent content**
   ```bash
   # Check feed manually
   curl "https://your-feed-url.com/rss" | grep -i pubDate
   ```

4. **Check database directly**
   ```bash
   psql -d gator -c "SELECT COUNT(*) FROM posts;"
   psql -d gator -c "SELECT title, published_at FROM posts ORDER BY published_at DESC LIMIT 5;"
   ```

### Permission denied errors

**Symptoms:**
```
permission denied: ~/.gatorconfig.json
./gator: permission denied
```

**Solutions:**

1. **Fix config file permissions**
   ```bash
   chmod 644 ~/.gatorconfig.json
   ```

2. **Fix executable permissions**
   ```bash
   chmod +x ./gator
   ```

3. **Run from correct directory**
   ```bash
   # Ensure you're in the project directory
   pwd
   ls -la gator
   ```

## Performance Issues

### Slow aggregation

**Symptoms:**
- Long delays between feed fetches
- High memory/CPU usage

**Solutions:**

1. **Increase aggregation interval**
   ```bash
   # Use longer intervals for less frequent updates
   ./gator agg 30m  # instead of 1m
   ```

2. **Check feed response times**
   ```bash
   time curl "https://slow-feed-url.com/rss"
   ```

3. **Monitor resource usage**
   ```bash
   # While aggregation is running
   ps aux | grep gator
   ```

### Database growing too large

**Symptoms:**
- Large disk usage
- Slow queries

**Solutions:**

1. **Check database size**
   ```bash
   psql -d gator -c "\dt+"  # Table sizes
   psql -d gator -c "SELECT COUNT(*) FROM posts;"
   ```

2. **Clean old posts** (manual cleanup)
   ```sql
   -- Delete posts older than 30 days
   DELETE FROM posts WHERE created_at < NOW() - INTERVAL '30 days';
   ```

3. **Implement retention policy** (requires code changes)

## Getting Additional Help

### Enable Debug Logging

Add logging to help diagnose issues:

```bash
# Run with output redirection to capture all logs
./gator agg 1m > debug.log 2>&1

# Monitor logs in real-time
tail -f debug.log
```

### Check System Resources

```bash
# Check disk space
df -h

# Check memory usage  
free -h

# Check PostgreSQL processes
ps aux | grep postgres
```

### Collect System Information

When reporting issues, include:

1. **Operating system and version**
   ```bash
   uname -a
   cat /etc/os-release  # Linux
   ```

2. **Software versions**
   ```bash
   go version
   psql --version
   sqlc version
   goose -version
   ```

3. **Configuration (without sensitive data)**
   ```bash
   # Remove password from output
   cat ~/.gatorconfig.json | sed 's/:.*@/:***@/'
   ```

4. **Error logs**
   - Include full error messages
   - Provide steps to reproduce

### Community Resources

- Check the project's issue tracker for similar problems
- Search for error messages in community forums
- Consider reaching out on relevant Go or PostgreSQL communities

## Still Having Issues?

If you're still experiencing problems:

1. **Double-check the [Installation Guide](INSTALLATION.md)**
2. **Review the [Usage Guide](USAGE.md)** for proper command usage
3. **Create a minimal reproduction case**
4. **File an issue with detailed information** including:
   - Operating system and version
   - Software versions
   - Full error messages
   - Steps to reproduce
   - Configuration (sanitized)
