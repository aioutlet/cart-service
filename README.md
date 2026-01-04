# 🛒 Cart Service

Shopping cart microservice for xshop.ai - high-performance, Redis-backed cart management with instant hot reload, guest support, and distributed locking.

## 🚀 Quick Start

### Prerequisites

- **Java** 21+ ([Download](https://adoptium.net/))
- **Maven** 3.9+ ([Install Guide](https://maven.apache.org/install.html)) or use included wrapper
- **Redis** 7.0+ ([Install Guide](https://redis.io/docs/getting-started/))
- **Dapr CLI** 1.16+ ([Install Guide](https://docs.dapr.io/getting-started/install-dapr-cli/))

### Setup

**1. Start Redis**
```bash
# Using Docker (recommended)
docker run -d -p 6379:6379 --name redis \
  redis:7-alpine redis-server --requirepass redis_dev_pass_123

# Or install Redis locally
```

**2. Clone & Build**
```bash
git clone https://github.com/xshopai/cart-service.git
cd cart-service
./mvnw clean install
```

**3. Run with Dapr (Recommended)**
```bash
# PowerShell
.\run.ps1

# Bash
./run.sh
```

**4. Or Run Without Dapr (Development)**
```bash
./mvnw quarkus:dev
```

**5. Verify**
```bash
# Check health
curl http://localhost:1008/health

# Swagger UI
Open http://localhost:1008/swagger-ui

# Quarkus Dev UI
Open http://localhost:1008/q/dev
```

### Common Commands

```bash
# Run tests
./mvnw test

# Run with coverage
./mvnw test jacoco:report

# Continuous testing (in dev mode, press 'r')
./mvnw quarkus:dev

# Build native image
./mvnw package -Pnative

# View coverage report
open target/site/jacoco/index.html
```

## 📚 Documentation

| Document | Description |
|----------|-------------|
| [📖 Developer Guide](docs/DEVELOPER_GUIDE.md) | Local setup, debugging, daily workflows |
| [📘 Technical Reference](docs/TECHNICAL.md) | Architecture, security, monitoring |
| [🤝 Contributing](docs/CONTRIBUTING.md) | Contribution guidelines and workflow |
| [🔐 Secrets Management](.dapr/SECRETS_README.md) | Dapr secret store configuration |

**API Documentation**: Swagger UI at `/swagger-ui` with interactive testing.

## ⚙️ Configuration

The Cart Service follows clean architecture principles with clear separation of concerns:

```text
cart-service/
├── src/
│   ├── main/
│   │   ├── java/com/xshopai/cartservice/
│   │   │   ├── client/          # External service clients (Dapr)
│   │   │   ├── dto/             # Data transfer objects
│   │   │   ├── exception/       # Custom exceptions
│   │   │   ├── model/           # Domain models
│   │   │   ├── repository/      # Data access layer (Redis)
│   │   │   ├── resource/        # REST endpoints (JAX-RS)
│   │   │   └── service/         # Business logic layer
│   │   └── resources/
│   │       └── application.properties
│   └── test/                    # Unit and integration tests
├── pom.xml                      # Maven dependencies
├── run.ps1                      # PowerShell run script with Dapr
└── run.sh                       # Bash run script with Dapr
```

## 🛠️ Technology Stack

- **Language**: Java 21
- **Framework**: Quarkus 3.6.4
- **Database**: Redis 7+ (with Lettuce client)
- **Authentication**: SmallRye JWT
- **Service Mesh**: Dapr 1.16+
- **REST**: JAX-RS with RESTEasy Reactive
- **Validation**: Jakarta Bean Validation
- **Documentation**: SmallRye OpenAPI 3.0
- **Testing**: JUnit 5, REST Assured
- **Build Tool**: Maven 3.9+
- **Logging**: JBoss Logging with JSON output

## 📋 Prerequisites

- Java 21 or higher
- Maven 3.9+ or use the included Maven wrapper (./mvnw)
- Redis 7.0 or higher
- Dapr CLI 1.16+ (for service mesh integration)
- Docker & Docker Compose (optional)

## 🚀 Quick Start

### Local Development

1. **Clone the repository**:

   ```bash
   git clone <repository-url>
   cd cart-service
   ```

2. **Start Redis**:

   ```bash
   docker run -d -p 6379:6379 --name redis redis:7-alpine redis-server --requirepass redis_dev_pass_123
   ```

3. **Run with Dapr** (Recommended):

   **PowerShell:**
   ```powershell
   .\run.ps1
   ```

   **Bash:**
   ```bash
   ./run.sh
   ```

4. **Or run without Dapr** (Development only):

   ```bash
   ./mvnw quarkus:dev
   ```

The service will be available at:
- **Application**: http://localhost:1008
- **Swagger UI**: http://localhost:1008/swagger-ui
- **Health**: http://localhost:1008/health
- **Dapr HTTP**: http://localhost:3508 (when using Dapr)

### Development Mode Features

Quarkus dev mode provides:
- ✨ **Instant hot reload** (< 1 second)
- 🔍 **Live coding** - changes reflect immediately
- 📊 **Dev UI** - http://localhost:1008/q/dev
- 🐛 **Remote debugging** ready on port 5005

### Docker Deployment

Build native image:
```bash
./mvnw package -Pnative
docker build -f src/main/docker/Dockerfile.native -t cart-service:native .
```

Run with Docker Compose:
```bash
docker-compose up cart-service
```

## 📚 API Documentation

### Base URL

```text
http://localhost:1008/api/v1
```

### Swagger UI

Access interactive API documentation at:
```text
http://localhost:1008/swagger-ui
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

## 🧪 Testing

### Run All Tests

```bash
./mvnw test
```

### Run Tests with Coverage

```bash
./mvnw test jacoco:report
```

### Run in Continuous Testing Mode

```bash
./mvnw quarkus:dev
# Then press 'r' to run tests
```

### View Coverage Report

```bash
open target/site/jacoco/index.html  # Mac/Linux
start target/site/jacoco/index.html # Windows
```

## 🔧 Configuration

Configuration is managed through `src/main/resources/application.properties`:

### Key Configuration Properties
## 🔍 Observability

### Distributed Tracing with Dapr

The cart service leverages Dapr's built-in distributed tracing capabilities:

- **Automatic Trace Propagation**: Dapr handles trace context across service calls
- **Service-to-Service Tracking**: Full visibility into product/inventory service calls
- **Correlation IDs**: Automatic correlation ID generation and propagation
- **OpenTelemetry Compatible**: Works with Jaeger, Zipkin, or any OTLP collector

### Health Checks

Quarkus provides comprehensive health checks:

```bash
# Liveness probe
curl http://localhost:1008/health/live

# Readiness probe  
curl http://localhost:1008/health/ready

# Full health check
curl http://localhost:1008/health
```

Health checks include:
- Redis connectivity
- Service readiness
- Custom business logic checks

### Metrics

Quarkus exposes Micrometer metrics:

```bash
# Prometheus format metrics
curl http://localhost:1008/q/metrics
```

Available metrics:
- HTTP request counts and durations
- Redis connection pool stats
- JVM memory and GC metrics
- Custom business metrics

### Logging

Structured JSON logging in production:

```json
{
  "timestamp": "2025-11-17T22:00:10Z",
  "level": "INFO",
  "logger": "com.xshopai.cartservice.service.CartService",
  "message": "Item added to cart",
  "userId": "379feb6c-7625-4a43-87c6-623c6665446d",
  "productId": "691b0a616f5dbbaca2b730ba",
  "quantity": 1
}
```

### Running with Observability Stack

Start Jaeger and monitoring:

```bash
docker-compose -f scripts/docker-compose/docker-compose.infrastructure.yml up jaeger grafana prometheus
```

Access UIs:
- **Jaeger**: http://localhost:16686
- **Grafana**: http://localhost:3000
- **Prometheus**: http://localhost:9090

## 🚀 Performance

### Development Experience

- **Hot Reload**: < 1 second (vs 10-15 seconds with Go)
- **First Startup**: ~5-10 seconds
- **Subsequent Changes**: Sub-second feedback

### Runtime Performance

- **JIT Optimization**: Adaptive optimization improves over time
- **Native Image**: Optional GraalVM native compilation for 0.02s startup
- **Memory**: ~100MB RSS in JVM mode, ~20MB in native mode
- **Throughput**: 10,000+ requests/second (JVM), 15,000+ (native)

## 🔐 Security

### Dapr Secret Management

The cart service uses **Dapr's Secret Management building block** for secure storage and retrieval of sensitive configuration:

- **Secret Store Support**: Local file, Azure Key Vault, AWS Secrets Manager, GCP Secret Manager, HashiCorp Vault
- **Component Integration**: Dapr components reference secrets via `secretKeyRef`
- **SDK Access**: `DaprSecretManager` provides programmatic access to secrets
- **Production Ready**: Easy migration from local to cloud-based secret stores

#### Configuration

Secrets are configured in `.dapr/secrets.json` (local) or cloud secret stores (production):

```json
{
  "redis": {
    "password": "redis_dev_pass_123"
  },
  "jwt": {
    "secret": "your-jwt-secret-key"
  }
}
```

Components reference secrets using `secretKeyRef`:

```yaml
metadata:
  - name: redisPassword
    secretKeyRef:
      name: redis:password
      key: redis:password
auth:
  secretStore: local-secret-store
```

Access secrets in code via `DaprSecretManager`:

```java
@Inject
DaprSecretManager secretManager;

String redisPassword = secretManager.getSecret("redis:password");
String jwtSecret = secretManager.getJwtSecret();
```

📖 **See `.dapr/SECRETS_README.md` for complete secret management documentation**

### Authentication

JWT-based authentication using SmallRye JWT:
- User ID extracted from `sub` or `userId` claims
- Role-based access control ready
- Token validation via public key

### Redis Security

- Password authentication via Dapr secrets
- Connection encryption via TLS (production)
- Key prefixing to prevent namespace collisions

### Input Validation

Jakarta Bean Validation ensures:
- Required fields present
- Valid quantity ranges (1-10 per item)
- Maximum cart size enforcement (100 items)

## 📦 Dependencies

### Core Dependencies

```xml
<!-- Quarkus Core -->
<dependency>
  <groupId>io.quarkus</groupId>
  <artifactId>quarkus-arc</artifactId>
</dependency>

<!-- REST -->
<dependency>
  <groupId>io.quarkus</groupId>
  <artifactId>quarkus-resteasy-reactive-jackson</artifactId>
</dependency>

<!-- Redis -->
<dependency>
  <groupId>io.quarkus</groupId>
  <artifactId>quarkus-redis-client</artifactId>
</dependency>

<!-- JWT -->
<dependency>
  <groupId>io.quarkus</groupId>
  <artifactId>quarkus-smallrye-jwt</artifactId>
</dependency>

<!-- REST Client (Dapr) -->
<dependency>
  <groupId>io.quarkus</groupId>
  <artifactId>quarkus-rest-client-reactive-jackson</artifactId>
</dependency>
```

### Why Quarkus?

**Developer Experience:**
- ⚡ Instant hot reload eliminates slow build cycles
- 🔍 Live coding - see changes immediately
- 🎯 Dev UI for easy testing and debugging
- 📊 Built-in observability and metrics

**Production Ready:**
- 🚀 High performance with low memory footprint
- ☁️ Cloud-native design
- 🔄 Reactive and imperative models
- 📦 Optional native compilation with GraalVM

**Enterprise Features:**
- 🔐 Comprehensive security (JWT, RBAC, encryption)
- 📈 Production-grade monitoring and tracing
- 🏥 Advanced health checks
- 🔧 Extensive configuration optionsCING_SAMPLE_RATE=1.0
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
