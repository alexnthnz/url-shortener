# URL Shortener

A scalable URL shortener service built in Go with PostgreSQL and Redis, designed to handle millions of daily requests with high availability and low latency.

## Features

- **Fast URL Shortening**: Counter-based short code generation using base62 encoding
- **Custom Aliases**: Support for user-defined short codes
- **High Performance**: Redis caching for frequent redirects
- **Analytics**: Click tracking and basic statistics
- **Security**: URL validation, rate limiting, and CORS support
- **Scalable**: Designed for horizontal scaling and high availability
- **Containerized**: Docker support for easy deployment

## Architecture

### System Components

- **API Layer**: RESTful endpoints built with Gin framework
- **Shortening Service**: Generates unique short codes and handles URL mappings
- **Database**: PostgreSQL for persistent storage of URL mappings and analytics
- **Cache**: Redis for high-performance URL lookups
- **Analytics Service**: Asynchronous click tracking and statistics

### Key Design Decisions

- **Counter-based Short Codes**: Ensures uniqueness and collision-free generation
- **Base62 Encoding**: Generates compact, URL-safe short codes (A-Z, a-z, 0-9)
- **Caching Strategy**: 24-hour TTL on Redis cache for hot URLs
- **Async Analytics**: Non-blocking click tracking for optimal redirect performance

## Quick Start

### Prerequisites

- Go 1.22+
- Docker and Docker Compose
- PostgreSQL 15+
- Redis 7+

### Development Setup

1. **Clone the repository**
```bash
git clone <repository-url>
cd url-shortener
```

2. **Start dependencies using Docker**
```bash
make docker-up
# or
docker-compose up -d postgres redis
```

3. **Copy environment configuration**
```bash
cp .env.example .env
```

4. **Run the application**
```bash
make run
# or
go run ./cmd/server
```

The server will start on `http://localhost:8080`

### Alternative: Full Docker Setup

To run everything with Docker:
```bash
make docker-run-all
# or
docker-compose up
```

## API Documentation

### Base URL
```
http://localhost:8080
```

### Endpoints

#### 1. Shorten URL
Create a short URL from a long URL.

**Request:**
```http
POST /api/v1/shorten
Content-Type: application/json

{
  "url": "https://example.com/very/long/url/that/needs/shortening",
  "custom_alias": "my-link" // optional
}
```

**Response:**
```http
HTTP/1.1 201 Created
Content-Type: application/json

{
  "short_code": "dnh",
  "short_url": "http://localhost:8080/dnh",
  "original_url": "https://example.com/very/long/url/that/needs/shortening"
}
```

#### 2. Redirect to Original URL
Access a short URL to redirect to the original URL.

**Request:**
```http
GET /{short_code}
```

**Response:**
```http
HTTP/1.1 301 Moved Permanently
Location: https://example.com/very/long/url/that/needs/shortening
```

#### 3. Get URL Statistics
Retrieve click statistics for a short URL.

**Request:**
```http
GET /api/v1/urls/{short_code}/stats
```

**Response:**
```http
HTTP/1.1 200 OK
Content-Type: application/json

{
  "short_code": "dnh",
  "original_url": "https://example.com/very/long/url/that/needs/shortening",
  "click_count": 42,
  "created_at": "2024-01-15T10:30:00Z"
}
```

#### 4. Health Check
Check service health.

**Request:**
```http
GET /health
```

**Response:**
```http
HTTP/1.1 200 OK
Content-Type: application/json

{
  "status": "healthy"
}
```

## Usage Examples

### cURL Examples

**Shorten a URL:**
```bash
curl -X POST http://localhost:8080/api/v1/shorten \
  -H "Content-Type: application/json" \
  -d '{"url": "https://github.com/alexnthnz/url-shortener"}'
```

**Shorten with custom alias:**
```bash
curl -X POST http://localhost:8080/api/v1/shorten \
  -H "Content-Type: application/json" \
  -d '{"url": "https://github.com/alexnthnz/url-shortener", "custom_alias": "my-repo"}'
```

**Get statistics:**
```bash
curl http://localhost:8080/api/v1/urls/dnh/stats
```

**Test redirect:**
```bash
curl -I http://localhost:8080/dnh
```

### JavaScript Example

