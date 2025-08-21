# 🚀 glab-tui Docker Image
FROM golang:1.21-alpine AS builder

# 📦 Install dependencies
RUN apk add --no-cache git ca-certificates tzdata

# 📁 Set working directory
WORKDIR /app

# 📋 Copy go mod files
COPY go.mod go.sum ./

# 📥 Download dependencies
RUN go mod download

# 📂 Copy source code
COPY . .

# 🏗️ Build the binary
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o glab-tui .

# 🎯 Final stage - minimal image
FROM alpine:latest

# 📦 Install runtime dependencies
RUN apk --no-cache add ca-certificates git

# 👤 Create non-root user
RUN adduser -D -s /bin/sh glab-user

# 📁 Set working directory
WORKDIR /home/glab-user

# 📋 Copy binary from builder
COPY --from=builder /app/glab-tui /usr/local/bin/glab-tui

# 🔧 Make binary executable
RUN chmod +x /usr/local/bin/glab-tui

# 👤 Switch to non-root user
USER glab-user

# 🎯 Set entrypoint
ENTRYPOINT ["glab-tui"]

# 📝 Default command
CMD ["help"]

# 🏷️ Labels
LABEL org.opencontainers.image.title="glab-tui"
LABEL org.opencontainers.image.description="A k9s-inspired TUI for GitLab CI/CD pipelines"
LABEL org.opencontainers.image.source="https://github.com/rkristelijn/glab-tui"
LABEL org.opencontainers.image.licenses="MIT"
