# QR Linker - URL Shortener

A secure URL shortener application built with Go, featuring authentication, user management, QR code generation, and Docker deployment.

## Features

- URL shortening with random hash generation
- QR code generation for shortened URLs
- Click tracking analytics
- User authentication with sessions
- Secure password hashing with bcrypt
- SQLite database for easy deployment
- Embedded static assets (single binary deployment)
- Responsive web interface with modal editing
- Docker deployment with Traefik support
- Built-in CLI tools for user management

## Quick Start

### Prerequisites

**For Development:**
- Go 1.23 or higher
- Air (optional, for hot reload): `go install github.com/air-verse/air@latest`

**For Production:**
- Docker and Docker Compose
- Traefik (for production deployment)

### Development Setup

```bash
# Clone the repository
git clone <your-repo-url>
cd qr-linker

# Install dependencies
go mod download

# Configure the application (optional)
cp .env.example .env
# Edit .env with your preferred settings

# Run with hot reload
air

# Or run directly
go run main.go
```

The application will start on `http://localhost:8080`

### Production Deployment (Docker)

```bash
# Clone and configure
git clone <your-repo-url>
cd qr-linker
cp .env.example .env

# Edit .env with your production settings
# Required: BASE_URL, TRAEFIK_DOMAIN, TRAEFIK_CERT_RESOLVER

# Deploy to production
./deploy.sh

# Create admin user
./deploy.sh adduser
```

### Local Docker Testing

```bash
# Test locally without Traefik
./deploy.sh local

# Manage users
./deploy.sh adduser
./deploy.sh manage-users

# View logs
./deploy.sh logs

# Stop containers
./deploy.sh stop
```

## User Management

### Development (Direct CLI)

```bash
# Add a single user
go run cmd/adduser/main.go

# Manage all users (interactive menu)
go run cmd/manageusers/main.go

# Quick list of all users
go run cmd/manageusers/main.go -list
```

### Production/Docker

```bash
# Add a single user
./deploy.sh adduser

# Manage all users (interactive menu)  
./deploy.sh manage-users

# For local Docker deployment
./deploy.sh local     # Start containers
./deploy.sh adduser   # Add user to local deployment
```

The CLI tools provide:
- **Add User**: Interactive prompts for username and password
- **Manage Users**: Menu-driven interface to list, add, delete users and change passwords
- **Automatic validation**: Username 3-50 chars, password minimum 6 chars
- **Database consistency**: All tools use the same database as the web application

### Default Credentials

On first run, the application creates a default admin user:
- Username: `admin`
- Password: `admin123`

**Important:** Change the default password immediately after first login!

## Configuration

### Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `BASE_URL` | `http://localhost:8080` | Public URL for your application |
| `PORT` | `8080` | Port the server listens on |
| `DB_PATH_DEV` | `urls-dev.db` | Development database file path |
| `DB_PATH` | `urls.db` | Production database file path |
| `TRAEFIK_DOMAIN` | - | Domain for Traefik routing (production only) |
| `TRAEFIK_CERT_RESOLVER` | - | Traefik certificate resolver (production only) |

### Development Setup

```bash
cp .env.example .env
# Edit .env with your settings
# Uses DB_PATH_DEV for separate development database
```

### Production Setup

```bash
cp .env.example .env
# Required for production:
BASE_URL=https://links.yourdomain.com
TRAEFIK_DOMAIN=links.yourdomain.com
TRAEFIK_CERT_RESOLVER=myresolver
```

The application automatically chooses:
- **Development**: Uses `DB_PATH_DEV` when running with `air` or `go run`
- **Production**: Uses `DB_PATH` when running in Docker

## Database

The application uses SQLite and stores data in the configured database path (default: `urls.db`). The database is created automatically on first run.

### Database Schema

**urls table:**
- `id` - Primary key
- `full_url` - Original URL
- `short_hash` - Generated short hash
- `created_at` - Timestamp
- `clicks` - Click counter

**users table:**
- `id` - Primary key
- `username` - Unique username
- `password_hash` - Bcrypt hashed password
- `created_at` - Timestamp

## Security

- All routes except `/login` require authentication
- Passwords are hashed using bcrypt
- Sessions expire after 7 days
- HttpOnly cookies for session management
- CSRF protection through SameSite cookies

## Development

### Project Structure

```
qr-linker/
├── main.go                 # Main application
├── cmd/
│   ├── adduser/           # CLI for adding single users
│   └── manageusers/       # CLI for user management
├── auth/                  # Authentication middleware
├── database/              # Database operations
├── utils/                 # Utility functions (hash generation)
├── templates/             # HTML templates
├── static/                # CSS files
└── urls.db               # SQLite database
```

### Available Commands

**Development:**
```bash
go run main.go                    # Run web application
air                              # Run with hot reload
go run cmd/adduser/main.go       # Add user (development DB)
go run cmd/manageusers/main.go   # Manage users (development DB)
go build -o qr-linker .          # Build single binary
```

**Production/Docker:**
```bash
./deploy.sh                      # Deploy to production
./deploy.sh local                # Deploy locally for testing
./deploy.sh adduser              # Add user to production DB
./deploy.sh manage-users         # Manage production users
./deploy.sh logs                 # View application logs
./deploy.sh health               # Check application health
./deploy.sh stop                 # Stop containers
./deploy.sh restart              # Restart containers
```

## Docker Deployment

### Production with Traefik

The application includes full Docker support with Traefik integration for production deployments.

**Prerequisites:**
- Docker and Docker Compose
- Traefik running with external network named `traefik`

**Quick Deploy:**
```bash
cp .env.example .env
# Edit .env with your domain and Traefik settings
./deploy.sh
./deploy.sh adduser
```

### Local Docker Testing

Test the Docker build locally without Traefik:

```bash
./deploy.sh local
```

This runs on `http://localhost:8080` with full database persistence.

### Docker Features

- **Multi-stage build**: Optimized Alpine-based production image
- **Built-in CLI tools**: User management tools included in container
- **Database persistence**: SQLite database stored in Docker volumes
- **Health checks**: Automatic container health monitoring
- **Traefik integration**: Automatic SSL certificates and routing
- **Non-root user**: Containers run as unprivileged user for security

### Deployment Script Commands

```bash
./deploy.sh deploy         # Deploy to production (default)
./deploy.sh local          # Deploy locally for testing
./deploy.sh adduser        # Add user interactively
./deploy.sh manage-users   # User management interface
./deploy.sh logs           # Show application logs
./deploy.sh health         # Check application health
./deploy.sh status         # Show container status
./deploy.sh stop           # Stop containers
./deploy.sh restart        # Restart containers
```

See [DEPLOYMENT.md](DEPLOYMENT.md) for detailed production deployment instructions.

## License

[Your License Here]