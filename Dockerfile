FROM golang:1.24-alpine AS builder

#current working Directory
WORKDIR /app

COPY go.mod go.sum ./

# Download all dependencies. Dependencies are cached if go.mod and go.sum are unchanged.
RUN go mod download


COPY . .

# Build the Go application for Linux.
# CGO_ENABLED=0 disables CGO, making the binary statically linked and more portable.
# -o /usr/local/bin/linksprint specifies the output path and name of the executable.
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o /usr/local/bin/linksprint ./cmd/api

# Use a minimal Alpine Linux image as the base for the final image.
# This makes the final image very small, which is good for deployment.
FROM alpine:latest

# Set the working directory for the application
WORKDIR /root/

# Copy the compiled executable from the 'builder' stage
COPY --from=builder /usr/local/bin/linksprint .

# Expose port 3000, which your Fiber app listens on
EXPOSE 3000

# Command to run the executable when the container starts
CMD ["./linksprint"]