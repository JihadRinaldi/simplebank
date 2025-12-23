# Simple Bank

A simple banking REST API built with Go, Gin, and PostgreSQL. This project demonstrates best practices in building production-ready backend services.

## Features

- **User Management**: User registration and JWT-based authentication
- **Account Management**: Create and manage bank accounts
- **Money Transfers**: Secure transfer of funds between accounts
- **Transaction History**: Track all account entries and transfers
- **JWT Authentication**: Secure API endpoints with token-based auth
- **Database Migrations**: Automated schema migrations
- **Dockerized**: Full Docker and Docker Compose support

## Tech Stack

- **Language**: Go 1.25.5
- **Web Framework**: Gin
- **Database**: PostgreSQL 12
- **Database Access**: sqlc (type-safe SQL)
- **Authentication**: JWT (golang-jwt/jwt)
- **Configuration**: Viper
- **Password Hashing**: bcrypt
- **Testing**: testify, gomock
- **Containerization**: Docker & Docker Compose
- **Database Migration**: golang-migrate

## Project Structure

```
.
├── api/                  # HTTP handlers and API logic
├── db/
│   ├── migration/       # Database migration files
│   └── sqlc/           # Generated SQL code and queries
├── token/              # JWT token implementation
├── util/               # Utility functions (config, password, etc.)
├── mocks/              # Mock implementations for testing
├── main.go             # Application entry point
├── Dockerfile          # Multi-stage Docker build
├── docker-compose.yml  # Docker Compose configuration
└── Makefile           # Development commands

```

## Prerequisites

- **Go 1.25.5** or higher
- **Docker** and **Docker Compose**
- **Make** (optional, for using Makefile commands)
- **golang-migrate** (for manual migrations)

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
   - Run database migrations automatically
   - Start the API server on port 8000

3. **Access the API**
   ```
   http://localhost:8000
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

### Public Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/users` | Register a new user |
| POST | `/users/login` | User login |

### Protected Endpoints (Requires Authentication)

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/accounts` | Create a new account |
| GET | `/accounts/:id` | Get account by ID |
| GET | `/accounts` | List all accounts |
| POST | `/transfers` | Create a transfer |

## API Usage Examples

### 1. Create a User

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

### 2. Login

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
  "user": {
    "username": "johndoe",
    "full_name": "John Doe",
    "email": "john@example.com"
  }
}
```

### 3. Create Account (Protected)

```bash
curl -X POST http://localhost:8000/accounts \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <your_access_token>" \
  -d '{
    "currency": "USD"
  }'
```

### 4. Create Transfer (Protected)

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

## Configuration

Configuration is managed through the [app.env](app.env) file:

```env
DB_DRIVER=postgres
DB_SOURCE=postgresql://root:secret@localhost:5432/simple_bank?sslmode=disable
HTTP_SERVER_ADDRESS=0.0.0.0:8000
GRPC_SERVER_ADDRESS=0.0.0.0:9000
TOKEN_SYMMETRIC_KEY=12345678901234567890123456789012
ACCESS_TOKEN_DURATION=10m
```

**Note**: For production, use environment variables and secure secret management.

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

## Database Schema

### Tables

- **users**: User accounts with hashed passwords
- **accounts**: Bank accounts with balance and currency
- **entries**: Account balance change records
- **transfers**: Money transfer records between accounts

### Key Features

- **ACID Transactions**: All money transfers are atomic
- **Foreign Keys**: Ensures referential integrity
- **Indexes**: Optimized queries on frequently accessed columns
- **Constraints**: Currency validation, positive balance checks

## Security

- Passwords are hashed using bcrypt
- JWT tokens for authentication
- SQL injection prevention through parameterized queries
- Input validation on all endpoints
- CORS can be configured in production

## Testing

The project includes:

- **Unit tests**: For business logic and handlers
- **Integration tests**: For database operations
- **Mock implementations**: Using gomock for isolated testing
- **Table-driven tests**: For comprehensive test coverage

## Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

