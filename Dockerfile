FROM golang:1.20-alpine

# Install necessary packages
RUN apk update && apk add --no-cache git ca-certificates && update-ca-certificates

# Set the current working directory inside the container
WORKDIR /app

# Copy the source code into the container
COPY . .

# Download all dependencies
RUN go mod download

# Build the application
RUN go build -o api ./cmd/api

# Expose the necessary port
EXPOSE 8000

# Command to run the application
CMD ["./api"]
