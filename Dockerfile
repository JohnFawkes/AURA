############################################################################
##### Stage 1: Build the backend application
############################################################################
FROM golang:alpine AS backend-builder

# Install required dependencies for cgo
RUN apk add --no-cache gcc musl-dev

# Set the working directory
WORKDIR /backend

# Copy the go.mod and go.sum files
COPY backend/go.mod backend/go.sum ./

# Download the dependencies
RUN go mod download

# Copy the source code
COPY backend/ ./

# Get the version from build arguments/environment variables
ARG APP_VERSION=dev

# Enable CGO and build the application 
ENV CGO_ENABLED=1
RUN go build -ldflags="-s -w -X main.APP_VERSION=$APP_VERSION" -o main .

############################################################################
##### Stage 2: Build the frontend application
############################################################################
FROM node:alpine AS frontend-builder

# Set the working directory
WORKDIR /frontend

# Copy the package.json and package-lock.json files
COPY frontend/package*.json ./

# Install the dependencies
RUN npm ci --omit=dev

# Copy the source code
COPY frontend/ ./

# Get the port number and version from build arguments/environment variables
ARG APP_VERSION=dev

# Set environment variables
ENV NEXT_PUBLIC_APP_VERSION=${APP_VERSION}
ENV NEXT_TELEMETRY_DISABLED=1

# Build the application
RUN npm run build || (echo "Build failed" && cat /frontend/.next/build-diagnostics.json && exit 1)

############################################################################
##### Stage 3: Build the final image
############################################################################
FROM node:alpine AS final

# Set the working directory
WORKDIR /app

# Install CA certificates and tzdata for timezone support
RUN apk add --no-cache ca-certificates tzdata

# Copy the backend application from the builder stage
COPY --from=backend-builder /backend/main .

# Copy the frontend build from the builder stage
COPY --from=frontend-builder /frontend/.next/standalone /app/
COPY --from=frontend-builder /frontend/.next/static /app/.next/static
COPY --from=frontend-builder /frontend/public /app/public

# Get the version from build arguments/environment variables
ARG APP_VERSION=dev

# Set environment variables
ENV NODE_ENV=production

# Expose the ports for both the backend and frontend
EXPOSE 3000
EXPOSE 8888

# Command to run both the backend and frontend
#CMD ["sh", "-c", "./main & NODE_ENV=production npm start --prefix /frontend"]
CMD ["sh", "-c", "./main & node server.js"]
