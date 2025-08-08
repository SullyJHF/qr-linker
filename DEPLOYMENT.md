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

## GitHub Actions CI/CD

The repository includes a GitHub Actions workflow for automatic deployment to your VPS when you push to the main branch.

### Setup GitHub Secrets

Configure these secrets in your GitHub repository settings (Settings → Secrets and variables → Actions):

| Secret Name | Description | Example Value |
|-------------|-------------|---------------|
| `VPS_HOST` | IP address or hostname of your VPS | `123.456.789.012` or `server.yourdomain.com` |
| `VPS_USERNAME` | SSH username for your VPS | `root` or `deploy` |
| `VPS_SSH_KEY` | Private SSH key for VPS access | Contents of your `~/.ssh/id_rsa` file |
| `VPS_PORT` | SSH port (optional, defaults to 22) | `22` |
| `DOMAIN` | Your application domain for health checks | `links.yourdomain.com` |

### SSH Key Setup

1. **Generate SSH key pair** (if you don't have one):
   ```bash
   ssh-keygen -t rsa -b 4096 -C "github-actions@yourdomain.com"
   ```

2. **Copy public key to your VPS**:
   ```bash
   ssh-copy-id -i ~/.ssh/id_rsa.pub username@your-vps-ip
   ```

3. **Add private key to GitHub secrets**:
   ```bash
   # Copy the ENTIRE private key including headers
   cat ~/.ssh/id_rsa
   ```
   Copy the complete output and paste it as the `VPS_SSH_KEY` secret.

### VPS Preparation

On your VPS, ensure:

1. **Repository is cloned**:
   ```bash
   cd ~
   git clone https://github.com/yourusername/qr-linker.git
   cd qr-linker
   ```

2. **Environment configured**:
   ```bash
   cp .env.example .env
   # Edit .env with your production settings
   ```

3. **Deploy script is executable**:
   ```bash
   chmod +x deploy.sh
   ```

4. **Docker and Traefik are running**:
   ```bash
   # Ensure Docker is installed and running
   docker --version
   docker-compose --version
   
   # Ensure Traefik network exists
   docker network ls | grep traefik
   ```

### Workflow Behavior

The GitHub Actions workflow:

1. **Triggers** on push to `main` branch or manual workflow dispatch
2. **Tests** the Go application with `go test`
3. **Deploys** by SSH'ing to your VPS and running `./deploy.sh deploy`
4. **Health checks** your deployed application
5. **Notifies** success or failure

### Manual Deployment

You can also trigger deployment manually:
- Go to your GitHub repository
- Click "Actions" tab
- Select "Deploy to VPS" workflow
- Click "Run workflow" → "Run workflow"

### Workflow Logs

View deployment logs in GitHub:
- Repository → Actions → Click on a workflow run
- Expand steps to see detailed output
- Check "Deploy to VPS" step for deployment details

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

### GitHub Actions deployment fails:

1. **SSH connection issues**:
   - Verify `VPS_HOST`, `VPS_USERNAME`, `VPS_PORT` secrets
   - Ensure SSH key has correct permissions
   - Test SSH connection manually: `ssh username@vps-host`

2. **Deployment script fails**:
   - Check VPS has Docker and docker-compose installed
   - Verify Traefik is running and network exists
   - Ensure .env file is properly configured on VPS

3. **Health check fails**:
   - Verify `DOMAIN` secret matches your actual domain
   - Check if SSL certificate is properly configured
   - Ensure application is accessible at `/login` endpoint

### Reset everything:
```bash
docker-compose down -v  # WARNING: This deletes all data!
docker-compose up -d
```