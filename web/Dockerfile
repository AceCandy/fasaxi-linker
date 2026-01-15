# Build stage
FROM node:18-alpine AS builder

WORKDIR /build

# Copy root package files
COPY package.json pnpm-lock.yaml pnpm-workspace.yaml ./

# Copy web workspace
COPY web ./web

# Install pnpm
RUN npm install -g pnpm@9.1.0

# Install dependencies
RUN pnpm install --frozen-lockfile

# Build the application (skip type checking for Docker build)
RUN cd web/client && npx vite build

# Runtime stage
FROM nginx:alpine

# Copy built files
COPY --from=builder /build/web/dist/web /usr/share/nginx/html

# Copy nginx configuration
COPY web/nginx.conf /etc/nginx/conf.d/default.conf

EXPOSE 80

CMD ["nginx", "-g", "daemon off;"]
