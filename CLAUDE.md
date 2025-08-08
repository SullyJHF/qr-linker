# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Commands

### Development
- `air` - Run development server with hot reload (watches .go, .html, .css files)
- `go run main.go` - Run application without hot reload
- Server runs on `http://localhost:8080`
- Uses `DB_PATH_DEV` environment variable for development database (default: `urls-dev.db`)

### Build
- `go build -o qr-linker .` - Build production binary with embedded assets
- `go build -o ./tmp/main .` - Build to tmp directory (used by Air)

### Dependencies
- `go mod tidy` - Update and clean dependencies
- `go get <package>` - Add new dependencies

## Architecture

### Core Structure
The application is a URL shortener with three main layers:

1. **HTTP Layer** (`main.go`): Handles routing and request processing
   - Routes: `/` (home), `/shorten` (POST), `/{hash}` (redirect)
   - Uses query parameters for success/error messages after form submission
   - Embedded templates and static files using Go's `embed` directive

2. **Database Layer** (`database/db.go`): SQLite persistence
   - Connection pooling and table initialization on startup
   - URLs table with short_hash index for fast lookups
   - Tracks click counts for analytics

3. **Utils Layer** (`utils/hash.go`): Hash generation
   - Generates 6-character URL-safe base64 hashes
   - Collision detection with automatic retry and length expansion
   - Uses crypto/rand for secure random generation

### Key Design Decisions
- **Embedded Assets**: All templates and CSS are embedded in the binary for single-file deployment
- **POST-Redirect-GET Pattern**: Form submissions redirect to `/` with query parameters to prevent duplicate submissions
- **Database File**: SQLite database (`urls.db`) created automatically on first run
- **No External Frontend Dependencies**: Pure HTML/CSS with minimal JavaScript for clipboard functionality

### Data Flow
1. User submits URL via form → POST to `/shorten`
2. Generate unique hash, save to database
3. Redirect to `/?success={hash}` 
4. Homepage displays success message and refreshed URL list
5. Short URL access (`/{hash}`) → Database lookup → 301 redirect to original URL

## Database Schema

```sql
urls table:
- id (INTEGER PRIMARY KEY AUTOINCREMENT)
- full_url (TEXT NOT NULL)
- short_hash (TEXT NOT NULL UNIQUE)
- created_at (DATETIME DEFAULT CURRENT_TIMESTAMP)
- clicks (INTEGER DEFAULT 0)
```

## Development Notes

- Air configuration in `.air.toml` - modify for custom build commands or watch patterns
- Database file `urls.db` is created in project root - add to .gitignore for production
- Template changes require server restart unless using Air
- URL validation adds https:// prefix if protocol is missing