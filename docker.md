# Docker Setup for Go Projects

This guide covers Docker deployment patterns for Go applications, with emphasis on best practices for building, deploying, and hosting container images.

## Overview

Modern Go applications benefit from containerization for consistent deployment across environments. This guide covers:

- Multi-stage Docker builds for minimal image sizes
- CGO vs pure Go considerations
- GitHub Container Registry (GHCR) integration
- Local development with docker-compose
- Security and optimization best practices

## When to Use Docker

**Recommended for:**

- Web applications and HTTP services
- Long-running daemons and background services
- Applications with persistent data (databases, file storage)
- Projects requiring consistent deployment environments
- Multi-service applications (web + database + cache)

**Not necessary for:**

- CLI tools (distribute as standalone binaries)
- Scripts and one-off utilities
- Development-only tools

## CGO Considerations

### Pure Go (No CGO)

**Advantages:**

- Smaller final images (can use `scratch` or `alpine`)
- Faster builds (no C compiler needed)
- True cross-compilation
- Simpler Dockerfile

**When to use:**

- Default choice for new projects
- When using pure Go libraries

**SQLite without CGO:**

```go
import _ "modernc.org/sqlite"  // Pure Go SQLite driver
```

### CGO-Enabled Builds

**When required:**

- Using `github.com/mattn/go-sqlite3` (C-based SQLite)
- Integrating with C libraries
- Performance-critical native code

**Trade-offs:**

- Requires C compiler in build stage
- Larger build images
- More complex Dockerfile
- Cannot use `scratch` base image

## Dockerfile Patterns

### Multi-Stage Build (Pure Go)

See `templates/docker/Dockerfile-go-pure` for a complete example.

**Key features:**

- Builder stage with Go toolchain
- Runtime stage with minimal base image (`scratch` or `alpine`)
- Final image ~10-20MB for simple applications
- Static binary compilation

**Example structure:**

```dockerfile
# Stage 1: Build
FROM golang:1.25-alpine AS builder
# ... build steps ...

# Stage 2: Runtime
FROM alpine:latest
# ... copy binary and run ...
```

### Multi-Stage Build (CGO)

See `templates/docker/Dockerfile-go-cgo` for a complete example.

**Key features:**

- Builder stage with CGO dependencies (gcc, musl-dev)
- Runtime stage with minimal C library support
- Final image ~20-30MB with CGO libraries
- CGO_ENABLED=1 for compilation

**Example structure:**

```dockerfile
# Stage 1: Build
FROM golang:1.25-alpine AS builder
RUN apk add --no-cache gcc musl-dev
# ... build with CGO ...

# Stage 2: Runtime
FROM alpine:latest
RUN apk add --no-cache ca-certificates
# ... copy binary and run ...
```

## Docker Compose for Local Development

See `templates/docker/docker-compose.yml` for a complete example.

**Common patterns:**

1. **Single service:**

   ```yaml
   services:
     app:
       build: .
       ports:
         - "8080:8080"
       volumes:
         - ./data:/data
   ```

2. **With database:**

   ```yaml
   services:
     app:
       build: .
       depends_on:
         - postgres
     postgres:
       image: postgres:16-alpine
       volumes:
         - pgdata:/var/lib/postgresql/data
   volumes:
     pgdata:
   ```

3. **Development with hot reload:**

   ```yaml
   services:
     app:
       build: .
       volumes:
         - .:/app
         - ./data:/data
       command: air  # Using cosmtrek/air for hot reload
   ```

## GitHub Container Registry (GHCR)

### Setup

1. **Enable GHCR in repository settings:**
   - Go to repository Settings → Actions → General
   - Under "Workflow permissions", select "Read and write permissions"
   - Save changes

2. **Add workflow file:**
   - See `templates/docker/github-workflows-docker.yml`
   - Place in `.github/workflows/docker-build.yml`
   - No additional secrets needed (uses `GITHUB_TOKEN`)

### Image Tagging Strategy

The template workflow creates multiple tags for flexibility:

- `latest` - Most recent main branch build
- `main` - Latest from main branch
- `sha-<short-sha>` - Specific commit (e.g., `sha-a1b2c3d`)
- `v1.2.3` - Semver tags (when pushing git tags)

**Examples:**

```bash
# Pull latest
docker pull ghcr.io/username/project:latest

# Pull specific commit
docker pull ghcr.io/username/project:sha-a1b2c3d

# Pull specific version
docker pull ghcr.io/username/project:v1.2.3
```

### Making Images Public

By default, GHCR images are private. To make public:

1. Go to `https://github.com/users/USERNAME/packages/container/PROJECT`
2. Click "Package settings"
3. Scroll to "Danger Zone"
4. Click "Change visibility" → Select "Public"

## Best Practices

### Security

1. **Use specific base image versions:**

   ```dockerfile
   FROM golang:1.25-alpine  # Pin major version
   FROM alpine:3.19         # Pin specific version
   ```

2. **Run as non-root user:**

   ```dockerfile
   RUN adduser -D -u 1000 appuser
   USER appuser
   ```

