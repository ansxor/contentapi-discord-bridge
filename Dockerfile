# Use the official Go image
FROM golang:1.20

# Set the working directory
WORKDIR /app

# Copy go mod files and download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy the source code
COPY . .

# Build the Go app
RUN go build -o contentapi-discord-bridge

# Run the binary
CMD ["./contentapi-discord-bridge"]