FROM golang:1.24.4-alpine AS builder
WORKDIR /src
ENV CGO_ENABLED=0

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o /app/server ./cmd/server

FROM alpine:3.18
RUN apk add --no-cache ca-certificates netcat-openbsd
COPY --from=builder /app/server /server
EXPOSE 8085
ENTRYPOINT ["/bin/sh","-c","until nc -z postgres 5432; do echo waiting for postgres; sleep 1; done; /server"]
