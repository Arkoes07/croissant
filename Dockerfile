FROM golang:1.24-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o croissant ./cmd/croissant

FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/croissant .
EXPOSE 8080
CMD ["./croissant"]
