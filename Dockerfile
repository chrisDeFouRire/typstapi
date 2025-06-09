# Build stage
FROM golang:1.24-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY *.go .

# Build the application
RUN go build -o typstapi

# Final stage
FROM alpine:3.19

# Install Typst CLI and pdfcpu
RUN apk add --no-cache curl \
    && curl -L https://github.com/typst/typst/releases/latest/download/typst-x86_64-unknown-linux-musl.tar.xz -o typst.tar.xz \
    && tar xf typst.tar.xz \
    && mv typst-x86_64-unknown-linux-musl/typst /usr/local/bin/ \
    && rm -rf typst.tar.xz typst-x86_64-unknown-linux-musl \
    && apk del curl

# Copy the compiled application from builder
COPY --from=builder /app/typstapi /usr/local/bin/

# Set environment variables
ENV PORT=8080

# Create a non-root user
RUN adduser -D appuser
USER appuser

# Expose the port
EXPOSE 8080

# Run the application
CMD ["typstapi"] 