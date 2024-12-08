# Product Management System

A scalable product management system built in Go, featuring image processing capabilities and caching mechanisms.

## Architecture

<?xml version="1.0" encoding="UTF-8"?>
<svg width="800" height="600" xmlns="http://www.w3.org/2000/svg">
    <!-- Background -->
    <rect width="100%" height="100%" fill="#f8f9fa"/>
    
    <!-- Title -->
    <text x="400" y="40" font-family="Arial" font-size="24" text-anchor="middle" font-weight="bold">
        Product Management System Architecture
    </text>

    <!-- API Service Container -->
    <rect x="50" y="80" width="200" height="160" rx="10" fill="#6cb2eb" opacity="0.8"/>
    <text x="150" y="110" font-family="Arial" font-size="16" text-anchor="middle" fill="#ffffff" font-weight="bold">
        API Service
    </text>
    <text x="150" y="140" font-family="Arial" font-size="12" text-anchor="middle" fill="#ffffff">
        - Gin Framework
    </text>
    <text x="150" y="160" font-family="Arial" font-size="12" text-anchor="middle" fill="#ffffff">
        - REST Endpoints
    </text>
    <text x="150" y="180" font-family="Arial" font-size="12" text-anchor="middle" fill="#ffffff">
        - Middleware
    </text>
    <text x="150" y="200" font-family="Arial" font-size="12" text-anchor="middle" fill="#ffffff">
        - Authentication
    </text>

    <!-- Image Processor Service -->
    <rect x="550" y="80" width="200" height="160" rx="10" fill="#38c172" opacity="0.8"/>
    <text x="650" y="110" font-family="Arial" font-size="16" text-anchor="middle" fill="#ffffff" font-weight="bold">
        Image Processor
    </text>
    <text x="650" y="140" font-family="Arial" font-size="12" text-anchor="middle" fill="#ffffff">
        - Async Processing
    </text>
    <text x="650" y="160" font-family="Arial" font-size="12" text-anchor="middle" fill="#ffffff">
        - Image Compression
    </text>
    <text x="650" y="180" font-family="Arial" font-size="12" text-anchor="middle" fill="#ffffff">
        - S3 Upload
    </text>
    <text x="650" y="200" font-family="Arial" font-size="12" text-anchor="middle" fill="#ffffff">
        - Retry Mechanism
    </text>

    <!-- Database Layer -->
    <rect x="50" y="400" width="200" height="120" rx="10" fill="#e3342f" opacity="0.8"/>
    <text x="150" y="430" font-family="Arial" font-size="16" text-anchor="middle" fill="#ffffff" font-weight="bold">
        PostgreSQL
    </text>
    <text x="150" y="460" font-family="Arial" font-size="12" text-anchor="middle" fill="#ffffff">
        - Product Data
    </text>
    <text x="150" y="480" font-family="Arial" font-size="12" text-anchor="middle" fill="#ffffff">
        - User Data
    </text>

    <!-- Cache Layer -->
    <rect x="300" y="400" width="200" height="120" rx="10" fill="#ffed4a" opacity="0.8"/>
    <text x="400" y="430" font-family="Arial" font-size="16" text-anchor="middle" fill="#000000" font-weight="bold">
        Redis Cache
    </text>
    <text x="400" y="460" font-family="Arial" font-size="12" text-anchor="middle" fill="#000000">
        - Product Cache
    </text>
    <text x="400" y="480" font-family="Arial" font-size="12" text-anchor="middle" fill="#000000">
        - Session Data
    </text>

    <!-- Message Queue -->
    <rect x="550" y="400" width="200" height="120" rx="10" fill="#9561e2" opacity="0.8"/>
    <text x="650" y="430" font-family="Arial" font-size="16" text-anchor="middle" fill="#ffffff" font-weight="bold">
        Kafka
    </text>
    <text x="650" y="460" font-family="Arial" font-size="12" text-anchor="middle" fill="#ffffff">
        - Image Processing Queue
    </text>
    <text x="650" y="480" font-family="Arial" font-size="12" text-anchor="middle" fill="#ffffff">
        - DLQ
    </text>

    <!-- Connections -->
    <!-- API to PostgreSQL -->
    <line x1="150" y1="240" x2="150" y2="400" stroke="#666666" stroke-width="2" stroke-dasharray="5,5"/>
    <!-- API to Redis -->
    <line x1="250" y1="160" x2="400" y2="400" stroke="#666666" stroke-width="2" stroke-dasharray="5,5"/>
    <!-- API to Kafka -->
    <line x1="250" y1="160" x2="650" y2="400" stroke="#666666" stroke-width="2" stroke-dasharray="5,5"/>
    <!-- Image Processor to Kafka -->
    <line x1="650" y1="240" x2="650" y2="400" stroke="#666666" stroke-width="2" stroke-dasharray="5,5"/>
    <!-- Image Processor to PostgreSQL -->
    <line x1="550" y1="160" x2="150" y2="400" stroke="#666666" stroke-width="2" stroke-dasharray="5,5"/>

    <!-- AWS S3 Cloud -->
    <rect x="300" y="80" width="200" height="100" rx="10" fill="#f66d9b" opacity="0.8"/>
    <text x="400" y="110" font-family="Arial" font-size="16" text-anchor="middle" fill="#ffffff" font-weight="bold">
        AWS S3
    </text>
    <text x="400" y="140" font-family="Arial" font-size="12" text-anchor="middle" fill="#ffffff">
        - Image Storage
    </text>
    
    <!-- Connection to S3 -->
    <line x1="550" y1="130" x2="500" y2="130" stroke="#666666" stroke-width="2" stroke-dasharray="5,5"/>
</svg>

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
