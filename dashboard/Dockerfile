# Stage 1: Build
FROM golang:1.22-alpine AS builder
RUN apk add --no-cache git
WORKDIR /src
COPY go.mod go.sum* ./
RUN go mod download
COPY . .
RUN go mod tidy && CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /arcticmon .

# Stage 2: Runtime
FROM alpine:3.19
RUN apk add --no-cache ca-certificates curl
COPY --from=builder /arcticmon /arcticmon
EXPOSE 3000
ENTRYPOINT ["/arcticmon"]
