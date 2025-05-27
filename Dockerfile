# Start from the official Go image for building
FROM golang:1.20-alpine AS builder

WORKDIR /app

# Copy go.mod and go.sum files and download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the source code
COPY . .

# Build the Go app
RUN go build -o main .

# Use a smaller image for running the app
FROM alpine:latest

# Copy the built binary from the builder stage
COPY --from=builder /app/main /main

# Expose port your app listens on
EXPOSE 8080

# Run the app
CMD ["/main"]
