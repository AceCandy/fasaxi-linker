# Stage 1: Build frontend
FROM node:18-alpine AS frontend-builder

WORKDIR /build

# Copy root package files
COPY package.json pnpm-lock.yaml pnpm-workspace.yaml ./

# Copy web workspace
COPY web ./web

# Install pnpm
RUN npm install -g pnpm@9.1.0

# Install dependencies
RUN pnpm install --frozen-lockfile

# Build the application
RUN cd web/client && npx vite build

# Stage 2: Build backend
FROM golang:alpine AS backend-builder

WORKDIR /build

# Copy go mod files
COPY server/go.mod server/go.sum ./
RUN go mod download

# Copy server source code
COPY server/ ./

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o server cmd/server/main.go

# Stage 3: Runtime
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /app

# Copy backend binary
COPY --from=backend-builder /build/server .

# Copy frontend static files
COPY --from=frontend-builder /build/web/dist/web ./static

# Set environment variable for static files path
ENV STATIC_PATH=/app/static

EXPOSE 9090

CMD ["./server"]
