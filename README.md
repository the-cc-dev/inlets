# inlets
an open-source HTTP tunnel for development


## Intro

inlets is a reverse proxy and HTTP tunnel built to help you expose your internal or development websites to the public internet.

Why do we need this project? Similar tools such as ngrok or Argo from CloudFlare are expensive and closed-source. Other open-source tunnel tools are designed to set up a static tunnel. inlets aims to dynamically bind and discover your local services to DNS entries with automated TLS certificates to a public exit node.

## Status

This is an early prototype to test out some ideas around using websockets.

## Testing it out

You will need Golang 1.10 or 1.11 on both the exit-node or server and the client.

* On the server or exit-node

Start the tunnel server on a machine with a publicly-accessible IPv4 IP address such as a VPS.

```
go get -u github.com/alexellis/inlets
cd $GOPATH/src/github.com/alexellis/inlets

go run -server=true -port=80
```

Note down your public IPv4 IP address i.e. 192.168.0.101

* On your dev machine start an example service

This service generates hashes and is an example we want to share online

```
go get -u github.com/alexellis/hash-browns
cd $GOPATH/src/github.com/alexellis/hash-browns

port=3000 go run server.go 
```

* On your dev machine

Start the tunnel client

```
go get -u github.com/alexellis/inlets
cd $GOPATH/src/github.com/alexellis/inlets

go run -server=false -remote=192.168.0.101:80
```

Finally with an example server running and a tunnel server and a tunnel client send a request to the public IP address i.e.:

```
curl -d "hash this" http://192.168.0.101/hash
```

You will see the traffic pass between the exit node / server and your development machine. You'll see the hash message appear in the logs as below:

```
~/go/src/github.com/alexellis/hash-browns$ port=3000 go run server.go 
2018/12/23 20:15:00 Listening on port: 3000
"hash this"
```

Now check the metrics:

```
curl http://192.168.0.101/metrics | grep hash
```

