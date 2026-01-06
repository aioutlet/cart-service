# Product Requirements Document (PRD)

## Cart Service - xshopai Platform

**Version:** 1.0  
**Last Updated:** November 10, 2025  
**Status:** Active Development  
**Owner:** xshopai Platform Team

---

## Table of Contents

1. [Product Overview](#1-product-overview)
2. [Technical Architecture](#2-technical-architecture)
3. [API Specifications](#3-api-specifications)
4. [Functional Requirements](#4-functional-requirements)
5. [Non-Functional Requirements](#5-non-functional-requirements)
6. [Dependencies](#6-dependencies)
7. [Testing Strategy](#7-testing-strategy)
8. [Deployment](#8-deployment)
9. [Monitoring & Alerts](#9-monitoring--alerts)

---

## 1. Product Overview

### 1.1 Product Vision

The Cart Service is a high-performance microservice within the xshopai e-commerce platform, responsible for managing shopping cart operations with support for both authenticated users and guest users. It provides lightning-fast cart operations using Dapr State Management with Redis as the backing store, enabling automatic expiration, concurrency control, and store-agnostic architecture.

### 1.2 Business Objectives

- **High Performance**: Sub-50ms response times for cart operations
- **Guest Support**: Enable shopping without registration
- **Seamless Migration**: Transfer guest carts to user accounts on login
- **Data Integrity**: Prevent race conditions with Dapr ETag-based concurrency control
- **Real-time Validation**: Validate products and inventory before cart operations
- **Scalability**: Support millions of active carts
- **Store Agnostic**: Dapr State Management enables easy migration to different backing stores

### 1.3 Success Metrics

| Metric                  | Target  | Current |
| ----------------------- | ------- | ------- |
| API Response Time (p95) | < 50ms  | TBD     |
| Service Uptime          | 99.9%   | TBD     |
| Cart Operation Success  | > 99.5% | TBD     |
| API Error Rate          | < 0.5%  | TBD     |
| Guest Cart Conversion   | > 30%   | TBD     |

### 1.4 Product Description

The Cart Service is a Go/Gin microservice that manages shopping cart operations using Dapr State Management with Redis as the backing store. It supports both authenticated users and guest users, with seamless cart transfer functionality when guests create accounts or log in. The use of Dapr provides store abstraction, built-in concurrency control, and operational benefits like automatic retries and observability.

### 1.5 Target Users

1. **Customers**: Authenticated users managing their shopping carts
2. **Guest Users**: Anonymous users shopping without registration
3. **Internal Services**: Order Service consuming cart data for checkout

### 1.6 Key Features

- ✅ Add, update, remove cart items
- ✅ Guest cart support with unique guest IDs
- ✅ Cart transfer from guest to authenticated user
- ✅ Real-time product validation
- ✅ Real-time inventory validation
- ✅ Automatic cart expiration (30 days for users, 3 days for guests)
- ✅ ETag-based optimistic concurrency control via Dapr
- ✅ Cart limits (max items, max quantity per item)
- ✅ Price calculation and totals
- ✅ Comprehensive health checks
- ✅ Store-agnostic architecture with Dapr State Management

---

## 2. Technical Architecture

### 2.1 Technology Stack

- **Runtime**: Go 1.21+
- **Framework**: Gin Web Framework
- **State Management**: Dapr State Management Building Block
- **Backing Store**: Redis 7.0+ (via Dapr)
- **Authentication**: JWT
- **Logging**: Uber Zap (structured logging)
- **Observability**: OpenTelemetry, Jaeger
- **Testing**: Testify, Go standard testing
- **Documentation**: Swagger/OpenAPI 3.0

### 2.2 Architecture Pattern

**Synchronous REST API with Dapr State Management**

```
┌─────────────────┐
│  Cart Service   │
│   (Gin/Go)      │
└────────┬────────┘
         │
         ├─► Dapr Sidecar (state management)
         │   └─► Redis (backing store)
         ├─► Product Service (product validation)
         └─► Inventory Service (stock validation)
```

### 2.2.1 Dapr State Management Benefits

**Abstraction**: Store-agnostic API allows switching backing stores without code changes

**Concurrency Control**: Built-in ETag-based optimistic locking prevents race conditions

**Operational Features**:

- Automatic retries with exponential backoff
- Circuit breaker for resilience
- Built-in observability and metrics
- State encryption at rest (optional)

**Developer Experience**:

- Simplified code without direct database client management
- Consistent API across all state stores
- Bulk operations support
- TTL/expiration built into state store metadata

### 2.3 Data Model

Cart Service stores cart data using Dapr State Management (with Redis backing store) with the following schema:

```json
{
  "userId": "user-123",
  "items": [
    {
      "productId": "507f1f77bcf86cd799439011",
      "productName": "Premium Cotton T-Shirt",
      "sku": "TS-BLK-001",
      "price": 29.99,
      "quantity": 2,
      "imageUrl": "https://cdn.xshopai.com/products/ts-blk-001.jpg",
      "category": "Clothing",
      "subtotal": 59.98,
      "addedAt": "2025-11-10T10:00:00Z"
    }
  ],
  "totalPrice": 59.98,
  "totalItems": 2,
  "createdAt": "2025-11-10T10:00:00Z",
  "updatedAt": "2025-11-10T10:05:00Z",
  "expiresAt": "2025-12-10T10:00:00Z"
}
```

**Key Design Decisions**:

- **Dapr State Management**: Store-agnostic abstraction layer with Redis backing store
- **ETag Concurrency**: Optimistic locking using Dapr ETags prevents race conditions
- **TTL Management**: Automatic expiration via Dapr metadata (30 days for users, 3 days for guests)
- **Denormalized Data**: Product details cached in cart for performance
- **Resilience**: Built-in retries and circuit breakers via Dapr

### 2.4 Dapr State Key Structure

**User Cart**: `cart:user:{userId}`  
**Guest Cart**: `cart:guest:{guestId}`

**State Metadata**:

```json
{
  "ttlInSeconds": 2592000, // 30 days for users, 259200 (3 days) for guests
  "contentType": "application/json"
}
```

**Concurrency Control**:

- Dapr automatically manages ETags for optimistic concurrency
- No separate lock keys needed (handled by Dapr State Management)

### 2.5 Inter-Service Communication

Cart Service communicates synchronously with:

**Product Service**:

- **Endpoint**: `GET /api/products/{productId}/exists`
- **Purpose**: Validate product exists and is active before adding to cart
- **Response**: `{ exists: true, isActive: true, price: 29.99 }`

**Inventory Service**:

- **Endpoint**: `GET /api/inventory/{productId}/availability`
- **Purpose**: Check stock availability before adding/updating cart items
- **Response**: `{ available: true, quantity: 50 }`

### 2.6 Environment Variables

#### 2.6.1 Server Configuration

| Variable               | Description            | Example      | Required | Default       |
| ---------------------- | ---------------------- | ------------ | -------- | ------------- |
| `ENVIRONMENT`          | Deployment environment | `production` | Yes      | `development` |
| `PORT`                 | HTTP server port       | `8085`       | No       | `8085`        |
| `SERVER_READ_TIMEOUT`  | HTTP read timeout      | `30s`        | No       | `30s`         |
| `SERVER_WRITE_TIMEOUT` | HTTP write timeout     | `30s`        | No       | `30s`         |

#### 2.6.2 Dapr Configuration

| Variable           | Description                     | Example        | Required | Default        |
| ------------------ | ------------------------------- | -------------- | -------- | -------------- |
| `DAPR_HTTP_PORT`   | Dapr HTTP sidecar port          | `3500`         | No       | `3500`         |
| `DAPR_GRPC_PORT`   | Dapr gRPC sidecar port          | `50001`        | No       | `50001`        |
| `DAPR_STATE_STORE` | Dapr state store component name | `statestore`   | Yes      | `statestore`   |
| `DAPR_APP_ID`      | Dapr application identifier     | `cart-service` | Yes      | `cart-service` |
| `DAPR_APP_PORT`    | Port where cart service listens | `8085`         | No       | `8085`         |

#### 2.6.3 Authentication Configuration

| Variable     | Description        | Example                               | Required | Default |
| ------------ | ------------------ | ------------------------------------- | -------- | ------- |
| `JWT_SECRET` | JWT signing secret | `8tDBDMcpxroHoHjXjk8xp/uAn8rzD4y8...` | Yes      | -       |

#### 2.6.4 Cart Configuration

| Variable                | Description                        | Example | Required | Default |
| ----------------------- | ---------------------------------- | ------- | -------- | ------- |
| `CART_DEFAULT_TTL`      | User cart TTL                      | `720h`  | No       | `720h`  |
| `CART_GUEST_TTL`        | Guest cart TTL                     | `72h`   | No       | `72h`   |
| `CART_MAX_ITEMS`        | Maximum items per cart             | `100`   | No       | `100`   |
| `CART_MAX_ITEM_QTY`     | Maximum quantity per item          | `10`    | No       | `10`    |
| `CART_CLEANUP_INTERVAL` | Cleanup interval for expired carts | `1h`    | No       | `1h`    |

#### 2.6.5 External Services Configuration

| Variable                | Description                | Example                 | Required | Default                 |
| ----------------------- | -------------------------- | ----------------------- | -------- | ----------------------- |
| `PRODUCT_SERVICE_URL`   | Product service base URL   | `http://localhost:8003` | Yes      | `http://localhost:8003` |
| `INVENTORY_SERVICE_URL` | Inventory service base URL | `http://localhost:8004` | Yes      | `http://localhost:8004` |

#### 2.6.6 CORS Configuration

| Variable               | Description                            | Example                                       | Required | Default |
| ---------------------- | -------------------------------------- | --------------------------------------------- | -------- | ------- |
| `CORS_ALLOWED_ORIGINS` | Allowed CORS origins (comma-separated) | `http://localhost:3000,http://localhost:8080` | No       | `*`     |

#### 2.6.7 Tracing Configuration

| Variable                  | Description                   | Example                             | Required | Default        |
| ------------------------- | ----------------------------- | ----------------------------------- | -------- | -------------- |
| `TRACING_ENABLED`         | Enable distributed tracing    | `true`                              | No       | `true`         |
| `TRACING_SERVICE_NAME`    | Service name for tracing      | `cart-service`                      | No       | `cart-service` |
| `TRACING_SERVICE_VERSION` | Service version               | `1.0.0`                             | No       | `1.0.0`        |
| `TRACING_JAEGER_ENDPOINT` | Jaeger collector endpoint     | `http://localhost:14268/api/traces` | No       | -              |
| `TRACING_SAMPLE_RATE`     | Trace sampling rate (0.0-1.0) | `1.0`                               | No       | `1.0`          |

#### 2.6.8 Dapr State Store Component Configuration

**Component YAML** (`.dapr/components/statestore.yaml`):

```yaml
apiVersion: dapr.io/v1alpha1
kind: Component
metadata:
  name: statestore
spec:
  type: state.redis
  version: v1
  metadata:
    - name: redisHost
      value: localhost:6379
    - name: redisPassword
      value: ''
    - name: actorStateStore
      value: 'false'
    - name: keyPrefix
      value: 'cart'
```

**State Operations**:

- **Save**: `POST http://localhost:3500/v1.0/state/statestore` with TTL metadata
- **Get**: `GET http://localhost:3500/v1.0/state/statestore/{key}`
- **Delete**: `DELETE http://localhost:3500/v1.0/state/statestore/{key}`
- **Bulk**: `POST http://localhost:3500/v1.0/state/statestore/bulk`

**Concurrency**:

- First write wins (optimistic locking with ETags)
- Last write wins (for non-critical updates)
- Configurable consistency: eventual or strong

---

## 3. API Specifications

### 3.1 Authentication

Cart endpoints support both authenticated and guest users:

- **Authenticated endpoints** require JWT token: `Authorization: Bearer <token>`
- **Guest endpoints** use guest ID in URL path

### 3.2 Authenticated Cart APIs

#### 3.2.1 Get User Cart

**Endpoint**: `GET /api/v1/cart`

**Description**: Retrieve the current user's shopping cart

**Authentication**: Required (JWT)

**Request Headers**:

```http
Authorization: Bearer <jwt_token>
X-Correlation-ID: req-abc-123 (optional)
```

**Response** (200 OK):

```json
{
  "success": true,
  "message": "Cart retrieved successfully",
  "data": {
    "userId": "user-123",
    "items": [
      {
        "productId": "507f1f77bcf86cd799439011",
        "productName": "Premium Cotton T-Shirt",
        "sku": "TS-BLK-001",
        "price": 29.99,
        "quantity": 2,
        "imageUrl": "https://cdn.xshopai.com/products/ts-blk-001.jpg",
        "category": "Clothing",
        "subtotal": 59.98,
        "addedAt": "2025-11-10T10:00:00Z"
      }
    ],
    "totalPrice": 59.98,
    "totalItems": 2,
    "createdAt": "2025-11-10T10:00:00Z",
    "updatedAt": "2025-11-10T10:05:00Z",
    "expiresAt": "2025-12-10T10:00:00Z"
  }
}
```

**Error Responses**:

- `401 Unauthorized` - Invalid or missing JWT token
- `404 Not Found` - Cart not found

#### 3.2.2 Add Item to Cart

**Endpoint**: `POST /api/v1/cart/items`

**Description**: Add a product to the user's cart

**Authentication**: Required (JWT)

**Request Headers**:

```http
Authorization: Bearer <jwt_token>
Content-Type: application/json
X-Correlation-ID: req-def-456
```

**Request Body**:

```json
{
  "productId": "507f1f77bcf86cd799439011",
  "quantity": 2
}
```

**Response** (201 Created):

```json
{
  "success": true,
  "message": "Item added to cart successfully",
  "data": {
    "userId": "user-123",
    "items": [
      {
        "productId": "507f1f77bcf86cd799439011",
        "productName": "Premium Cotton T-Shirt",
        "sku": "TS-BLK-001",
        "price": 29.99,
        "quantity": 2,
        "imageUrl": "https://cdn.xshopai.com/products/ts-blk-001.jpg",
        "category": "Clothing",
        "subtotal": 59.98,
        "addedAt": "2025-11-10T10:00:00Z"
      }
    ],
    "totalPrice": 59.98,
    "totalItems": 2,
    "createdAt": "2025-11-10T10:00:00Z",
    "updatedAt": "2025-11-10T10:00:00Z",
    "expiresAt": "2025-12-10T10:00:00Z"
  }
}
```

**Error Responses**:

- `400 Bad Request` - Invalid request (missing fields, invalid quantity)
- `401 Unauthorized` - Invalid JWT token
- `404 Not Found` - Product not found
- `409 Conflict` - Insufficient stock, max items exceeded, max quantity exceeded

#### 3.2.3 Update Cart Item Quantity

**Endpoint**: `PUT /api/v1/cart/items/{productId}`

**Description**: Update the quantity of an item in the cart

**Authentication**: Required (JWT)

**Request Headers**:

```http
Authorization: Bearer <jwt_token>
Content-Type: application/json
```

**Request Body**:

```json
{
  "quantity": 3
}
```

**Response** (200 OK):

```json
{
  "success": true,
  "message": "Cart item updated successfully",
  "data": {
    "userId": "user-123",
    "items": [
      {
        "productId": "507f1f77bcf86cd799439011",
        "productName": "Premium Cotton T-Shirt",
        "sku": "TS-BLK-001",
        "price": 29.99,
        "quantity": 3,
        "imageUrl": "https://cdn.xshopai.com/products/ts-blk-001.jpg",
        "category": "Clothing",
        "subtotal": 89.97,
        "addedAt": "2025-11-10T10:00:00Z"
      }
    ],
    "totalPrice": 89.97,
    "totalItems": 3,
    "createdAt": "2025-11-10T10:00:00Z",
    "updatedAt": "2025-11-10T10:10:00Z",
    "expiresAt": "2025-12-10T10:00:00Z"
  }
}
```

**Note**: Setting quantity to 0 removes the item from cart

**Error Responses**:

- `400 Bad Request` - Invalid quantity
- `401 Unauthorized` - Invalid JWT token
- `404 Not Found` - Item not found in cart
- `409 Conflict` - Max quantity exceeded

#### 3.2.4 Remove Item from Cart

**Endpoint**: `DELETE /api/v1/cart/items/{productId}`

**Description**: Remove an item from the cart

**Authentication**: Required (JWT)

**Request Headers**:

```http
Authorization: Bearer <jwt_token>
```

**Response** (200 OK):

```json
{
  "success": true,
  "message": "Item removed from cart successfully",
  "data": {
    "userId": "user-123",
    "items": [],
    "totalPrice": 0,
    "totalItems": 0,
    "createdAt": "2025-11-10T10:00:00Z",
    "updatedAt": "2025-11-10T10:15:00Z",
    "expiresAt": "2025-12-10T10:00:00Z"
  }
}
```

**Error Responses**:

- `401 Unauthorized` - Invalid JWT token
- `404 Not Found` - Item not found in cart

#### 3.2.5 Clear Cart

**Endpoint**: `DELETE /api/v1/cart`

**Description**: Remove all items from the cart

**Authentication**: Required (JWT)

**Request Headers**:

```http
Authorization: Bearer <jwt_token>
```

**Response** (200 OK):

```json
{
  "success": true,
  "message": "Cart cleared successfully",
  "data": {
    "userId": "user-123",
    "items": [],
    "totalPrice": 0,
    "totalItems": 0,
    "createdAt": "2025-11-10T10:00:00Z",
    "updatedAt": "2025-11-10T10:20:00Z",
    "expiresAt": "2025-12-10T10:00:00Z"
  }
}
```

**Error Responses**:

- `401 Unauthorized` - Invalid JWT token

#### 3.2.6 Transfer Guest Cart to User

**Endpoint**: `POST /api/v1/cart/transfer`

**Description**: Transfer guest cart items to authenticated user's cart (used after login/registration)

**Authentication**: Required (JWT)

**Request Headers**:

```http
Authorization: Bearer <jwt_token>
Content-Type: application/json
```

**Request Body**:

```json
{
  "guestId": "guest-xyz-789"
}
```

**Response** (200 OK):

```json
{
  "success": true,
  "message": "Guest cart transferred successfully",
  "data": {
    "userId": "user-123",
    "items": [
      {
        "productId": "507f1f77bcf86cd799439011",
        "productName": "Premium Cotton T-Shirt",
        "sku": "TS-BLK-001",
        "price": 29.99,
        "quantity": 2,
        "imageUrl": "https://cdn.xshopai.com/products/ts-blk-001.jpg",
        "category": "Clothing",
        "subtotal": 59.98,
        "addedAt": "2025-11-10T09:00:00Z"
      }
    ],
    "totalPrice": 59.98,
    "totalItems": 2,
    "createdAt": "2025-11-10T10:30:00Z",
    "updatedAt": "2025-11-10T10:30:00Z",
    "expiresAt": "2025-12-10T10:30:00Z"
  }
}
```

**Transfer Logic**:

- If user cart is empty, guest cart items are transferred directly
- If user cart has items, guest cart items are merged:
  - Duplicate products: quantities are added together (respecting max quantity limit)
  - New products: added to user cart (respecting max items limit)
- Guest cart is deleted after successful transfer

**Error Responses**:

- `400 Bad Request` - Guest ID missing or invalid
- `401 Unauthorized` - Invalid JWT token
- `404 Not Found` - Guest cart not found
- `409 Conflict` - Max items or quantity limit exceeded during merge

---

### 3.3 Guest Cart APIs

#### 3.3.1 Get Guest Cart

**Endpoint**: `GET /api/v1/guest/cart/{guestId}`

**Description**: Retrieve a guest user's shopping cart

**Authentication**: None

**Path Parameters**:

- `guestId` (required): Unique identifier for guest user (generated client-side)

**Response** (200 OK):

```json
{
  "success": true,
  "message": "Cart retrieved successfully",
  "data": {
    "userId": "guest-xyz-789",
    "items": [
      {
        "productId": "507f1f77bcf86cd799439011",
        "productName": "Premium Cotton T-Shirt",
        "sku": "TS-BLK-001",
        "price": 29.99,
        "quantity": 1,
        "imageUrl": "https://cdn.xshopai.com/products/ts-blk-001.jpg",
        "category": "Clothing",
        "subtotal": 29.99,
        "addedAt": "2025-11-10T09:00:00Z"
      }
    ],
    "totalPrice": 29.99,
    "totalItems": 1,
    "createdAt": "2025-11-10T09:00:00Z",
    "updatedAt": "2025-11-10T09:00:00Z",
    "expiresAt": "2025-11-13T09:00:00Z"
  }
}
```

**Error Responses**:

- `404 Not Found` - Guest cart not found

#### 3.3.2 Add Item to Guest Cart

**Endpoint**: `POST /api/v1/guest/cart/{guestId}/items`

**Description**: Add a product to a guest cart

**Authentication**: None

**Path Parameters**:

- `guestId` (required): Unique identifier for guest user

**Request Body**:

```json
{
  "productId": "507f1f77bcf86cd799439011",
  "quantity": 1
}
```

**Response** (201 Created):

```json
{
  "success": true,
  "message": "Item added to cart successfully",
  "data": {
    "userId": "guest-xyz-789",
    "items": [
      {
        "productId": "507f1f77bcf86cd799439011",
        "productName": "Premium Cotton T-Shirt",
        "sku": "TS-BLK-001",
        "price": 29.99,
        "quantity": 1,
        "imageUrl": "https://cdn.xshopai.com/products/ts-blk-001.jpg",
        "category": "Clothing",
        "subtotal": 29.99,
        "addedAt": "2025-11-10T09:00:00Z"
      }
    ],
    "totalPrice": 29.99,
    "totalItems": 1,
    "createdAt": "2025-11-10T09:00:00Z",
    "updatedAt": "2025-11-10T09:00:00Z",
    "expiresAt": "2025-11-13T09:00:00Z"
  }
}
```

**Error Responses**:

- `400 Bad Request` - Invalid request
- `404 Not Found` - Product not found
- `409 Conflict` - Insufficient stock, max items exceeded

#### 3.3.3 Update Guest Cart Item Quantity

**Endpoint**: `PUT /api/v1/guest/cart/{guestId}/items/{productId}`

**Description**: Update quantity of an item in guest cart

**Authentication**: None

**Path Parameters**:

- `guestId` (required): Unique identifier for guest user
- `productId` (required): Product ID to update

**Request Body**:

```json
{
  "quantity": 2
}
```

**Response** (200 OK):

```json
{
  "success": true,
  "message": "Cart item updated successfully",
  "data": {
    "userId": "guest-xyz-789",
    "items": [
      {
        "productId": "507f1f77bcf86cd799439011",
        "productName": "Premium Cotton T-Shirt",
        "sku": "TS-BLK-001",
        "price": 29.99,
        "quantity": 2,
        "imageUrl": "https://cdn.xshopai.com/products/ts-blk-001.jpg",
        "category": "Clothing",
        "subtotal": 59.98,
        "addedAt": "2025-11-10T09:00:00Z"
      }
    ],
    "totalPrice": 59.98,
    "totalItems": 2,
    "createdAt": "2025-11-10T09:00:00Z",
    "updatedAt": "2025-11-10T09:05:00Z",
    "expiresAt": "2025-11-13T09:00:00Z"
  }
}
```

**Error Responses**:

- `400 Bad Request` - Invalid quantity
- `404 Not Found` - Item not found in cart
- `409 Conflict` - Max quantity exceeded

#### 3.3.4 Remove Item from Guest Cart

**Endpoint**: `DELETE /api/v1/guest/cart/{guestId}/items/{productId}`

**Description**: Remove an item from guest cart

**Authentication**: None

**Path Parameters**:

- `guestId` (required): Unique identifier for guest user
- `productId` (required): Product ID to remove

**Response** (200 OK):

```json
{
  "success": true,
  "message": "Item removed from cart successfully",
  "data": {
    "userId": "guest-xyz-789",
    "items": [],
    "totalPrice": 0,
    "totalItems": 0,
    "createdAt": "2025-11-10T09:00:00Z",
    "updatedAt": "2025-11-10T09:10:00Z",
    "expiresAt": "2025-11-13T09:00:00Z"
  }
}
```

**Error Responses**:

- `404 Not Found` - Item not found in cart

#### 3.3.5 Clear Guest Cart

**Endpoint**: `DELETE /api/v1/guest/cart/{guestId}`

**Description**: Remove all items from guest cart

**Authentication**: None

**Path Parameters**:

- `guestId` (required): Unique identifier for guest user

**Response** (200 OK):

```json
{
  "success": true,
  "message": "Cart cleared successfully",
  "data": {
    "userId": "guest-xyz-789",
    "items": [],
    "totalPrice": 0,
    "totalItems": 0,
    "createdAt": "2025-11-10T09:00:00Z",
    "updatedAt": "2025-11-10T09:15:00Z",
    "expiresAt": "2025-11-13T09:00:00Z"
  }
}
```

---

### 3.4 Health Check APIs

#### 3.4.1 Liveness Probe

**Endpoint**: `GET /health`

**Description**: Kubernetes liveness probe - checks if service is running

**Authentication**: None

**Response** (200 OK):

```json
{
  "status": "healthy",
  "service": "cart-service",
  "timestamp": "2025-11-10T10:00:00Z"
}
```

#### 3.4.2 Readiness Probe

**Endpoint**: `GET /health/ready`

**Description**: Kubernetes readiness probe - checks if service can accept traffic

**Authentication**: None

**Response** (200 OK):

```json
{
  "status": "ready",
  "service": "cart-service",
  "timestamp": "2025-11-10T10:00:00Z",
  "dependencies": {
    "daprStateStore": "healthy",
    "productService": "healthy",
    "inventoryService": "healthy"
  }
}
```

**Response** (503 Service Unavailable):

```json
{
  "status": "not-ready",
  "service": "cart-service",
  "timestamp": "2025-11-10T10:00:00Z",
  "dependencies": {
    "daprStateStore": "unhealthy",
    "productService": "healthy",
    "inventoryService": "healthy"
  },
  "reason": "Dapr state store unavailable"
}
```

---

### 3.5 Error Code Catalog

#### 3.5.1 Client Error Codes (4xx)

| Error Code              | HTTP Status | Description                    | Client Action                        |
| ----------------------- | ----------- | ------------------------------ | ------------------------------------ |
| `VALIDATION_ERROR`      | 400         | Request validation failed      | Review request body, check API docs  |
| `INVALID_QUANTITY`      | 400         | Quantity is invalid            | Provide valid positive quantity      |
| `INVALID_GUEST_ID`      | 400         | Guest ID format invalid        | Use valid UUID format for guest ID   |
| `UNAUTHORIZED`          | 401         | Authentication failed          | Re-authenticate, obtain fresh token  |
| `TOKEN_EXPIRED`         | 401         | JWT token expired              | Re-authenticate with credentials     |
| `CART_NOT_FOUND`        | 404         | Cart does not exist            | Verify cart exists or add first item |
| `ITEM_NOT_FOUND`        | 404         | Item not found in cart         | Verify product ID in cart            |
| `PRODUCT_NOT_FOUND`     | 404         | Product does not exist         | Verify product ID is valid           |
| `INSUFFICIENT_STOCK`    | 409         | Not enough inventory available | Reduce quantity or try later         |
| `MAX_ITEMS_EXCEEDED`    | 409         | Cart item limit reached        | Remove items before adding more      |
| `MAX_QUANTITY_EXCEEDED` | 409         | Item quantity limit reached    | Reduce quantity to max allowed       |
| `CART_EXPIRED`          | 410         | Cart has expired               | Create new cart                      |
| `TOO_MANY_REQUESTS`     | 429         | Rate limit exceeded            | Implement backoff, reduce frequency  |

#### 3.5.2 Server Error Codes (5xx)

| Error Code                | HTTP Status | Description                       | Client Action                       |
| ------------------------- | ----------- | --------------------------------- | ----------------------------------- |
| `INTERNAL_SERVER_ERROR`   | 500         | Unexpected server error           | Retry with backoff, contact support |
| `STATE_STORE_ERROR`       | 500         | Dapr state store operation failed | Retry request                       |
| `CONCURRENCY_ERROR`       | 409         | ETag mismatch (concurrent update) | Retry with latest state             |
| `PRODUCT_SERVICE_ERROR`   | 500         | Product service unavailable       | Retry request                       |
| `INVENTORY_SERVICE_ERROR` | 500         | Inventory service unavailable     | Retry request                       |
| `SERVICE_UNAVAILABLE`     | 503         | Service temporarily unavailable   | Wait and retry                      |
| `GATEWAY_TIMEOUT`         | 504         | Request timeout                   | Retry with exponential backoff      |

#### 3.5.3 Error Response Format

```json
{
  "success": false,
  "message": "Human-readable error message",
  "error": "ERROR_CODE",
  "correlationId": "req-abc-123"
}
```

---

## 4. Functional Requirements

### 4.1 Cart Management

**FR-1**: System SHALL allow authenticated users to add products to their cart  
**FR-2**: System SHALL allow guest users to add products to cart using guest ID  
**FR-3**: System SHALL validate product existence before adding to cart  
**FR-4**: System SHALL validate inventory availability before adding to cart  
**FR-5**: System SHALL update item quantity if product already exists in cart  
**FR-6**: System SHALL allow users to update item quantities in cart  
**FR-7**: System SHALL allow users to remove items from cart  
**FR-8**: System SHALL allow users to clear entire cart  
**FR-9**: System SHALL calculate and return cart totals (price, item count)

### 4.2 Guest Cart Support

**FR-10**: System SHALL support guest carts without authentication  
**FR-11**: System SHALL expire guest carts after 72 hours  
**FR-12**: System SHALL allow transfer of guest cart to authenticated user  
**FR-13**: System SHALL merge guest cart with existing user cart on transfer  
**FR-14**: System SHALL delete guest cart after successful transfer

### 4.3 Cart Limits

**FR-15**: System SHALL enforce maximum 100 items per cart  
**FR-16**: System SHALL enforce maximum 10 quantity per item  
**FR-17**: System SHALL prevent adding items when limits are exceeded

### 4.4 Data Management

**FR-18**: System SHALL expire user carts after 30 days of inactivity  
**FR-19**: System SHALL cache product details in cart for performance  
**FR-20**: System SHALL recalculate totals after any cart modification

### 4.5 Concurrency Control

**FR-21**: System SHALL use Dapr ETag-based optimistic concurrency control to prevent race conditions  
**FR-22**: System SHALL retrieve ETag when reading cart state  
**FR-23**: System SHALL include ETag when saving cart state for concurrency validation  
**FR-24**: System SHALL retry with latest state on ETag mismatch (concurrent modification detected)

---

## 5. Non-Functional Requirements

### 5.1 Performance

**NFR-1**: Cart read operations SHALL complete in < 50ms (p95)  
**NFR-2**: Cart write operations SHALL complete in < 100ms (p95)  
**NFR-3**: Product validation calls SHALL complete in < 100ms (p95)  
**NFR-4**: System SHALL support 10,000 concurrent users

### 5.2 Availability

**NFR-5**: Service SHALL maintain 99.9% uptime  
**NFR-6**: Service SHALL handle Redis failures gracefully  
**NFR-7**: Service SHALL implement health checks for monitoring

### 5.3 Scalability

**NFR-8**: System SHALL support horizontal scaling  
**NFR-9**: System SHALL handle 1M+ active carts  
**NFR-10**: System SHALL support 1000 requests/second per instance

### 5.4 Security

**NFR-11**: System SHALL validate JWT tokens for authenticated endpoints  
**NFR-12**: System SHALL prevent unauthorized access to user carts  
**NFR-13**: System SHALL sanitize all user inputs  
**NFR-14**: System SHALL log security events

### 5.5 Observability

**NFR-15**: System SHALL implement structured logging  
**NFR-16**: System SHALL support distributed tracing with OpenTelemetry  
**NFR-17**: System SHALL expose metrics for monitoring  
**NFR-18**: System SHALL include correlation IDs in all logs

### 5.6 Reliability

**NFR-19**: System SHALL implement retry logic for external service calls  
**NFR-20**: System SHALL handle partial failures gracefully  
**NFR-21**: System SHALL validate data before state operations  
**NFR-22**: System SHALL leverage Dapr's built-in retry and circuit breaker policies  
**NFR-23**: System SHALL handle ETag conflicts with automatic retry logic

---

## 6. Dependencies

### 6.1 External Services

**Product Service**:

- Purpose: Product validation, price lookup
- Endpoint: `GET /api/products/{productId}/exists`
- Criticality: High (cart operations fail without it)

**Inventory Service**:

- Purpose: Stock availability validation
- Endpoint: `GET /api/inventory/{productId}/availability`
- Criticality: High (cart operations fail without it)

### 6.2 Infrastructure Dependencies

**Dapr Runtime**:

- Version: 1.11+
- Purpose: State management abstraction, sidecar communication
- Criticality: Critical (service cannot function without it)

**Redis** (via Dapr):

- Version: 7.0+
- Purpose: Backing store for Dapr state management
- Criticality: Critical (state operations fail without it)

**JWT Authentication**:

- Purpose: User authentication for protected endpoints
- Criticality: High (authenticated operations fail without it)

### 6.3 Development Dependencies

- Go 1.21+
- Gin Web Framework
- Dapr Go SDK v1.8+
- Testify (testing)
- Uber Zap (logging)
- OpenTelemetry Go SDK

---

## 7. Testing Strategy

### 7.1 Unit Testing

**Scope**: Test individual functions and methods

**Coverage Target**: 80%+

**Key Test Areas**:

- Cart model logic (add, update, remove items)
- Total calculation
- Cart expiration logic
- Validation functions

**Tools**: Go standard testing, Testify

### 7.2 Integration Testing

**Scope**: Test service integration with Dapr state store and external services

**Key Test Areas**:

- Dapr state operations (save, get, delete, bulk)
- ETag-based concurrency control
- Product service integration
- Inventory service integration
- TTL expiration behavior

**Tools**: Go testing with Dapr sidecar in self-hosted mode, HTTP mocks

### 7.3 API Testing

**Scope**: Test HTTP endpoints end-to-end

**Key Test Areas**:

- All cart endpoints
- Guest cart endpoints
- Authentication flow
- Error handling

**Tools**: Go HTTP testing, Testify HTTP assertions

### 7.4 Performance Testing

**Scope**: Validate performance requirements

**Key Metrics**:

- Response time (p50, p95, p99)
- Throughput (requests/second)
- Concurrent users
- Redis operation latency

**Tools**: Apache JMeter, k6

---

## 8. Deployment

### 8.1 Environment Configuration

**Development**:

- Local Redis instance
- Mock external services
- Debug logging enabled
- Tracing sample rate: 1.0

**Staging**:

- Shared Redis cluster
- Staging external services
- Info logging
- Tracing sample rate: 0.5

**Production**:

- Production Redis cluster (HA setup)
- Production external services
- Error logging
- Tracing sample rate: 0.1

### 8.2 Container Configuration

**Dockerfile**:

- Base image: `golang:1.21-alpine`
- Multi-stage build for minimal image size
- Non-root user for security
- Health check configured

**Resource Requirements**:

- CPU: 500m (request), 1000m (limit)
- Memory: 256Mi (request), 512Mi (limit)

### 8.3 Kubernetes Deployment

**Deployment Strategy**: Rolling update

**Replicas**: 3 (minimum)

**Health Checks**:

- Liveness: `GET /health` (interval: 30s, timeout: 3s)
- Readiness: `GET /health/ready` (interval: 10s, timeout: 5s)

**Auto-scaling**:

- Min replicas: 3
- Max replicas: 10
- CPU threshold: 70%
- Memory threshold: 80%

---

## 9. Monitoring & Alerts

### 9.1 Metrics

**Service Metrics**:

- Request rate (requests/second)
- Response time (p50, p95, p99)
- Error rate (%)
- Active carts count

**Redis Metrics**:

- Connection pool usage
- Operation latency
- Cache hit rate
- Memory usage

**Business Metrics**:

- Cart creation rate
- Cart abandonment rate
- Average cart value
- Guest cart conversion rate

### 9.2 Alerts

**Critical Alerts**:

- Service unavailable (5xx errors > 5% for 5 minutes)
- Redis connection failure
- Response time > 500ms (p95) for 10 minutes
- Error rate > 1% for 5 minutes

**Warning Alerts**:

- Response time > 200ms (p95) for 15 minutes
- Redis memory usage > 80%
- Connection pool exhaustion
- External service errors > 5%

### 9.3 Dashboards

**Service Dashboard**:

- Request rate and response time
- Error rate by endpoint
- Active connections
- Service health status

**Business Dashboard**:

- Active carts
- Cart conversion funnel
- Average cart value
- Top products in carts

---

## Approval & Sign-off

| Role             | Name | Signature | Date |
| ---------------- | ---- | --------- | ---- |
| Product Manager  | TBD  |           |      |
| Engineering Lead | TBD  |           |      |
| QA Lead          | TBD  |           |      |
| DevOps Lead      | TBD  |           |      |

---

## Revision History

| Version | Date       | Author | Changes              |
| ------- | ---------- | ------ | -------------------- |
| 1.0     | 2025-11-10 | Team   | Initial PRD creation |

---

## Appendix

### Glossary

- **Cart**: A temporary collection of products a user intends to purchase
- **Guest User**: An unauthenticated user browsing the platform
- **TTL**: Time-To-Live, expiration time for cached data
- **Distributed Lock**: Mechanism to prevent concurrent modifications

### References

- Product Service PRD: `services/product-service/docs/PRD.md`
- Cart Service README: `services/cart-service/README.md`
- Platform Architecture: `docs/PLATFORM_ARCHITECTURE.md`
