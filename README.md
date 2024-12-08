# Product Management System

A scalable product management system built in Go, featuring image processing capabilities and caching mechanisms.

## Architecture

<div align="center">
  <img src="https://raw.githubusercontent.com/iSparshP/product-management-system/refs/heads/main/architecture.svg?sanitize=true" alt="Product Management System Architecture" width="800"/>
</div>


### Overview

The system follows a clean architecture pattern with the following main components:

- API Service (cmd/api)
- Image Processing Service (cmd/image-processor)
- Internal packages structured by domain

### Key Components

#### 1. API Layer

- REST API built with Gin framework
- Endpoints for product management (create, get, list)
- Structured routing with versioning (/api/v1)
- Health check endpoint included

#### 2. Core Services

- Product Management
- Asynchronous Image Processing
- Caching Layer

#### 3. Infrastructure

- PostgreSQL for persistent storage
- Redis for caching
- Kafka for async image processing
- AWS S3 for image storage
- Zap for structured logging

### Design Patterns

- Clean Architecture
- Repository Pattern
- Dependency Injection
- Event-Driven Architecture (using Kafka)

## Setup Instructions

### Prerequisites

- Go 1.x
- PostgreSQL
- Redis
- Kafka
- AWS Account (for S3)

### Environment Variables

Create a `.env` file with the following configuration:

```env
# Server
PORT=8080
LOG_LEVEL=info

# PostgreSQL
POSTGRES_HOST=localhost
POSTGRES_PORT=5432
POSTGRES_USER=user
POSTGRES_PASSWORD=password
POSTGRES_DB=productdb

# Redis
REDIS_ADDR=localhost:6379

# Kafka
KAFKA_BROKERS=localhost:9092

# AWS
AWS_ACCESS_KEY=your_access_key
AWS_SECRET_KEY=your_secret_key
AWS_REGION=your_region
AWS_S3_BUCKET=your_bucket
```

### Running the Services

1. Start API Service:

```bash
go run cmd/api/main.go
```

2. Start Image Processor:

```bash
go run cmd/image-processor/main.go
```

## Technical Assumptions

### 1. Data Storage

- Products are stored in PostgreSQL with UUID as primary keys
- Image URLs are stored as JSON arrays
- Timestamps are managed at the application level

### 2. Caching

- Product details are cached in Redis for 30 minutes
- Cache invalidation occurs on product updates
- Cache-aside pattern is implemented

### 3. Image Processing

- Asynchronous processing via Kafka
- Retry mechanism with max 3 attempts
- Failed tasks are sent to a Dead Letter Queue
- Compressed images are stored in S3

### 4. Security

- Authentication is required (implementation not shown in the code)
- User ID is required for product creation
- API versioning for backward compatibility

### 5. Performance

- Connection pooling for database
- Optimized image compression
- Distributed system ready

## API Endpoints

```
POST /api/v1/products - Create a new product
GET /api/v1/products/:id - Get product by ID
GET /api/v1/products - List products with filters
GET /health - Health check endpoint
```

## Error Handling

- Structured error responses
- Logging with correlation IDs
- Graceful shutdown handling
- Retry mechanisms for external services

## Future Considerations

- Rate limiting
- Circuit breakers
- API documentation (Swagger)
- Metrics and monitoring
- Container orchestration support
