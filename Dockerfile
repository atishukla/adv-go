# Use the official Golang image to build the application
FROM golang:1.23 AS builder

WORKDIR /app

# Copy go.mod and go.sum files to download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy the entire project
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o pod-logger .

# Use a minimal base image to run the application
FROM alpine:latest

WORKDIR /root/

# Copy the binary from the builder stage
COPY --from=builder /app/pod-logger .

# Command to run the application
CMD ["./pod-logger"]
