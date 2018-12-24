FROM golang:1.10.4 as build

WORKDIR /go/src/github.com/alexellis/inlets

COPY vendor     vendor
COPY main.go  .

RUN CGO_ENABLED=0 go build -a -installsuffix cgo --ldflags "-s -w" -o /usr/bin/server

FROM alpine:3.8

COPY --from=build /usr/bin/server /root/

EXPOSE 80

WORKDIR /root/

CMD ["./server"]
