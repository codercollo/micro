FROM alpine:latest
RUN mkdir /app

COPY frontEndLinux /app

CMD ["/app/frontEndLinux"]