FROM golang:1.23-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o avito-internship ./cmd/server

FROM alpine:3.20

WORKDIR /app

COPY --from=builder /app/avito-internship /app/avito-internship
COPY migrations ./migrations

ENV HTTP_ADDR=:8080

CMD ["/app/avito-internship"]
