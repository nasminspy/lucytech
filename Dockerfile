# Use the official Go image as the base
FROM golang:1.20-alpine

# Set the working directory inside the container
WORKDIR /app

# Install necessary packages
RUN apk add --no-cache git

# Copy go.mod and go.sum files to leverage Docker cache
COPY go.mod go.sum ./

# Download Go module dependencies
RUN go mod download

# Copy the rest of the application source code
COPY . .

# Build the Go application
RUN go build -o lucytech .

# Expose the application's port
EXPOSE 8080

# Set the entry point for the container
CMD ["./lucytech"]
