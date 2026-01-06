FROM golang:1.24-alpine AS builder

WORKDIR /app

RUN apk add --no-cache git

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-s -w" \
    -o /app/main \
    .

FROM alpine:latest

RUN apk --no-cache add ca-certificates curl

RUN addgroup -g 1000 appuser && \
    adduser -D -u 1000 -G appuser appuser

COPY --from=builder /app/main /main

RUN chown appuser:appuser /main

USER appuser

EXPOSE 8080

ENTRYPOINT ["/main"]