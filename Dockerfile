FROM golang:1.20-alpine AS builder

WORKDIR /app

# Copy go.mod and go.sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o inventory-api ./cmd/api

# Use a minimal alpine image for the final container
FROM alpine:latest  

WORKDIR /root/

# Add ca-certificates for HTTPS
RUN apk --no-cache add ca-certificates

# Copy the binary from builder
COPY --from=builder /app/inventory-api .
COPY --from=builder /app/.env.example ./.env

# Expose port
EXPOSE 8080

# Command to run
CMD ["./inventory-api"]