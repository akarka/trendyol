FROM golang:1.23-alpine
WORKDIR /app
COPY cmd/webhook_tester/main_standalone.go .
RUN go build -o /app_bin main_standalone.go
EXPOSE 8080
CMD ["/app_bin"]
