# Simple Bank

A production-ready banking application with both REST API and gRPC services built with Go, PostgreSQL, and modern backend best practices. Features async task processing with Redis and comprehensive test coverage.

## Features

- **Dual API Support**: Both REST (HTTP) and gRPC APIs
- **User Management**: Registration with email verification, JWT authentication, and profile updates
- **Account Management**: Create and manage multi-currency bank accounts
- **Money Transfers**: Secure, atomic transfer of funds between accounts
- **Transaction History**: Complete audit trail of all account activities
- **Email Verification**: Async email verification with secret codes
- **Async Task Processing**: Background workers using Asynq and Redis
- **JWT Authentication**: Secure endpoints with access and refresh tokens
- **Database Migrations**: Automated schema version control
- **Comprehensive Testing**: 71% test coverage with unit and integration tests
- **Dockerized**: Full Docker and Docker Compose support

## Tech Stack

- **Language**: Go 1.25.5
- **Web Frameworks**: Gin (REST), gRPC
- **Database**: PostgreSQL 12
- **Cache/Queue**: Redis (Asynq)
- **Database Access**: sqlc (type-safe SQL)
- **Authentication**: JWT (golang-jwt/jwt)
- **Configuration**: Viper
- **Password Hashing**: bcrypt
- **Testing**: testify/mock (71% coverage)
- **Protocol Buffers**: protoc, protoc-gen-go, protoc-gen-grpc-gateway
- **Task Queue**: Asynq (Redis-based async workers)
- **Containerization**: Docker & Docker Compose
- **Database Migration**: golang-migrate

## Project Structure

```
.
├── api/                  # REST API handlers (Gin)
├── gapi/                 # gRPC API handlers with comprehensive tests
├── db/
│   ├── migration/       # Database migration files
│   ├── query/          # SQL queries for sqlc
│   └── sqlc/           # Generated SQL code and transactions
├── pb/                  # Generated Protocol Buffer code
├── proto/              # Protocol Buffer definitions
├── token/              # JWT token implementation
├── validator/          # Input validation utilities
├── worker/             # Async task workers (Asynq)
├── util/               # Utility functions (config, password, random)
├── mocks/              # Mock implementations for testing
├── main.go             # Application entry point
├── Dockerfile          # Multi-stage Docker build
├── docker-compose.yml  # Docker Compose configuration
└── Makefile           # Development commands
```

## Prerequisites

- **Go 1.25.5** or higher
- **Docker** and **Docker Compose**
- **PostgreSQL 12** or higher
- **Redis** (for async task queue)
- **Make** (optional, for using Makefile commands)
- **golang-migrate** (for manual migrations)
- **Protocol Buffers compiler** (protoc) - for gRPC development

## Getting Started

### Option 1: Using Docker Compose (Recommended)

1. **Clone the repository**
   ```bash
   git clone https://github.com/JihadRinaldi/simplebank.git
   cd simplebank
   ```

2. **Start the application**
   ```bash
   docker-compose up
   ```

   This will:
   - Start PostgreSQL database
   - Start Redis for task queue
   - Run database migrations automatically
   - Start the HTTP API server on port 8000
   - Start the gRPC server on port 9090

3. **Access the APIs**
   ```
   REST API:  http://localhost:8000
   gRPC API:  localhost:9090
   ```

### Option 2: Local Development

1. **Start PostgreSQL**
   ```bash
   make postgres
   ```

2. **Create database**
   ```bash
   make createdb
   ```

3. **Run migrations**
   ```bash
   make migrateup
   ```

4. **Start the server**
   ```bash
   make server
   ```

## API Endpoints

### REST API (HTTP) - Port 8000

#### Public Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/users` | Register a new user |
| POST | `/users/login` | User login |

