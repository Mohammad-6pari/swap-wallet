FROM docker.haiocloud.com/golang:1.23-bullseye

WORKDIR /app

COPY .env .env

RUN go mod init swap-wallet

COPY . .

RUN go mod tidy

RUN go build -o swap-wallet main.go

EXPOSE 8080

CMD ["./swap-wallet"]