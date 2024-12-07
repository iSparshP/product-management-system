# cmd/image-processor/Dockerfile

# Stage 1: Build the application
FROM golang:1.23-alpine AS builder

# Install necessary packages
RUN apk update && apk add --no-cache git

# Set working directory
WORKDIR /app

# Copy go.mod and go.sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the source code
COPY . .

# Build the application
RUN go build -o image-processor ./cmd/image-processor

# Stage 2: Run the application
FROM alpine:latest

# Install necessary packages
RUN apk --no-cache add ca-certificates

# Set working directory
WORKDIR /root/

# Copy the binary from builder
COPY --from=builder /app/image-processor .

# Expose any necessary ports (if applicable)
EXPOSE 9090

# Command to run the executable
CMD ["./image-processor"]
