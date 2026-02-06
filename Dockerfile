# Build stage
FROM golang:1.22-alpine AS builder

WORKDIR /app

# Install git for go modules
RUN apk add --no-cache git

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o /sentinel ./cmd/sentinel

# Runtime stage
FROM alpine:3.19

WORKDIR /app

# Install git and docker CLI for tools
RUN apk add --no-cache git docker-cli ca-certificates

# Copy binary from builder
COPY --from=builder /sentinel /app/sentinel

# Create work directory
RUN mkdir -p /tmp/sentinel

EXPOSE 8080

CMD ["/app/sentinel"]
