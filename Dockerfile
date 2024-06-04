# First stage: build the application
FROM golang:1.20-alpine AS builder

# Install necessary packages
RUN apk update && apk add --no-cache git ca-certificates && update-ca-certificates

# Set the current working directory inside the container
WORKDIR /app

# Copy the go.mod and go.sum files to download dependencies first
COPY go.mod go.sum ./

# Copy the source code into the container
COPY . .

# Build the application
RUN go build -o api ./cmd/api

# Second stage: run the application
FROM alpine:latest

# Install necessary certificates
RUN apk add --no-cache ca-certificates && update-ca-certificates

# Set the current working directory inside the container
WORKDIR /app

# Copy the built binary from the builder stage
COPY --from=builder /app .

# Expose the necessary port
EXPOSE 8000

# Command to run the application
CMD ["./api"]
