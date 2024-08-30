# Use the official Golang image as the base image
FROM golang:1.22-alpine AS build

# Set the working directory inside the container
WORKDIR /app

# Copy the Go module files and download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the source code
COPY . .

# Build the Go application
RUN go build -o /go-tunnel

# Use a smaller base image for the final container
FROM alpine:latest

# Set the working directory inside the container
WORKDIR /root/

# Copy the compiled Go binary from the build stage
COPY --from=build /go-tunnel .

# Expose the necessary ports
EXPOSE 3000 4455

# Set the entrypoint to the binary, so it can receive command-line arguments
ENTRYPOINT ["./go-tunnel"]

# Default command to run (optional, can be overridden by passing arguments)
CMD ["--mode=server"]
