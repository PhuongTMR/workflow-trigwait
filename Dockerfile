# Build stage
FROM golang:1.21-alpine AS builder

WORKDIR /app
COPY go.mod ./
COPY cmd/ ./cmd/

RUN go build -ldflags="-s -w" -o workflow-trigwait ./cmd/main.go

# Runtime stage
FROM alpine:3.15.0

RUN apk update && apk --no-cache add ca-certificates

COPY --from=builder /app/workflow-trigwait /workflow-trigwait

ENTRYPOINT ["/workflow-trigwait"]