#### Protected Endpoints (Requires Authentication)

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/accounts` | Create a new account |
| GET | `/accounts/:id` | Get account by ID |
| GET | `/accounts` | List all accounts |
| POST | `/transfers` | Create a transfer |
| PATCH | `/users/:username` | Update user profile |

### gRPC API - Port 9090

All gRPC services with comprehensive unit test coverage:

| Service | Method | Description | Test Coverage |
|---------|--------|-------------|---------------|
| SimpleBank | CreateUser | Register user with email verification | ✅ 6 test cases |
| SimpleBank | LoginUser | Authenticate user and create session | ✅ 6 test cases |
| SimpleBank | UpdateUser | Update user profile (authenticated) | ✅ 8 test cases |
| SimpleBank | VerifyEmail | Verify email with secret code | ✅ 5 test cases |

**Total Test Coverage**: 25 test cases across all gRPC endpoints with 71% code coverage

## API Usage Examples

### REST API Examples

#### 1. Create a User

```bash
curl -X POST http://localhost:8000/users \
  -H "Content-Type: application/json" \
  -d '{
    "username": "johndoe",
    "password": "secret123",
    "full_name": "John Doe",
    "email": "john@example.com"
  }'
```

#### 2. Login

```bash
curl -X POST http://localhost:8000/users/login \
  -H "Content-Type: application/json" \
  -d '{
    "username": "johndoe",
    "password": "secret123"
  }'
```

Response:
```json
{
  "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "session_id": "550e8400-e29b-41d4-a716-446655440000",
  "user": {
    "username": "johndoe",
    "full_name": "John Doe",
    "email": "john@example.com"
  }
}
```

#### 3. Create Account (Protected)

```bash
curl -X POST http://localhost:8000/accounts \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <your_access_token>" \
  -d '{
    "currency": "USD"
  }'
```

#### 4. Create Transfer (Protected)

```bash
curl -X POST http://localhost:8000/transfers \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <your_access_token>" \
  -d '{
    "from_account_id": 1,
    "to_account_id": 2,
    "amount": 100,
    "currency": "USD"
  }'
```

### gRPC API Examples

Use tools like [grpcurl](https://github.com/fullstorydev/grpcurl) or [BloomRPC](https://github.com/bloomrpc/bloomrpc) to test gRPC endpoints.

#### 1. CreateUser (gRPC)

```bash
grpcurl -plaintext -d '{
  "username": "johndoe",
  "password": "secret123",
  "full_name": "John Doe",
  "email": "john@example.com"
}' localhost:9090 simplebank.SimpleBank/CreateUser
```

#### 2. LoginUser (gRPC)

```bash
grpcurl -plaintext -d '{
  "username": "johndoe",
  "password": "secret123"
}' localhost:9090 simplebank.SimpleBank/LoginUser
```

#### 3. UpdateUser (gRPC - Authenticated)

```bash
grpcurl -plaintext \
  -H "authorization: Bearer <your_access_token>" \
  -d '{
    "username": "johndoe",
    "full_name": "John Updated Doe",
    "email": "johnupdated@example.com"
  }' localhost:9090 simplebank.SimpleBank/UpdateUser
```

#### 4. VerifyEmail (gRPC)

```bash
grpcurl -plaintext -d '{
  "email_id": 1,
  "secret_code": "abc123xyz789"
}' localhost:9090 simplebank.SimpleBank/VerifyEmail
```

## Configuration

Configuration is managed through the [app.env](app.env) file:

```env
# Database
DB_DRIVER=postgres
DB_SOURCE=postgresql://root:secret@localhost:5432/simple_bank?sslmode=disable

# Server Addresses
HTTP_SERVER_ADDRESS=0.0.0.0:8000
GRPC_SERVER_ADDRESS=0.0.0.0:9090

# Redis (for async tasks)
REDIS_ADDRESS=localhost:6379

# Authentication
TOKEN_SYMMETRIC_KEY=12345678901234567890123456789012
ACCESS_TOKEN_DURATION=15m
REFRESH_TOKEN_DURATION=24h

# Email (optional - for production)
EMAIL_SENDER_NAME=Simple Bank
EMAIL_SENDER_ADDRESS=noreply@simplebank.com
```

**Note**: For production, use environment variables and secure secret management (e.g., AWS Secrets Manager, HashiCorp Vault).

## Development

### Available Make Commands

```bash
make postgres          # Start PostgreSQL container
make createdb          # Create database
make dropdb           # Drop database
make migrateup        # Run all migrations
make migratedown      # Rollback all migrations
make migrateup_n      # Run n migrations
make migratedown_n    # Rollback n migrations
make sqlc             # Generate SQL code
make test             # Run tests
make server           # Start the server
make mockery          # Generate mocks
```

### Running Tests

```bash
# Run all tests
make test

# Run tests with coverage
go test -v -cover ./...

