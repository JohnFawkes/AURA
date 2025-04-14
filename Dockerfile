##### Stage 1: Build the backend application
FROM golang:1.24 AS backend-builder

# Set the working directory
WORKDIR /backend

# Copy the go.mod and go.sum files
COPY backend/go.mod backend/go.sum ./

# Download the dependencies
RUN go mod download

# Copy the source code
COPY backend/ ./

# Build the application for aarch64
RUN GOARCH=arm64 go build -o main .

##### Stage 2: Build the frontend application
FROM node:latest AS frontend-builder

# Set the working directory
WORKDIR /frontend

# Copy the package.json and package-lock.json files
COPY frontend/package*.json ./

# Install the dependencies
RUN npm ci

# Copy the source code
COPY frontend/ ./

# Build the application
RUN npm run build

##### Stage 3: Build the final image
FROM debian:bookworm-slim

# Install CA certificates
RUN apt-get update && apt-get install -y ca-certificates && rm -rf /var/lib/apt/lists/*

# Set the working directory
WORKDIR /app

# Copy the backend application from the builder stage
COPY --from=backend-builder /backend/main .

# Copy the frontend build from the builder stage
COPY --from=frontend-builder /frontend/dist /frontend/dist

# Get the port number from the environment variable
ARG APP_PORT=8888
ENV APP_PORT=${APP_PORT}
ENV VITE_APP_PORT=${APP_PORT}
# Expose the port
EXPOSE ${APP_PORT}

# Command to run the application
CMD ["./main"]
