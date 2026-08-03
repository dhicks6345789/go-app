# Stage 1: Build Go Binary
FROM golang:1.24-alpine AS go-builder
WORKDIR /app
COPY go.mod ./
COPY main.go api.go ./
COPY ui/ ./ui/
COPY docs/ ./docs/
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o go-app main.go api.go

# Stage 2: Minimal Production Runtime
FROM alpine:latest
RUN apk add --no-cache ca-certificates tzdata
WORKDIR /app
COPY --from=go-builder /app/go-app /app/go-app

EXPOSE 8080
ENV APP_MODE=server
ENV PORT=8080

ENTRYPOINT ["/app/go-app"]
