package main

import (
	"flag"
	"log"

	"github.com/alexellis/inlets/pkg/exit"
	"github.com/alexellis/inlets/pkg/host"
)

type Args struct {
	Port     int
	Server   bool
	Remote   string
	Upstream string
}

func main() {
	args := Args{}
	flag.IntVar(&args.Port, "port", 8000, "port for server")
	flag.BoolVar(&args.Server, "server", true, "server or client")
	flag.StringVar(&args.Remote, "remote", "127.0.0.1:8000", " server address i.e. 127.0.0.1:8000")
	flag.StringVar(&args.Upstream, "upstream", "http://127.0.0.1:3000", "upstream server i.e. http://127.0.0.1:3000")

	flag.Parse()

	log.Printf("Upstream: %s", args.Upstream)

	if args.Server {
		exit.StartServer(args.Remote, args.Port)

	} else {
		host.RunClient(args.Remote)
	}
}