```javascript
// Shorten a URL
const response = await fetch('http://localhost:8080/api/v1/shorten', {
  method: 'POST',
  headers: {
    'Content-Type': 'application/json',
  },
  body: JSON.stringify({
    url: 'https://example.com/long-url',
    custom_alias: 'my-link' // optional
  })
});

const data = await response.json();
console.log('Short URL:', data.short_url);

// Get statistics
const statsResponse = await fetch(`http://localhost:8080/api/v1/urls/${data.short_code}/stats`);
const stats = await statsResponse.json();
console.log('Click count:', stats.click_count);
```

## Configuration

The application can be configured using environment variables:

| Variable | Description | Default |
|----------|-------------|---------|
| `PORT` | Server port | `8080` |
| `ENVIRONMENT` | Environment (development/production) | `development` |
| `BASE_URL` | Base URL for short links | `http://localhost:8080` |
| `DATABASE_URL` | PostgreSQL connection string | `postgres://localhost:5432/urlshortener?sslmode=disable` |
| `REDIS_URL` | Redis connection string | `redis://localhost:6379` |

## Development

### Available Make Commands

```bash
make build       # Build the application
make run         # Run the application
make test        # Run tests
make clean       # Clean build artifacts
make docker-up   # Start PostgreSQL and Redis
make docker-down # Stop development dependencies
make deps        # Download dependencies
make fmt         # Format code
make lint        # Run linter
make dev         # Full development setup
```

### Project Structure

```
url-shortener/
├── cmd/server/           # Application entry point
├── internal/
│   ├── config/          # Configuration management
│   ├── handlers/        # HTTP handlers and middleware
│   ├── models/          # Data models
│   ├── repository/      # Database and cache layers
│   └── services/        # Business logic
├── docker-compose.yml   # Development dependencies
├── Dockerfile          # Production container
├── Makefile           # Development commands
└── .env.example       # Environment configuration template
```

## Performance Characteristics

- **Latency**: < 10ms average response time for cached redirects
- **Throughput**: Handles 10K+ requests per second per instance
- **Cache Hit Ratio**: > 90% for hot URLs
- **Scalability**: Horizontally scalable with load balancing

## Security Features

- **URL Validation**: Prevents malicious redirects (XSS, file://, etc.)
- **Rate Limiting**: 100 requests per minute per IP address
- **Input Sanitization**: Validates and sanitizes all user inputs
- **HTTPS Support**: Enforced in production environments
- **Custom Alias Validation**: Prevents reserved words and invalid characters

## Monitoring and Observability

- **Structured Logging**: JSON-formatted logs with request tracing
- **Health Checks**: Built-in health endpoint for load balancers
- **Metrics**: Ready for Prometheus integration
- **Analytics**: Click tracking and statistics

## Production Deployment

### Docker

1. **Build the image:**
```bash
make docker-build
```

2. **Run with Docker Compose:**
```bash
docker-compose up
```

### Environment Setup

For production deployment, ensure:

1. Set `ENVIRONMENT=production`
2. Use secure database credentials
3. Configure Redis with persistence
4. Set up SSL/TLS termination
5. Configure monitoring and alerting

## API Rate Limits

- **Default**: 100 requests per minute per IP address
- **Shorten endpoint**: Same rate limit applies
- **Redirect endpoint**: No additional rate limiting (cached responses)
- **Stats endpoint**: Same rate limit applies

## Error Handling

The API returns appropriate HTTP status codes:

- `200` - Success
- `201` - Resource created
- `301` - Permanent redirect
- `400` - Bad request (invalid input)
- `404` - Short URL not found
- `429` - Rate limit exceeded
- `500` - Internal server error

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests for new functionality
5. Run `make test` and `make lint`
6. Submit a pull request

## License

This project is licensed under the MIT License - see the LICENSE file for details.

## System Design

This implementation follows the system design specifications for a scalable URL shortener:

- **Counter-based short code generation** for guaranteed uniqueness
- **Base62 encoding** for compact, URL-safe identifiers
- **PostgreSQL** for reliable data persistence
- **Redis caching** for high-performance redirects
- **Horizontal scalability** through stateless design
- **Comprehensive analytics** for usage tracking
- **Security measures** against malicious usage

For detailed system design documentation, see the original design specification.