# Sukuk POC Backend - Web3 API

A production-ready Web3 backend for Sukuk (Islamic bonds) on Base Testnet, providing APIs for dApps and frontends while processing blockchain events from an indexer. Features a clean domain-driven architecture with centralized response handling and comprehensive API documentation.

## 🚀 Technology Stack

- **Language**: Go 1.21+
- **Web Framework**: [Gin](https://gin-gonic.com/) - High-performance HTTP web framework
- **Database**: PostgreSQL with [GORM](https://gorm.io/) ORM
- **API Documentation**: [Swagger](https://swagger.io/) - Interactive API documentation
- **Configuration**: [godotenv](https://github.com/joho/godotenv) - Environment configuration
- **Validation**: [go-playground/validator](https://github.com/go-playground/validator) - Struct and field validation
- **Logging**: [Logrus](https://github.com/sirupsen/logrus) - Structured logging
- **Testing**: [Testify](https://github.com/stretchr/testify) - Testing framework
- **Security**: Custom middleware for authentication and CORS
- **File Uploads**: Built-in multipart form handling with validation

## 📋 Features

### Web3 Integration

- Blockchain event processing from indexer
- Wallet address management and validation
- Transaction history tracking
- Sukuk token lifecycle management
- Off-chain data management before deployment

### API Features

- RESTful API with domain-driven design
- Centralized response and error handling
- Real-time blockchain data queries
- Wallet portfolio endpoints
- Sukuk holdings and yield calculations
- Structured API responses with proper models

### Infrastructure

- Clean architecture with 6 domain-focused handlers
- Centralized error handling and response models
- Comprehensive Swagger documentation
- Request validation and error handling
- Structured logging with blockchain context
- Comprehensive testing (unit & integration)
- Environment-based configuration
- Security best practices (API keys, CORS)
- Health checks and monitoring
- File upload management (logos, prospectus PDFs)

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
│   │   └── migrations/          # SQL migration scripts
│   ├── handlers/                # Domain-driven HTTP handlers
│   │   ├── company.go          # Company management (public + admin)
│   │   ├── sukuk.go            # Sukuk management (public + admin)
│   │   ├── investment.go       # Investment tracking
│   │   ├── yield.go            # Yield distribution and claims
│   │   ├── redemption.go       # Redemption requests
│   │   ├── system.go           # System management
│   │   └── responses.go        # Centralized response models
│   ├── logger/                  # Structured logging (logrus)
│   ├── middleware/              # HTTP middleware (CORS, auth, logging)
│   ├── models/                  # Clean domain models
│   │   ├── company.go          # Company entity
│   │   ├── sukuk.go            # Sukuk entity (renamed from SukukSeries)
│   │   ├── investment.go       # Investment entity
│   │   ├── yield.go            # Yield entity (renamed from YieldClaim)
│   │   ├── redemption.go       # Redemption entity
│   │   └── system.go           # System state entity
│   ├── server/                  # Server setup and routes
│   ├── services/                # Business logic services
│   │   └── blockchain_sync.go  # Blockchain event synchronization
│   ├── testutil/                # Test utilities and helpers
│   └── utils/                   # Utility functions (file upload, etc.)
├── docs/                        # Swagger documentation
│   ├── swagger.yaml            # API specification
│   └── swagger.json            # API specification (JSON)
├── coverage/                    # Test coverage reports
├── uploads/                     # File upload storage
│   ├── logos/                   # Company logos
│   └── prospectus/              # Sukuk prospectus PDFs
├── Makefile                     # Build automation
├── go.mod & go.sum             # Go dependency management
└── README.md                   # This file
```

## 🔧 Prerequisites

- Go 1.21 or higher
- PostgreSQL 15 or higher
- Make (optional, for using Makefile commands)

## 🚀 Getting Started

### 1. Clone the repository

```bash
git clone https://sukuk-be.git
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

# (Optional) Run database schema migration for model changes
psql -h localhost -U postgres -d sukuk_poc -f internal/database/migrations/migrate_to_new_models.sql
psql -h localhost -U postgres -d sukuk_poc -f internal/database/migrations/fix_yield_claims.sql
psql -h localhost -U postgres -d sukuk_poc -f internal/database/migrations/add_distribution_date.sql
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
make swag                   # Generate Swagger documentation
make docs                   # Generate docs and show access info
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
- `internal/models`: 61.5% coverage (includes new Sukuk and Yield models)
- `internal/handlers`: Comprehensive test suites for all endpoints

Tests are co-located with source code following Go conventions:

- `internal/config/config_test.go`
- `internal/models/models_test.go`
- `internal/handlers/*_test.go`

## 📚 API Documentation

Interactive API documentation is available via Swagger UI.

### Generate Documentation

```bash
# Generate Swagger documentation
make swag

# Or manually
$(HOME)/go/bin/swag init -g cmd/server/main.go -o docs --parseDependency --parseInternal
```

### Access Documentation

1. Start the server:
   ```bash
   make run
   ```
2. Open your browser and go to:
   ```
   http://localhost:8080/swagger/index.html
   ```

### Available Documentation

The Swagger UI provides:

- **Interactive API Testing**: Test endpoints directly from the browser
- **Request/Response Examples**: See actual request and response formats with proper models
- **Authentication Guide**: Learn how to use API keys for admin endpoints
- **Complete Endpoint Coverage**: All 34 public and admin endpoints documented
- **Structured Response Models**: No more generic `additionalProperties`, all responses use typed models
- **Domain Models**: Clear separation between Company, Sukuk, Investment, Yield, and Redemption domains

## 🔐 API Security

The API uses rate limiting and API key authentication for admin operations.

### Public Endpoints (No Authentication)

- `/health` - Health check endpoint
- `/api/v1/companies` - List all companies
- `/api/v1/companies/:id` - Get company details
- `/api/v1/companies/:id/sukuks` - Get company's Sukuk series
- `/api/v1/sukuks` - List all Sukuk series with pagination
- `/api/v1/sukuks/:id` - Get Sukuk details
- `/api/v1/sukuks/:id/metrics` - Get Sukuk performance metrics
- `/api/v1/sukuks/:id/holders` - Get Sukuk holders with pagination
- `/api/v1/investments` - List investments
- `/api/v1/investments/investor/:address` - Get investments by investor
- `/api/v1/portfolio/:address/investments` - Get investor portfolio
- `/api/v1/portfolio/:address/yields/pending` - Get pending yields
- `/api/v1/yield-claims` - List yield claims
- `/api/v1/yield-claims/investor/:address` - Get yields by investor
- `/api/v1/yield-claims/sukuk/:sukukId` - Get yields by Sukuk
- `/api/v1/redemptions` - List redemptions
- `/api/v1/redemptions/investor/:address` - Get redemptions by investor
- `/api/v1/redemptions/sukuk/:sukukId` - Get redemptions by Sukuk

### Protected Admin Endpoints (API Key Required)

- `POST /api/v1/admin/companies` - Create new company
- `PUT /api/v1/admin/companies/:id` - Update company
- `POST /api/v1/admin/companies/:id/upload-logo` - Upload company logo
- `POST /api/v1/admin/sukuks` - Create new Sukuk series (off-chain data)
- `PUT /api/v1/admin/sukuks/:id` - Update Sukuk series
- `POST /api/v1/admin/sukuks/:id/upload-prospectus` - Upload Sukuk prospectus PDF
- `GET /api/v1/admin/redemptions/pending` - Get all pending redemptions
- `GET /api/v1/admin/yields/pending` - Get all pending yields
- `GET /api/v1/admin/yields/distributions` - Get yield distribution summary
- `GET /api/v1/admin/system/sync-status` - Get blockchain sync status
- `POST /api/v1/admin/system/force-sync` - Force blockchain sync

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

## 🚨 Recent Architecture Changes

The project has undergone significant refactoring to improve maintainability and developer experience:

### Model Refactoring
- **Renamed Models**: `SukukSeries` → `Sukuk`, `YieldClaim` → `Yield` for cleaner naming
- **Updated Fields**: Standardized field names across models (e.g., `Amount` → `InvestmentAmount`, `TransactionHash` → `TxHash`)
- **Domain Separation**: Clear separation between Company, Sukuk, Investment, Yield, and Redemption domains

### Handler Refactoring
- **Domain-Driven Design**: Consolidated 11+ fragmented handler files into 6 clean domain-focused files
- **Centralized Response System**: All endpoints now use structured response models
- **Consistent Error Handling**: Standardized error responses across all endpoints
- **Pagination Support**: Added pagination to list endpoints with proper metadata

### API Documentation
- **Proper Response Models**: Replaced generic `map[string]interface{}` with typed response models
- **Complete Coverage**: All 34 endpoints fully documented with request/response examples
- **Interactive Testing**: Test endpoints directly from Swagger UI

## 🤝 Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

Please ensure:

- All tests pass (`make test`)
- No linting errors (`make lint`)
- Update Swagger documentation (`make swag`)
- Follow the domain-driven structure
- Use centralized response models

## 📄 License

This project is currently unlicensed.

## 🙏 Acknowledgments

- [Gin Web Framework](https://gin-gonic.com/)
- [GORM](https://gorm.io/)
- [All other amazing open-source projects used](go.mod)

## 📞 Support

For questions or support, please open an issue in the GitHub repository.

---
