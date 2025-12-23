# Build stage
FROM golang:1.25.5-alpine3.21 AS builder
WORKDIR /app
COPY . .
RUN go build -o main main.go

# Run stage
FROM alpine:3.21
RUN apk add --no-cache netcat-openbsd
WORKDIR /app
COPY --from=builder /app/main .
COPY app.env .
COPY start.sh .
COPY db/migration ./db/migration

RUN chmod +x /app/start.sh

EXPOSE 8000
CMD [ "/app/main" ]
ENTRYPOINT [ "/app/start.sh" ]
