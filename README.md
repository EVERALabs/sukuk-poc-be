# Sukuk POC Backend - Web3 API

A production-ready Web3 backend for Sukuk (Islamic bonds) on Base Testnet, providing APIs for dApps and frontends while processing blockchain events from an indexer.

## 🚀 Technology Stack

- **Language**: Go 1.24+
- **Web Framework**: [Gin](https://gin-gonic.com/) - High-performance HTTP web framework
- **Database**: PostgreSQL with [GORM](https://gorm.io/) ORM
- **Configuration**: [godotenv](https://github.com/joho/godotenv) - Environment configuration
- **Validation**: [go-playground/validator](https://github.com/go-playground/validator) - Struct and field validation
- **Logging**: [Logrus](https://github.com/sirupsen/logrus) - Structured logging
- **Testing**: [Testify](https://github.com/stretchr/testify) - Testing framework
- **Security**: Custom middleware for authentication and CORS
- **File Uploads**: Built-in multipart form handling with validation

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
- Environment-based configuration
- Security best practices (API keys, rate limiting)
- Health checks and monitoring
- File upload management (logos, PDFs)

## 🏗️ Project Structure

```
sukuk-poc-be/
├── cmd/                         # Application entry points
│   ├── migrate/                 # Database migration command
│   ├── seed/                    # Database seeding command
│   └── server/                  # Main API server
├── internal/                    # Internal packages (Go convention)
│   ├── config/                  # Configuration management (godotenv)
│   ├── database/                # Database connection and setup
│   ├── handlers/                # HTTP request handlers with tests
│   ├── logger/                  # Structured logging (logrus)
│   ├── middleware/              # HTTP middleware (CORS, auth, logging)
│   ├── models/                  # Database models with tests
│   ├── server/                  # Server setup and routes
│   ├── testutil/                # Test utilities and helpers
│   └── utils/                   # Utility functions (file upload, etc.)
├── coverage/                    # Test coverage reports
├── uploads/                     # File upload storage
│   ├── logos/                   # Company logos
│   └── prospectus/              # Sukuk prospectus PDFs
├── Makefile                     # Build automation
├── go.mod & go.sum             # Go dependency management
└── README.md                   # This file
```

## 🔧 Prerequisites

- Go 1.24 or higher
- PostgreSQL 15 or higher
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
make migrate
```

### 5. Run the application

```bash
# Development mode
make run

# Or directly
go run cmd/server/main.go
```

The API will be available at `http://localhost:8080`

## 📝 Available Commands

```bash
make help                   # Show available commands
make run                    # Run the application
make build                  # Build binary
make test                   # Run all tests
make test-coverage          # Run tests with coverage report
make test-coverage-profile  # Generate detailed coverage profile
make lint                   # Run linter (if available)
make clean                  # Clean build artifacts
make migrate                # Run database migrations
make seed                   # Seed database with sample data
```

## 🧪 Testing

The project uses Go's built-in testing framework with testify for assertions:

```bash
# Run all tests
make test

# Run tests with coverage
make test-coverage

# Generate detailed coverage profile
make test-coverage-profile

# Run specific package tests
go test -v ./internal/config
go test -cover ./internal/models

# Run specific test
go test -v -run TestHealthTestSuite ./internal/handlers
```

### Test Coverage

Current test coverage by package:
- `internal/config`: 86.0% coverage
- `internal/models`: 61.5% coverage
- `internal/handlers`: Comprehensive test suites for all endpoints

Tests are co-located with source code following Go conventions:
- `internal/config/config_test.go`
- `internal/models/models_test.go`
- `internal/handlers/*_test.go`

## 📚 API Documentation

API documentation is currently under development. The API follows RESTful principles with JSON responses.

## 🔐 API Security

The API uses rate limiting and API key authentication for admin operations.

### Public Endpoints (No Authentication)
- `/health` - Health check endpoint
- `/api/v1/companies` - List all companies
- `/api/v1/companies/:id` - Get company details
- `/api/v1/companies/:id/sukuks` - Get company's Sukuk series
- `/api/v1/sukuks` - List all Sukuk series
- `/api/v1/sukuks/:id` - Get Sukuk details
- `/api/v1/portfolio/:address` - Get investor portfolio
- `/api/v1/investments` - List investments
- `/api/v1/yield-claims` - List yield claims
- `/api/v1/redemptions` - List redemptions

### Protected Admin Endpoints (API Key Required)
- `/api/v1/admin/companies` - Create/update companies
- `/api/v1/admin/sukuks` - Create/update Sukuk series
- `/api/v1/admin/events/webhook` - Process blockchain events from indexer

Include API key in headers for admin endpoints:
```
X-API-Key: <your-api-key>
```

## 🌐 Environment Variables

See `.env.example` for all available configuration options. Key variables include:

### Application
- `APP_NAME` - Application name
- `APP_ENV` - Application environment (development, staging, production)
- `APP_PORT` - Server port (default: 8080)
- `APP_DEBUG` - Debug mode (true/false)
- `APP_UPLOAD_DIR` - File upload directory

### Database
- `DB_HOST` - PostgreSQL host
- `DB_PORT` - PostgreSQL port
- `DB_NAME` - Database name
- `DB_USER` - Database user
- `DB_PASSWORD` - Database password

### Blockchain (Base Testnet)
- `BLOCKCHAIN_CHAIN_ID` - Chain ID (84532 for Base Testnet)
- `BLOCKCHAIN_RPC_ENDPOINT` - Base Testnet RPC endpoint
- `BLOCKCHAIN_CONTRACT_ADDRESS` - Your Sukuk contract address

### API Security
- `API_API_KEY` - API key for protected admin endpoints
- `API_RATE_LIMIT_PER_MIN` - Rate limit per minute
- `API_ALLOWED_ORIGINS` - CORS allowed origins

### Logging
- `LOGGER_LEVEL` - Log level (debug, info, warn, error)
- `LOGGER_FORMAT` - Log format (json, text)

## 🚦 Health Check

The application provides a comprehensive health check endpoint:

- `GET /health` - Complete health check including:
  - Database connectivity and performance
  - System resources (CPU, memory, goroutines)
  - Application statistics
  - File upload directory status

## 🤝 Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

Please ensure:
- All tests pass (`make test`)
- No linting errors (`make lint`)
- Update documentation if needed

## 📄 License

This project is currently unlicensed.

## 🙏 Acknowledgments

- [Gin Web Framework](https://gin-gonic.com/)
- [GORM](https://gorm.io/)
- [All other amazing open-source projects used](go.mod)

## 📞 Support

For questions or support, please open an issue in the GitHub repository.

---

For detailed implementation instructions, see [TODO.md](TODO.md)