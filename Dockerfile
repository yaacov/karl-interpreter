# Build stage
FROM golang:1.24.7-alpine3.19 AS builder

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o karl ./cmd/karl

# Final stage
FROM alpine:3.19

RUN apk --no-cache upgrade && apk --no-cache add ca-certificates

WORKDIR /root/

# Copy the binary from builder stage
COPY --from=builder /app/karl .

# Make it executable
RUN chmod +x ./karl

# Set entrypoint
ENTRYPOINT ["./karl"]
