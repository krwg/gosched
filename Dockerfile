FROM golang:1.25-alpine AS builder

WORKDIR /src
RUN apk add --no-cache git ca-certificates

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/gosched ./cmd/scheduler

FROM alpine:3.20

RUN apk add --no-cache ca-certificates wget
WORKDIR /app

COPY --from=builder /out/gosched /usr/local/bin/gosched
COPY tests/fixtures/tasks.json /app/tasks.json

EXPOSE 8080 50051

ENTRYPOINT ["gosched"]
CMD ["serve", "--http", ":8080", "--grpc", ":50051", "--algorithm", "edf", "--workers", "4"]
