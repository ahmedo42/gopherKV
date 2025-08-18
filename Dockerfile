# Stage 1: The Build Stage
FROM golang:1.24.5-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .


RUN CGO_ENABLED=0 GOOS=linux go build -o /go-app ./src

# Stage 2: Use a minimal Alpine image to reduce the final image size.
FROM alpine:latest

WORKDIR /root/

COPY --from=builder /go-app .

EXPOSE 8080

CMD ["./go-app"]
