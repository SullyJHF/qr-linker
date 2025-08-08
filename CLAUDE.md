# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Commands

### Development
- `air` - Run development server with hot reload (watches .go, .html, .css files)
- `go run main.go` - Run application without hot reload
- `go run cmd/adduser/main.go` - Add user to development database
- `go run cmd/manageusers/main.go` - Manage users in development database
- Server runs on `http://localhost:8080`
- Uses `DB_PATH_DEV` environment variable for development database (default: `urls-dev.db`)

### Build
- `go build -o qr-linker .` - Build production binary with embedded assets
- `go build -o ./tmp/main .` - Build to tmp directory (used by Air)

### Dependencies
- `go mod tidy` - Update and clean dependencies
- `go get <package>` - Add new dependencies

### Docker Deployment
- `./deploy.sh` - Deploy to production with Traefik (uses docker-compose.yml)
- `./deploy.sh local` - Deploy locally for testing (uses docker-compose.local.yml)
- `./deploy.sh adduser` - Add user to production database via Docker
- `./deploy.sh manage-users` - Manage users in production database via Docker
- `./deploy.sh logs` - View application logs
- `./deploy.sh health` - Check application health
- `./deploy.sh stop` - Stop containers
- `./deploy.sh restart` - Restart containers
- `./deploy.sh status` - Show container status

### Docker Testing
- Local testing: `./deploy.sh local` runs on `http://localhost:8080` without Traefik
- Production testing: `./deploy.sh` requires Traefik configuration and domain setup
- Both deployments use persistent Docker volumes for database storage

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
- **Dual Database Support**: Uses `DB_PATH_DEV` for development, `DB_PATH` for production
- **Modal UI**: Edit URLs directly from the main interface without page navigation
- **QR Code Generation**: Built-in QR codes for all shortened URLs
- **Docker Integration**: Full containerization with built-in CLI tools and persistent storage
- **Authentication**: Session-based auth with bcrypt password hashing
- **No External Frontend Dependencies**: Pure HTML/CSS with minimal JavaScript

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

users table:
- id (INTEGER PRIMARY KEY AUTOINCREMENT)
- username (TEXT NOT NULL UNIQUE)
- password_hash (TEXT NOT NULL)
- created_at (DATETIME DEFAULT CURRENT_TIMESTAMP)
```

### Database Files
- **Development**: `urls-dev.db` (used by `air`, `go run`)
- **Production**: `/app/data/urls.db` (used by Docker deployments)
- **CLI Tools**: Automatically use correct database based on environment variables

## Development Notes

- Air configuration in `.air.toml` - modify for custom build commands or watch patterns
- Development database `urls-dev.db` is separate from production database
- Template changes require server restart unless using Air
- URL validation adds https:// prefix if protocol is missing
- Modal UI allows editing URLs without page reload
- QR codes are generated server-side and cached
- CLI tools read same environment variables as main application for database consistency

## Docker Notes

- Multi-stage Alpine-based build for minimal image size
- CLI tools (`adduser`, `manageusers`) built into container for user management
- Database persists in Docker volumes across container restarts
- Environment variables passed through docker-compose configuration
- Health checks monitor application availability
- Non-root user for security compliance
- Traefik integration for automatic SSL/TLS certificates in production