FROM alpine:latest

WORKDIR /app

COPY loggerServiceApp .

CMD ["./loggerServiceApp"]
