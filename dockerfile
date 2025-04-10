FROM golang:1.24

WORKDIR ${GOPATH}/PVZ-service/
COPY . ${GOPATH}/PVZ-service/

RUN go build -o /build ./cmd \
    && go clean -cache -modcache

EXPOSE 8080

CMD ["/build"]