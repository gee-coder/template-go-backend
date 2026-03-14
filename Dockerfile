FROM golang:1.26.1-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /bin/app ./cmd/api

FROM alpine:3.21

WORKDIR /app
COPY --from=builder /bin/app /app/app
COPY configs /app/configs
COPY docs /app/docs

EXPOSE 8080
CMD ["/app/app"]

