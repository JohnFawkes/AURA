############################################################################
##### Stage 1: Build the backend application
############################################################################
FROM golang:1.24 AS backend-builder

# Install required dependencies for cgo
RUN apt-get update && apt-get install -y gcc libc6-dev && rm -rf /var/lib/apt/lists/*

# Set the working directory
WORKDIR /backend

# Copy the go.mod and go.sum files
COPY backend/go.mod backend/go.sum ./

# Download the dependencies
RUN go mod download

# Copy the source code
COPY backend/ ./

# Enable CGO and build the application 
ENV CGO_ENABLED=1
ARG APP_VERSION=dev
RUN go build -ldflags "-X main.APP_VERSION=$APP_VERSION" -o main .

############################################################################
##### Stage 2: Build the frontend application
############################################################################
FROM node:20-alpine AS frontend-builder

# Set the working directory
WORKDIR /frontend

# Copy the package.json and package-lock.json files
COPY frontend/package*.json ./

# Install the dependencies
RUN npm ci

# Copy the source code
COPY frontend/ ./

# Get the port number and version from build arguments/environment variables
ARG APP_VERSION=dev
# ARG FRONTEND_PORT=3000
# ARG BACKEND_PORT=8888

# Set environment variables
# ENV NEXT_PUBLIC_BACKEND_PORT=${BACKEND_PORT}
# ENV NEXT_PUBLIC_FRONTEND_PORT=${FRONTEND_PORT}
ENV NEXT_PUBLIC_APP_VERSION=${APP_VERSION}
ENV NEXT_TELEMETRY_DISABLED=1

# Build the application
RUN npm run build || (echo "Build failed" && cat /frontend/.next/build-diagnostics.json && exit 1)

############################################################################
##### Stage 3: Build the final image
############################################################################
FROM node:20

# Set the working directory
WORKDIR /app

# Copy the backend application from the builder stage
COPY --from=backend-builder /backend/main .

# Copy the frontend build from the builder stage
COPY --from=frontend-builder /frontend/.next /frontend/.next
COPY --from=frontend-builder /frontend/public /frontend/public
COPY --from=frontend-builder /frontend/package.json /frontend/package.json
COPY --from=frontend-builder /frontend/node_modules /frontend/node_modules

# Get the port number and version from build arguments/environment variables
# ARG FRONTEND_PORT=3000
# ARG BACKEND_PORT=8888
ARG APP_VERSION=dev

# Set environment variables
# ENV FRONTEND_PORT=${FRONTEND_PORT}
# ENV BACKEND_PORT=${BACKEND_PORT}
ENV NODE_ENV=production

# Expose the ports for both the backend and frontend
# EXPOSE ${FRONTEND_PORT}
# EXPOSE ${BACKEND_PORT}
EXPOSE 3000

# Command to run both the backend and frontend
CMD ["sh", "-c", "./main & NODE_ENV=production npm start --prefix /frontend"]

