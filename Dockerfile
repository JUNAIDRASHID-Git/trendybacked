# Use official Go image as build stage
FROM golang:1.20-alpine AS builder

# Set working directory inside the container
WORKDIR /app

# Copy go.mod and go.sum files and download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy all the source code files
COPY . .

# Build the Go binary (adjust main.go or your entry file)
RUN go build -o main .

# Use a smaller image for the final container
FROM alpine:latest

# Copy the binary from the builder stage
COPY --from=builder /app/main /app/main

# Set working directory
WORKDIR /app

# Expose port (adjust if your app uses a different port)
EXPOSE 8080

# Run the executable
CMD ["./main"]