# Run specific package tests
go test -v ./api/...
go test -v ./gapi/...
```

### Database Migrations

Create a new migration:

```bash
migrate create -ext sql -dir db/migration -seq <migration_name>
```

Apply migrations:

```bash
make migrateup
```

Rollback migrations:

```bash
make migratedown
```

## Docker

### Build Docker Image

```bash
docker build -t simplebank:latest .
```

### Run with Docker Compose

```bash
# Start services
docker-compose up -d

# View logs
docker-compose logs -f

# Stop services
docker-compose down

# Remove volumes (deletes database data)
docker-compose down -v
```

### Docker Architecture

The Dockerfile uses a **multi-stage build**:

1. **Build Stage**: Compiles the Go application and downloads migration tool
2. **Run Stage**: Uses Alpine Linux for a minimal runtime image (~20MB)

Benefits:
- Small final image size
- Secure (minimal attack surface)
- Fast deployment

## Async Task Processing

The application uses **Asynq** (Redis-based task queue) for asynchronous background processing:

### Features

- **Email Verification**: Sends verification emails asynchronously after user registration
- **Task Retry**: Configurable retry logic (up to 10 retries)
- **Delayed Processing**: Tasks can be scheduled with delays
- **Priority Queues**: Critical tasks use high-priority queue
- **Worker Monitoring**: Built-in monitoring and observability

### Worker Tasks

1. **SendVerifyEmail**
   - Triggered on user registration
   - Sends verification email with secret code
   - Retries: 10 times with exponential backoff
   - Queue: Critical priority
   - Delay: 10 seconds after creation

## Database Schema

### Tables

- **users**: User accounts with hashed passwords and email verification status
- **verify_emails**: Email verification records with secret codes
- **sessions**: User sessions with refresh tokens
- **accounts**: Bank accounts with balance and currency
- **entries**: Account balance change records
- **transfers**: Money transfer records between accounts

### Key Features

- **ACID Transactions**: All money transfers and user creation are atomic
- **Foreign Keys**: Ensures referential integrity
- **Indexes**: Optimized queries on frequently accessed columns
- **Constraints**: Currency validation, positive balance checks
- **Email Verification**: Secure email verification workflow

## Security

- Passwords are hashed using bcrypt
- JWT tokens for authentication
- SQL injection prevention through parameterized queries
- Input validation on all endpoints
- CORS can be configured in production

## Testing

The project includes comprehensive test coverage with modern testing practices:

### Test Structure

- **Unit Tests**: Business logic and handlers with mocked dependencies
- **Integration Tests**: Database operations with real PostgreSQL
- **Mock Implementations**: Using testify/mock for isolated testing
- **Table-Driven Tests**: Comprehensive edge case coverage

### gRPC Test Coverage (71%)

All gRPC endpoints have complete unit test coverage:

#### CreateUser Tests (6 cases)
- ✅ Successful user creation with email verification
- ✅ Internal database errors
- ✅ Invalid username validation
- ✅ Invalid email validation
- ✅ Password too short
- ✅ Invalid full name

#### LoginUser Tests (6 cases)
- ✅ Successful login with token generation
- ✅ User not found
- ✅ Incorrect password
- ✅ Invalid username format
- ✅ Password too short
- ✅ Session creation failure

#### UpdateUser Tests (8 cases)
- ✅ Update full name and email
- ✅ Update only full name
- ✅ Update only email
- ✅ Update only password (with hash verification)
- ✅ Invalid email format
- ✅ Expired authentication token
- ✅ Missing authorization
- ✅ Permission denied (update other user)

#### VerifyEmail Tests (5 cases)
- ✅ Successful email verification
- ✅ Invalid email ID
- ✅ Invalid secret code
- ✅ Verification record not found
- ✅ Internal database errors

### Running Tests

```bash
# Run all tests
make test

# Run tests with coverage
go test -v -cover ./...

# Run specific package tests
go test -v ./api/...
go test -v ./gapi/...

# Run gRPC tests only
go test -v -cover ./gapi/...
# Output: coverage: 71.0% of statements
```

### Test Features

- **Mock Store**: Full database mock implementation
- **Mock TaskDistributor**: Async task queue mocking
- **Context with Auth**: Bearer token helper for authenticated tests
- **Password Verification**: Hash validation in tests
- **gRPC Status Codes**: Proper error code validation
- **Shared Test Utilities**: Reusable helpers in `main_test.go`

## Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

