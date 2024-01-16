FROM golang:1.21 AS builder
WORKDIR /app
COPY . /app
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o main .

######
FROM alpine:latest AS runner

RUN apk add bash

WORKDIR /app

COPY --from=builder /app/main /app/spotify-history-saver
COPY run.sh /app/run.sh

RUN chmod +x /app/run.sh
RUN chmod +x /app/spotify-history-saver

EXPOSE 8080

CMD ["/app/run.sh"]
