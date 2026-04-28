FROM golang:1.26-alpine

WORKDIR /app

# Install Air for hot reload
RUN go install github.com/air-verse/air@latest

# Copy go mod files first for layer caching
COPY go.mod go.sum* ./
RUN go mod download && go mod tidy 2>/dev/null || true

CMD ["air"]
