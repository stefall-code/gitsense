FROM golang:1.24-alpine AS builder

ENV GOPROXY=https://mirrors.tencent.com/go,direct
ENV GOSUMDB=sum.golang.google.cn

WORKDIR /app

COPY go.mod ./
COPY go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build -o /gitsense ./cmd/server

FROM alpine:latest
RUN apk add --no-cache ca-certificates
COPY --from=builder /gitsense /gitsense
ENTRYPOINT ["/gitsense"]
