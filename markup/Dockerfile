# Use the official Go image
FROM oven/bun:latest

# Set the working directory
WORKDIR /app

# Copy go mod files and download dependencies
COPY package.json bun.lockb ./

RUN bun install

# Copy the source code
COPY . .

# Run the binary
CMD ["bun", "run", "/app/server.ts"]