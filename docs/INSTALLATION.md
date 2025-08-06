# Installation Guide

This guide will walk you through installing and setting up the Gator RSS Aggregator.

## Prerequisites

Ensure you have administrative privileges on your system for installing dependencies.

## Step 1: Install System Dependencies

### Go Programming Language

#### macOS
```bash
# Using Homebrew
brew install go

# Verify installation
go version
```

#### Ubuntu/Debian
```bash
# Update package list
sudo apt update

# Install Go
sudo apt install golang-go

# Verify installation
go version
```

#### Windows
1. Download the installer from [https://golang.org/dl/](https://golang.org/dl/)
2. Run the installer and follow the setup wizard
3. Open Command Prompt and verify: `go version`

#### Manual Installation (All Platforms)
1. Download from [https://golang.org/dl/](https://golang.org/dl/)
2. Extract to `/usr/local` (Linux/macOS) or `C:\Go` (Windows)
3. Add to PATH environment variable
4. Restart terminal and verify: `go version`

### PostgreSQL Database

#### macOS
```bash
# Using Homebrew
brew install postgresql

# Start PostgreSQL service
brew services start postgresql

# Connect to verify installation
psql postgres
```

#### Ubuntu/Debian
```bash
# Update package list
sudo apt update

# Install PostgreSQL
sudo apt install postgresql postgresql-contrib

# Start and enable service
sudo systemctl start postgresql
sudo systemctl enable postgresql

# Switch to postgres user
sudo -u postgres psql
```

#### Windows
1. Download installer from [https://www.postgresql.org/download/windows/](https://www.postgresql.org/download/windows/)
2. Run installer and follow setup wizard
3. Remember the password you set for the `postgres` user
4. Add PostgreSQL bin directory to PATH (usually `C:\Program Files\PostgreSQL\15\bin`)

### sqlc (SQL Code Generator)

```bash
# Install sqlc
go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest

# Verify installation
sqlc version
```

### goose (Database Migration Tool)

```bash
# Install goose
go install github.com/pressly/goose/v3/cmd/goose@latest

# Verify installation
goose -version
```

**Note**: Ensure your Go bin directory is in your PATH. Usually `~/go/bin` on Linux/macOS or `%USERPROFILE%\go\bin` on Windows.

## Step 2: Download and Build Gator

### Clone Repository
```bash
# Clone the repository (replace with actual URL)
git clone https://github.com/danalytis/rss-gator.git
cd gator
```

### Install Go Dependencies
```bash
# Download and install Go module dependencies
go mod tidy
```

### Generate Database Code
```bash
# Generate Go code from SQL files using sqlc
sqlc generate
```

### Build Application
```bash
# Build the gator executable
go build -o gator

# On Windows
go build -o gator.exe
```

## Step 3: Database Setup

### Create Database and User

#### Connect to PostgreSQL
```bash
# Connect as postgres superuser
psql -U postgres

# On Windows, you might need:
psql -U postgres -h localhost
```

#### Create Database
```sql
-- Create the gator database
CREATE DATABASE gator;

-- Create a dedicated user (optional but recommended)
CREATE USER gator_user WITH PASSWORD 'secure_password';

-- Grant privileges
GRANT ALL PRIVILEGES ON DATABASE gator TO gator_user;

-- Exit PostgreSQL
\q
```

### Run Database Migrations

Navigate to the schema directory and run migrations:

```bash
# Navigate to migrations directory
cd sql/schema

# Run all migrations (using postgres superuser)
goose postgres "postgres://postgres:@localhost:5432/gator?sslmode=disable" up

# Or using custom user (replace with your password)
goose postgres "postgres://gator_user:secure_password@localhost:5432/gator?sslmode=disable" up

# Return to project root
cd ../..
```

### Verify Database Setup
```bash
# Connect to verify tables were created
psql -U postgres -d gator

# List tables
\dt

# You should see: users, feeds, feed_follows, posts

# Exit
\q
```

## Step 4: Configuration

### Create Configuration File

Create the configuration file in your home directory:

#### Linux/macOS
```bash
# Create configuration file
cat > ~/.gatorconfig.json << EOF
{
  "db_url": "postgres://postgres:@localhost:5432/gator?sslmode=disable",
  "current_user_name": ""
}
EOF
```

#### Windows (PowerShell)
```powershell
# Create configuration file
@"
{
  "db_url": "postgres://postgres:@localhost:5432/gator?sslmode=disable",
  "current_user_name": ""
}
"@ | Out-File -FilePath "$env:USERPROFILE\.gatorconfig.json" -Encoding UTF8
```

#### Windows (Command Prompt)
```cmd
echo { > %USERPROFILE%\.gatorconfig.json
echo   "db_url": "postgres://postgres:@localhost:5432/gator?sslmode=disable", >> %USERPROFILE%\.gatorconfig.json
echo   "current_user_name": "" >> %USERPROFILE%\.gatorconfig.json
echo } >> %USERPROFILE%\.gatorconfig.json
```

### Update Database URL

If you created a custom user or have different connection details, update the `db_url` in your configuration file:

```json
{
  "db_url": "postgres://gator_user:secure_password@localhost:5432/gator?sslmode=disable",
  "current_user_name": ""
}
```

**Security Note**: The configuration file contains database credentials. Ensure it has appropriate permissions:

```bash
# Linux/macOS: Restrict access to owner only
chmod 600 ~/.gatorconfig.json
```

## Step 5: Verify Installation

### Test the Application
```bash
# Test basic functionality
./gator register testuser

# You should see: "successfully registered user: testuser"

# Clean up test user
./gator reset
```

### Check Dependencies
```bash
# Verify all tools are accessible
go version
psql --version
sqlc version
goose -version
```

## Troubleshooting Installation

### Go Not in PATH
Add Go's bin directory to your PATH:

```bash
# Linux/macOS (add to ~/.bashrc or ~/.zshrc)
export PATH=$PATH:$(go env GOPATH)/bin

# Windows: Add to system PATH environment variable
# %USERPROFILE%\go\bin
```

### PostgreSQL Connection Issues
- Ensure PostgreSQL service is running
- Check if PostgreSQL is listening on port 5432: `netstat -an | grep 5432`
- Verify authentication in `pg_hba.conf` if using custom users

### sqlc/goose Not Found
Ensure Go's bin directory is in PATH and reinstall if necessary:
```bash
go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest
go install github.com/pressly/goose/v3/cmd/goose@latest
```

### Database Migration Failures
- Ensure database exists before running migrations
- Check PostgreSQL logs for detailed error messages
- Verify user has sufficient privileges

## Next Steps

Once installation is complete, proceed to the [Usage Guide](USAGE.md) to learn how to use Gator.
