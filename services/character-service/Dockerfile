FROM golang:1.19-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o /app/character-service ./services/character-service/cmd

FROM alpine:latest

RUN apk --no-cache add ca-certificates
WORKDIR /root/

COPY --from=builder /app/character-service .
COPY --from=builder /app/services/character-service/configs ./configs

EXPOSE 50051

CMD ["./character-service"]