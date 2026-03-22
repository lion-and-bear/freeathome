# === Build Stage ===
FROM golang:1.26 AS builder

ARG commit=unknown
ARG version=unknown

WORKDIR /app

# Download modules if necessary
COPY go.mod go.sum ./
RUN go mod download

# Copy the source code...
COPY pkg ./pkg
COPY internal ./internal
COPY cmd/cli ./cmd/cli

# ... and compile
RUN CGO_ENABLED=0 GOOS=linux go build \
  -ldflags="-s -w -X 'main.version=$version' -X 'main.commit=$commit'" \
  -o /app/fh /app/cmd/cli/main.go

# === Bonus Stage: Create non-root user ===
FROM alpine:3.22 AS security_provider
RUN addgroup -S app && \
    adduser -S app -G app

# === Runtime Stage (Scratch) ===
FROM scratch
LABEL org.opencontainers.image.source=https://github.com/lion-and-bear/freeathome
LABEL org.opencontainers.image.description="The free@home CLI application"
LABEL org.opencontainers.image.licenses=MIT

# Copy the passwd file from the security provider stage
COPY --from=security_provider /etc/passwd /etc/passwd
USER app

# Copy the compiled binary from the build stage
COPY --from=builder /app/fh /fh

# Set the entrypoint and command
ENTRYPOINT ["/fh"]
