FROM golang:1.24-alpine AS builder

# Set working directory
WORKDIR /app

# Install dependencies
RUN apk add --no-cache git

# Copy go.mod and go.sum
COPY go.mod ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o address-validator .

# Create a minimal image
FROM alpine:latest

# Install CA certificates for HTTPS
RUN apk --no-cache add ca-certificates

WORKDIR /app

# Copy the binary from the builder stage
COPY --from=builder /app/address-validator .

# Create a non-root user and switch to it
RUN addgroup -S appgroup && adduser -S appuser -G appgroup
USER appuser

# Expose the port
EXPOSE 8080

# Run the application
CMD ["./address-validator"]