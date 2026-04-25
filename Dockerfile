#---------------------BUILD-----------------
FROM golang:1.25 AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o app ./cmd/main.go


#---------------------RUN-----------------
FROM ubuntu:22.04
RUN apt-get update && apt-get install -y ffmpeg && rm -rf /var/lib/apt/lists/*
WORKDIR /root/
COPY --from=builder /app/app .
CMD ["./app"]