3. **Minimal runtime dependencies:**
   - Use `scratch` for pure Go when possible
   - Use `alpine` for minimal Linux utilities
   - Only install necessary packages

4. **Multi-stage builds:**
   - Never ship build tools in runtime image
   - Keep builder and runtime stages separate

### Optimization

1. **Layer caching:**

   ```dockerfile
   # Copy go.mod/go.sum first for better caching
   COPY go.mod go.sum ./
   RUN go mod download

   # Then copy source code
   COPY . .
   RUN go build
   ```

2. **Static binary compilation:**

   ```dockerfile
   RUN CGO_ENABLED=0 go build -ldflags="-w -s" -o app
   ```

   - `-w` - Omit DWARF symbol table
   - `-s` - Omit symbol table and debug info
   - Results in smaller binaries

3. **.dockerignore:**
   - See `templates/docker/.dockerignore`
   - Exclude unnecessary files from build context
   - Reduces build time and context size

### Data Persistence

1. **Volume mounting:**

   ```yaml
   volumes:
     - ./data:/data        # Development: local directory
     - appdata:/data       # Production: named volume
   ```

2. **Database files:**
   - Always use volumes for SQLite databases
   - Never store in container filesystem (lost on restart)
   - Use named volumes in production

3. **Configuration:**

   ```yaml
   # Option 1: Bake into image (recommended)
   COPY config.yaml /app/config.yaml

   # Option 2: Volume mount (for frequent changes)
   volumes:
     - ./config.yaml:/app/config.yaml:ro
   ```

### Health Checks

1. **In Dockerfile:**

   ```dockerfile
   HEALTHCHECK --interval=30s --timeout=3s \
     CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1
   ```

2. **HTTP endpoint:**

   ```go
   r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
       w.WriteHeader(http.StatusOK)
       w.Write([]byte("OK"))
   })
   ```

3. **Extended health check (with dependencies):**

   ```go
   r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
       // Check database connectivity
       if err := db.Ping(); err != nil {
           w.WriteHeader(http.StatusServiceUnavailable)
           return
       }
       w.WriteHeader(http.StatusOK)
       w.Write([]byte("OK"))
   })
   ```

### Environment Variables

1. **Configuration pattern:**

   ```go
   port := os.Getenv("PORT")
   if port == "" {
       port = "8080"  // Default
   }
   ```

2. **In docker-compose:**

   ```yaml
   environment:
     - PORT=8080
     - DATABASE_URL=postgres://user:pass@db:5432/dbname
   ```

3. **Secrets management:**
   - Use environment variables for sensitive data
   - Never commit secrets to Dockerfile
   - Use Docker secrets or external secret management

## Deployment Workflow

### Development

```bash
# Build locally
docker build -t myapp:dev .

# Run with docker-compose
docker-compose up

# Run standalone
docker run -p 8080:8080 -v ./data:/data myapp:dev
```

### Production

```bash
# Pull from GHCR
docker pull ghcr.io/username/myapp:latest

# Run with volume
docker run -d \
  --name myapp \
  -p 8080:8080 \
  -v /data/myapp:/data \
  --restart unless-stopped \
  ghcr.io/username/myapp:latest

# Check logs
docker logs -f myapp
```

### CI/CD Integration

The GitHub Actions workflow automatically:

1. Triggers on push to main or version tags
2. Builds the Docker image
3. Tags appropriately (latest, sha, version)
4. Pushes to GHCR
5. Makes image available for deployment

**Triggering a deployment:**

```bash
# Option 1: Push to main (creates 'latest' tag)
git push origin main

# Option 2: Create version tag (creates 'v1.2.3' tag)
git tag v1.2.3
git push origin v1.2.3
```

## Troubleshooting

### Build Failures

**"Cannot find module":**

- Ensure `go.mod` and `go.sum` are in build context
- Check `.dockerignore` isn't excluding them
- Run `go mod tidy` locally first

**CGO compilation errors:**

- Verify `gcc` and `musl-dev` are installed in builder stage
- Check `CGO_ENABLED=1` is set
- For cross-compilation, use appropriate toolchain

**Large image size:**

- Use multi-stage builds
- Use alpine base images
- Build with `-ldflags="-w -s"`
- Check what's in the final image: `docker run --rm -it <image> sh`

### Runtime Issues

**"Permission denied":**

- Check file permissions in volumes
- Ensure user in container has access
- On Linux, consider user ID mapping

**"Cannot connect to database":**

- Check service names in docker-compose
- Verify network connectivity: `docker-compose exec app ping db`
- Check environment variables are set correctly

**Volume data not persisting:**

- Use named volumes, not container paths
- Check volume is correctly mounted: `docker inspect <container>`
- Verify data path matches application expectations

## References

- [Docker official documentation](https://docs.docker.com/)
- [Go Docker best practices](https://docs.docker.com/language/golang/)
- [GitHub Container Registry docs](https://docs.github.com/en/packages/working-with-a-github-packages-registry/working-with-the-container-registry)
- Example implementation: [family-tasks](https://github.com/lepinkainen/family-tasks)
