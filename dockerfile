FROM golang:1.23

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod tidy

COPY . .

COPY .env .env

COPY .env.secret .env.secret

RUN mkdir -p /app/build \
    && go build -o /app/build/pvz-service ./cmd/pvz-service \
    && go clean -cache -modcache

EXPOSE 8080

CMD ["/app/build/pvz-service"]
