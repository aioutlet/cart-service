# # Cart Service 🛒

A high-performance, Redis-backed microservice for managing shopping carts in the AI Outlet e-commerce platform. Built with Go, Gin, and following microservices best practices.

## 🚀 Features

- **High Performance**: Built with Go and Gin for maximum throughput
- **Redis Storage**: Lightning-fast cart operations with automatic expiration
- **JWT Authentication**: Secure user authentication with flexible claims
- **Guest Support**: Full cart functionality for anonymous users
- **Distributed Locking**: Prevents race conditions in concurrent operations
- **Auto-validation**: Real-time product and inventory validation
- **Cart Transfer**: Seamless guest-to-user cart migration
- **Comprehensive Monitoring**: Structured logging and health checks
- **Distributed Tracing**: OpenTelemetry integration with Jaeger support
- **API Documentation**: Complete OpenAPI/Swagger documentation

## 🏗️ Architecture

The Cart Service follows clean architecture principles with clear separation of concerns:

```text
cart-service/
├── cmd/server/          # Application entrypoint
├── internal/
│   ├── config/          # Configuration management
│   ├── handlers/        # HTTP request handlers
│   ├── middleware/      # HTTP middleware
│   ├── models/          # Domain models
│   ├── repository/      # Data access layer
│   └── services/        # Business logic layer
├── pkg/
│   ├── clients/         # External service clients
│   ├── logger/          # Logging utilities
│   └── redis/           # Redis client setup
├── tests/
│   ├── mocks/           # Test mocks
│   ├── testutils/       # Test utilities
│   └── unit/            # Unit tests
├── docs/                # Swagger documentation
└── deployments/         # Docker configurations
```

## 🛠️ Technology Stack

- **Language**: Go 1.21+
- **Web Framework**: Gin
- **Database**: Redis 7+
- **Authentication**: JWT
- **Distributed Tracing**: OpenTelemetry + Jaeger
- **Documentation**: Swagger/OpenAPI 3.0
- **Testing**: Testify, Go standard testing
- **Containerization**: Docker & Docker Compose
- **Logging**: Uber Zap

## 📋 Prerequisites

- Go 1.21 or higher
- Redis 7.0 or higher
- Docker & Docker Compose (optional)

## 🚀 Quick Start

### Local Development

1. **Clone the repository**:

   ```bash
   git clone <repository-url>
   cd cart-service
   ```

2. **Install dependencies**:

   ```bash
   make deps
   ```

3. **Copy environment configuration**:

   ```bash
   cp .env.example .env
   ```

4. **Start Redis**:

   ```bash
   make redis-start
   ```

5. **Run the service**:

   ```bash
   make run
   ```

The service will be available at `http://localhost:8085`

### Docker Deployment

1. **Start with Docker Compose**:

   ```bash
   make docker-run
   ```

2. **View logs**:

   ```bash
   make docker-logs
   ```

3. **Stop services**:

   ```bash
   make docker-stop
   ```

## 📚 API Documentation

### Base URL

```text
http://localhost:8085/api/v1
```

### Authentication

Add JWT token to requests:

```text
Authorization: Bearer <your-jwt-token>
```

### Core Endpoints

#### Authenticated Cart Operations

```http
GET    /cart                    # Get user's cart
POST   /cart/items             # Add item to cart
PUT    /cart/items/{productId} # Update item quantity
DELETE /cart/items/{productId} # Remove item from cart
DELETE /cart                   # Clear entire cart
POST   /cart/transfer          # Transfer guest cart to user
```

#### Guest Cart Operations

```http
GET    /guest/cart/{guestId}                    # Get guest cart
POST   /guest/cart/{guestId}/items             # Add item to guest cart
PUT    /guest/cart/{guestId}/items/{productId} # Update guest cart item
DELETE /guest/cart/{guestId}/items/{productId} # Remove guest cart item
DELETE /guest/cart/{guestId}                   # Clear guest cart
```

#### System Endpoints

```http
GET /health                     # Health check
GET /swagger/*                  # API documentation
```

## 🧪 Testing

### Run All Tests

```bash
make test
```

### Run Tests with Coverage

```bash
make test-coverage
```

### Run Specific Tests

```bash
go test ./internal/models -v
go test ./internal/services -v
go test ./internal/handlers -v
```

## 🔧 Configuration

Environment variables can be set in `.env` file. See `.env.example` for all available options.

## 🔍 Distributed Tracing

The cart service includes comprehensive distributed tracing using OpenTelemetry and Jaeger:

### Features

- **Full Request Tracing**: Every HTTP request creates a trace with correlation ID
- **Service-to-Service Propagation**: Trace context propagated to external service calls
- **Detailed Spans**: Individual operations (cart operations, product lookups, inventory checks) are tracked as spans
- **Error Tracking**: Errors and exceptions are recorded with trace context
- **Performance Monitoring**: Request latency and operation timing tracked

### Configuration

Set the following environment variables:

```bash
TRACING_ENABLED=true
TRACING_SERVICE_NAME=cart-service
TRACING_SERVICE_VERSION=1.0.0
TRACING_JAEGER_ENDPOINT=http://localhost:14268/api/traces
TRACING_SAMPLE_RATE=1.0
```

### Running with Jaeger

1. **Start Jaeger using Docker Compose**:

   ```bash
   docker-compose up jaeger redis
   ```

2. **Run the cart service**:

   ```bash
   make run
   ```

3. **Access Jaeger UI**:

   Open <http://localhost:16686> to view traces

### Trace Information

Each request includes:

- **Trace ID**: Unique identifier for the entire request flow
- **Span ID**: Identifier for individual operations
- **Correlation ID**: Custom correlation ID for cross-service tracking
- **User Context**: User ID and authentication information
- **Operation Metadata**: Product IDs, quantities, cart details
- **Performance Metrics**: Request duration and latency

### Example Trace Flow

```text
Request: POST /api/v1/cart/items
├── CartHandler.AddItem (span)
│   ├── CartService.AddItem (span)
│   │   ├── CartRepository.AcquireLock (span)
│   │   ├── ProductClient.GetProduct (span)
│   │   ├── InventoryClient.CheckAvailability (span)
│   │   ├── CartService.GetCart (span)
│   │   └── CartRepository.SaveCart (span)
│   └── Response (200 OK)
```

## 📝 License

This project is licensed under the MIT License.
