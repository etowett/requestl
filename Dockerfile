# Build stage
FROM golang:1.14-alpine as builder
RUN apk add git bash
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN ./build.sh

# Run image
FROM iron/go:1.10.2
RUN apk update && \
    apk add mailcap tzdata && \
    rm /var/cache/apk/*
COPY --from=builder /app/requestl /usr/bin
CMD ["requestl"]
