# QR Linker - URL Shortener

A secure URL shortener application built with Go, featuring authentication and user management.

## Features

- URL shortening with random hash generation
- Click tracking analytics
- User authentication with sessions
- Secure password hashing with bcrypt
- SQLite database for easy deployment
- Embedded static assets (single binary deployment)
- Responsive web interface

## Quick Start

### Prerequisites

- Go 1.23 or higher
- Air (optional, for hot reload): `go install github.com/air-verse/air@latest`

### Installation

```bash
# Clone the repository
git clone <your-repo-url>
cd qr-linker

# Install dependencies
go mod download

# Configure the application (optional)
cp .env.example .env
# Edit .env with your preferred settings

# Build the application
go build -o qr-linker .
```

### Running the Application

```bash
# Run the application
go run main.go

# Or with hot reload during development (requires air)
air
```

The application will start on `http://localhost:8080`

## User Management

The application includes CLI tools for managing users:

### Add a Single User

```bash
# Interactive mode (prompts for username and password)
go run cmd/adduser/main.go

# Specify username, prompt for password only
go run cmd/adduser/main.go -username john

# Show help
go run cmd/adduser/main.go -help
```

This will interactively prompt for:
- Username (must be unique, 3-50 characters)
- Password (minimum 6 characters, hidden input)
- Password confirmation

### Manage Multiple Users

```bash
# Interactive menu mode
go run cmd/manageusers/main.go

# Quick list of all users
go run cmd/manageusers/main.go -list

# Show help
go run cmd/manageusers/main.go -help
```

Interactive mode provides a menu-driven interface to:
1. List all users
2. Add new users
3. Delete users
4. Change user passwords

### Default Credentials

On first run, the application creates a default admin user:
- Username: `admin`
- Password: `admin123`

**Important:** Change the default password immediately after first login!

## Configuration

The application can be configured using environment variables or a `.env` file.

### Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `BASE_URL` | `http://localhost:8080` | Base URL for displaying short URLs in the interface |
| `PORT` | `8080` | Port the server listens on |
| `DB_PATH` | `urls.db` | Path to the SQLite database file |

### Setup for Production

1. Copy the example environment file:
   ```bash
   cp .env.example .env
   ```

2. Edit `.env` with your production values:
   ```bash
   BASE_URL=https://yourdomain.com
   PORT=3000
   DB_PATH=/var/data/urls.db
   ```

3. The `BASE_URL` is crucial - it's used to display the correct short URLs in the web interface.

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

```bash
# Main application
go run main.go              # Run the web application
air                         # Run with hot reload (requires air)

# User management
go run cmd/adduser/main.go       # Add a single user
go run cmd/manageusers/main.go   # Full user management interface

# Build for production
go build -o qr-linker .     # Build single binary
```

## License

[Your License Here]