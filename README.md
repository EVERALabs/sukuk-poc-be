# Sukuk POC Backend - Web3 API

A production-ready Web3 backend for Sukuk (Islamic bonds) on blockchain, providing APIs for dApps and frontends while processing blockchain events from an indexer.

## 🚀 Technology Stack

- **Language**: Go 1.21+
- **Web Framework**: [Gin](https://gin-gonic.com/) - High-performance HTTP web framework
- **Database**: PostgreSQL with [GORM](https://gorm.io/) ORM
- **Database Migrations**: [golang-migrate](https://github.com/golang-migrate/migrate)
- **Validation**: [go-playground/validator](https://github.com/go-playground/validator) - Struct and field validation
- **Logging**: [Logrus](https://github.com/sirupsen/logrus) + Gin's built-in logger
- **Testing**: [Testify](https://github.com/stretchr/testify) with [gotestsum](https://github.com/gotestyourself/gotestsum)
- **API Documentation**: [Swag](https://github.com/swaggo/swag) - Swagger for Go
- **Email**: [Gomail](https://github.com/go-gomail/gomail) - Simple and efficient email sending
- **Configuration**: [Viper](https://github.com/spf13/viper) - Complete configuration solution
- **Security**: Custom middleware for security headers
- **CORS**: [gin-contrib/cors](https://github.com/gin-contrib/cors)
- **Compression**: [gin-contrib/gzip](https://github.com/gin-contrib/gzip)
- **Containerization**: Docker & Docker Compose
- **Code Quality**: [golangci-lint](https://golangci-lint.run/)

## 📋 Features

### Web3 Integration
- Blockchain event processing via webhooks
- Wallet address management and validation
- Ethereum signature verification
- Transaction history tracking
- Sukuk token lifecycle management

### API Features
- RESTful API for dApps/frontends
- Real-time blockchain data queries
- Wallet portfolio endpoints
- Sukuk holdings and profit calculations
- Event status tracking

### Infrastructure
- Shared database with blockchain indexer
- Request validation and error handling
- Structured logging with blockchain context
- Comprehensive testing (unit & integration)
- API documentation with Swagger UI
- Environment-based configuration
- Security best practices (API keys, rate limiting)
- Docker support for easy deployment
- Health checks and monitoring
- Email notifications (optional)

## 🏗️ Project Structure

```
sukuk-poc-be/
├── cmd/
│   └── api/
│       └── main.go              # Application entry point
├── internal/
│   ├── config/                  # Configuration management
│   ├── database/                # Database connection and migrations
│   ├── handlers/                # HTTP request handlers
│   ├── middleware/              # HTTP middleware
│   ├── models/                  # Database models (Wallet, Sukuk, Event, etc.)
│   ├── routes/                  # Route definitions
│   ├── services/                # Business logic
│   │   ├── blockchain/          # Blockchain-related services
│   │   ├── events/              # Event processing
│   │   └── wallet/              # Wallet management
│   ├── utils/                   # Utility functions
│   └── validators/              # Custom validators (addresses, signatures)
├── tests/                       # Test files
├── docs/                        # API documentation
├── scripts/                     # Utility scripts
├── .env.example                 # Environment variables example
├── .gitignore
├── .golangci.yml               # Linter configuration
├── Dockerfile
├── docker-compose.yml
├── go.mod
├── go.sum
├── Makefile                    # Build commands
├── README.md
└── TODO.md                     # Detailed implementation guide
```

## 🔧 Prerequisites

- Go 1.21 or higher
- PostgreSQL 15 or higher
- Docker and Docker Compose (optional)
- Make (optional, for using Makefile commands)

## 🚀 Getting Started

### 1. Clone the repository

```bash
git clone https://github.com/kadzu/sukuk-poc-be.git
cd sukuk-poc-be
```

### 2. Set up environment variables

```bash
cp .env.example .env
# Edit .env with your configuration
# Make sure to set blockchain RPC endpoints and contract addresses
```

### 3. Install dependencies

```bash
go mod download
```

### 4. Set up the database

```bash
# Create database
createdb sukuk_poc

# Run migrations
make migrate-up
```

### 5. Run the application

```bash
# Development mode
make run

# Or directly
go run cmd/api/main.go
```

The API will be available at `http://localhost:8080`

## 🐳 Docker Setup

### Using Docker Compose (Recommended)

```bash
# Build and start all services
docker-compose up -d

# View logs
docker-compose logs -f

# Stop all services
docker-compose down
```

### Using Docker directly

```bash
# Build image
docker build -t sukuk-poc-api .

# Run container
docker run -p 8080:8080 --env-file .env sukuk-poc-api
```

## 📝 Available Commands

```bash
make help          # Show available commands
make run           # Run the application
make build         # Build binary
make test          # Run all tests
make test-coverage # Run tests with coverage report
make lint          # Run linter
make fmt           # Format code
make migrate-up    # Run database migrations
make migrate-down  # Rollback database migrations
make swag          # Generate API documentation
make docker-build  # Build Docker image
make docker-up     # Start Docker containers
```

## 🧪 Testing

```bash
# Run all tests
make test

# Run tests with coverage
make test-coverage

# Run specific test
go test -v ./internal/handlers/...
```

## 📚 API Documentation

API documentation is automatically generated using Swagger.

1. Generate documentation:
   ```bash
   make swag
   ```

2. Access Swagger UI:
   ```
   http://localhost:8080/swagger/index.html
   ```

## 🔐 API Security

The API uses multiple security mechanisms:

### Public Endpoints
- `/api/v1/sukuks` - List all Sukuks
- `/api/v1/sukuks/:id` - Get Sukuk details
- `/api/v1/wallet/:address` - Get wallet information
- `/health` - Health check endpoint

### Protected Endpoints (API Key Required)
- `/api/v1/events/webhook` - Process blockchain events from indexer
- `/api/v1/wallet/link-email` - Link email to wallet address

Include API key in headers:
```
X-API-Key: <your-api-key>
```

## 🌐 Environment Variables

See `.env.example` for all available configuration options. Key variables include:

### Application
- `APP_ENV` - Application environment (development, staging, production)
- `APP_PORT` - Server port (default: 8080)

### Database (Shared with Indexer)
- `DB_HOST` - PostgreSQL host
- `DB_PORT` - PostgreSQL port
- `DB_NAME` - Database name (shared with indexer)
- `DB_USER` - Database user
- `DB_PASSWORD` - Database password

### Blockchain
- `CHAIN_ID` - Blockchain network ID
- `RPC_ENDPOINT` - Blockchain RPC endpoint
- `CONTRACT_ADDRESS` - Sukuk contract address

### Security
- `API_KEY` - API key for protected endpoints
- `RATE_LIMIT_PER_MIN` - Rate limit per minute

## 🚦 Health Check

The application provides health check endpoints:

- `GET /health` - Basic health check
- `GET /health/detailed` - Detailed health check including:
  - Database connectivity
  - Blockchain node status
  - Event processing status

## 🤝 Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

Please ensure:
- All tests pass
- Code is properly formatted (`make fmt`)
- No linting errors (`make lint`)
- Update documentation if needed

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🙏 Acknowledgments

- [Gin Web Framework](https://gin-gonic.com/)
- [GORM](https://gorm.io/)
- [All other amazing open-source projects used](go.mod)

## 📞 Support

For questions or support, please open an issue in the GitHub repository.

---

For detailed implementation instructions, see [TODO.md](TODO.md)