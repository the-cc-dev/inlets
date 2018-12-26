package exit

import (
	"fmt"
	"log"
	"net/http"

	"github.com/alexellis/inlets/pkg/handler"
)

func StartServer(upstream string, port int) {
	ch := make(chan *http.Response)
	outgoing := make(chan *http.Request)
	http.HandleFunc("/ws", handler.MakeServeWebsocketHandler(ch, outgoing))
	http.HandleFunc("/", handler.MakeProxyHandler(ch, outgoing, upstream))
	if err := http.ListenAndServe(fmt.Sprintf(":%d", port), nil); err != nil {
		log.Fatal(err)
	}
}
