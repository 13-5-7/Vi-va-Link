FROM golang:1.26-alpine

# Install and configure golangci-lint for static analysis
RUN apk add --no-cache ca-certificates git
RUN go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.64.5
ENV PATH=$PATH:/go/bin

WORKDIR /app

# Install Air for hot reload
RUN go install github.com/air-verse/air@latest

# Copy go mod files first for layer caching
COPY go.mod go.sum* ./
RUN go mod download && go mod tidy 2>/dev/null || true

CMD ["air"]
