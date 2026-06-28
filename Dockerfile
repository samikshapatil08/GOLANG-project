FROM golang:alpine AS builder
ENV http_proxy=""
ENV https_proxy=""
ENV HTTP_PROXY=""
ENV HTTPS_PROXY=""
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o server ./cmd/server/main.go

FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/server .
COPY .env.example .env
EXPOSE 8080
CMD ["./server"]
