FROM golang:1.26-alpine AS builder
WORKDIR /app
COPY go.mod ./
RUN go mod download
COPY . .
RUN go build -o dailycode-service .

FROM alpine:3.23
WORKDIR /app
COPY --from=builder /app/dailycode-service .
EXPOSE 9090
ENTRYPOINT [ "./dailycode-service" ]