# Stage 1: Build React Frontend
FROM node:20-alpine AS ui-builder
WORKDIR /app/ui
COPY ui/package*.json ./
RUN npm install
COPY ui/ ./
RUN npm run build

# Stage 2: Build Go Binary
FROM golang:1.24-alpine AS go-builder
WORKDIR /app
COPY go.mod ./
COPY internal/ ./internal/
COPY docs/ ./docs/
COPY main.go ./
COPY --from=ui-builder /app/ui/dist ./ui/dist
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o go-app main.go

# Stage 3: Minimal Production Runtime
FROM alpine:latest
RUN apk add --no-cache ca-certificates tzdata
WORKDIR /app
COPY --from=go-builder /app/go-app /app/go-app

EXPOSE 8080
ENV APP_MODE=server
ENV PORT=8080

ENTRYPOINT ["/app/go-app"]
