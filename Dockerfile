# --- STAGE 1: Build ---
# Use the official Golang image as the build environment
FROM golang:1.26.2-alpine3.23 AS builder

# Set the working directory inside the container
WORKDIR /app

# Copy dependency files first to leverage Docker layer caching
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the source code
COPY . .

# Build the Go binary
# CGO_ENABLED=0 creates a statically linked binary (needed for minimal images)
# -ldflags="-s -w" reduces binary size by stripping debug symbols
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o result_lookup_api ./cmd/api

# --- STAGE 2: Run ---
# Use a minimal base image for the final production container
FROM alpine:3.19

# Define username
ARG USERNAME=appuser

# Create the user and group
RUN addgroup $USERNAME \
    && adduser -S -G $USERNAME $USERNAME

# Set working directory for the final image
WORKDIR /app

# Copy only the compiled binary from the builder stage
COPY --chown=$USERNAME:$USERNAME --from=builder /app/result_lookup_api .

# Switch to the non-root user
USER $USERNAME

# Expose the port your app runs on
EXPOSE 8080

# Command to run the application
CMD ["./result_lookup_api"]