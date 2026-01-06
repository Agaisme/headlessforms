# Stage 1: Build Frontend (SvelteKit)
FROM node:20-alpine AS frontend-builder
WORKDIR /app/web
COPY web/package*.json ./
RUN npm install
COPY web/src ./src
COPY web/static ./static
COPY web/svelte.config.js web/vite.config.ts web/tsconfig.json web/tailwind.config.js web/postcss.config.js ./
RUN npm run build

# Stage 2: Build the Go binary
FROM golang:1.24-alpine AS backend-builder
WORKDIR /app
# Install C compiler for SQLite (CGO)
RUN apk add --no-cache gcc musl-dev

COPY go.mod go.sum ./
RUN go mod download

# Create the web directory structure
RUN mkdir -p web

# Copy frontend build output
COPY --from=frontend-builder /app/web/build ./web/build

# Copy the embed.go file that references the build directory
COPY web/embed.go ./web/

# Now copy Go source code
COPY cmd/ ./cmd/
COPY internal/ ./internal/
COPY docs/ ./docs/
COPY Makefile ./

# Build Static Binary with CGO enabled (required for sqlite)
ENV CGO_ENABLED=1
RUN go build -ldflags="-s -w" -o server ./cmd/server

# Stage 3: Final Production Image
FROM alpine:latest
WORKDIR /app

# Install ca-certificates for HTTPS
RUN apk add --no-cache ca-certificates tzdata

# Create non-root user for security
RUN addgroup -g 1001 -S appgroup && \
    adduser -u 1001 -S appuser -G appgroup

COPY --from=backend-builder /app/server .

# Expose port
EXPOSE 8080
ENV PORT=8080
ENV DATA_DIR=/data

# Create directory for SQLite file with proper permissions
RUN mkdir /data && chown -R appuser:appgroup /data /app
VOLUME /data

# Switch to non-root user
USER appuser

CMD ["./server"]
