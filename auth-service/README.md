# Auth Service

Authentication service for the Sales Tracking App.

## Features

- User registration with email verification
- JWT-based authentication
- Role-based access control (RBAC)
- Password recovery system
- PostgreSQL database integration
- Docker support
- Swagger/OpenAPI documentation

## Prerequisites

- Go 1.21+
- Docker and Docker Compose
- PostgreSQL 15+

## Setup

1. Clone the repository
2. Create a `.env` file in the root directory with your configuration
3. Build and run the service:

```bash
# Build and run with Docker Compose
docker-compose up --build

# Or run locally
go run cmd/main.go
```

## API Documentation

The API documentation is available at:

```
http://localhost:8080/api/swagger
```

## API Endpoints

### Public Endpoints

- `POST /api/auth/register` - Register a new user
- `POST /api/auth/login` - Login and get JWT token
- `POST /api/auth/forgot-password` - Request password reset
- `POST /api/auth/reset-password` - Reset password with token

### Protected Endpoints

- `GET /api/health` - Health check endpoint (requires valid JWT token)

## Environment Variables

- `AUTH_PORT`: Service port (default: 8080)
- `AUTH_JWT_SECRET`: JWT secret key
- `AUTH_DATABASE_URL`: PostgreSQL connection string
- `AUTH_DATABASE_NAME`: Database name
- `AUTH_DATABASE_USER`: Database user
- `AUTH_DATABASE_PASS`: Database password
- `AUTH_DATABASE_HOST`: Database host
- `AUTH_DATABASE_PORT`: Database port
- `AUTH_SMTP_HOST`: SMTP server host
- `AUTH_SMTP_PORT`: SMTP server port
- `AUTH_SMTP_USER`: SMTP user
- `AUTH_SMTP_PASS`: SMTP password
- `AUTH_FRONTEND_URL`: Frontend URL for email links
- `AUTH_PASSWORD_RESET`: Password reset path
- `AUTH_VERIFICATION`: Email verification path
