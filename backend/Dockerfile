FROM golang:1.24-alpine AS builder

WORKDIR /app
COPY go.mod ./
RUN go mod download
COPY . .
RUN go mod tidy && CGO_ENABLED=0 go build -o /gitsense ./cmd/server

FROM golang:1.24-alpine
RUN apk --no-cache add ca-certificates
COPY --from=builder /gitsense /usr/local/bin/gitsense
EXPOSE 8080
CMD ["gitsense"]
