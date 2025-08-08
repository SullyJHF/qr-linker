# QR Linker Production Deployment

This guide covers deploying QR Linker using Docker, Docker Compose, and Traefik for production environments.

## Prerequisites

- Docker and Docker Compose installed
- Traefik reverse proxy running on your server
- A domain name pointed to your server

## Quick Start

1. **Clone and configure**:
   ```bash
   git clone <your-repo>
   cd qr-linker
   cp .env.example .env
   ```

2. **Edit environment variables**:
   ```bash
   # Edit .env file with your settings
   BASE_URL=https://links.yourdomain.com
   TRAEFIK_DOMAIN=links.yourdomain.com
   TRAEFIK_CERT_RESOLVER=myresolver
   ```

3. **One-command deploy**:
   ```bash
   ./deploy.sh
   ```

4. **Create admin user**:
   ```bash
   ./deploy.sh adduser
   ```

## Deployment Script Commands

The `deploy.sh` script provides several useful commands:

```bash
./deploy.sh deploy         # Build and deploy to production (default)
./deploy.sh adduser        # Add a new user interactively
./deploy.sh manage-users   # Open user management interface
./deploy.sh logs           # Show application logs
./deploy.sh stop           # Stop application
./deploy.sh restart        # Restart application
./deploy.sh status         # Show container status
./deploy.sh health         # Check application health
./deploy.sh help           # Show help message
```

## Environment Variables

| Variable | Description | Default | Production Example |
|----------|-------------|---------|-------------------|
| `BASE_URL` | Public URL of your application | `http://localhost:8080` | `https://links.yourdomain.com` |
| `TRAEFIK_DOMAIN` | Domain for Traefik routing | `links.yourdomain.com` | `links.yourdomain.com` |
| `TRAEFIK_CERT_RESOLVER` | Traefik certificate resolver name | `myresolver` | `myresolver` |
| `PORT` | Internal port (don't change for Docker) | `8080` | `8080` |
| `DB_PATH_DEV` | Development database file path | `urls-dev.db` | `urls-dev.db` |
| `DB_PATH` | Production database file path | `/app/data/urls.db` | `/app/data/urls.db` |

### Database Configuration

The application supports separate database files for development and production:

- **Development** (`air`, `go run`): Uses `DB_PATH_DEV` (e.g., `urls-dev.db`)
- **Production** (Docker): Uses `DB_PATH` (e.g., `/app/data/urls.db`)

This allows you to use the same `.env` file for both environments while keeping separate databases.

## SSL/TLS Configuration

SSL/TLS is handled automatically by Traefik using the configured certificate resolver. Make sure your Traefik instance is properly configured with Let's Encrypt or your preferred certificate provider.

Example Traefik configuration for Let's Encrypt:
```yaml
certificatesResolvers:
  myresolver:
    acme:
      email: your-email@example.com
      storage: acme.json
      httpChallenge:
        entryPoint: web
```

## User Management

The CLI tools are built into the Docker container, so you can manage users directly:

### Add a new user:
```bash
docker-compose exec qr-linker ./adduser
```

### Manage users interactively:
```bash
# List, add, delete, or change passwords
docker-compose exec qr-linker ./manageusers
```

### Examples:
```bash
# Add user with prompts
docker-compose exec qr-linker ./adduser

# Interactive management menu
docker-compose exec qr-linker ./manageusers

# View help
docker-compose exec qr-linker ./adduser -h
docker-compose exec qr-linker ./manageusers -h
```

## Monitoring and Logs

### View logs:
```bash
docker-compose logs -f qr-linker
```

### Health checks:
The application includes health checks accessible at:
- Internal: `http://localhost:8080/login`
- External: `https://your-domain.com/login`

### Backup database:
```bash
# Create backup
docker-compose exec qr-linker cp /app/data/urls.db /app/data/urls.db.backup

# Copy to host
docker cp qr-linker:/app/data/urls.db.backup ./backup-$(date +%Y%m%d-%H%M%S).db
```

## Security Considerations

1. **Change default session secret** in production
2. **Use HTTPS** in production (configure BASE_URL accordingly)
3. **Regular database backups**
4. **Keep Docker images updated**
5. **Use strong passwords** for admin users
6. **Consider rate limiting** at the reverse proxy level

## Scaling

For high-traffic deployments:

1. **Use external database**: Consider PostgreSQL or MySQL instead of SQLite
2. **Load balancing**: Run multiple instances behind a load balancer
3. **Caching**: Add Redis for session storage
4. **CDN**: Use a CDN for static assets

## Troubleshooting

### Container won't start:
```bash
docker-compose logs qr-linker
```

### Database permission issues:
```bash
docker-compose exec qr-linker ls -la /app/data/
```

### SSL certificate issues:
```bash
docker-compose logs nginx
```

### Reset everything:
```bash
docker-compose down -v  # WARNING: This deletes all data!
docker-compose up -d
```