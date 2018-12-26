package handler

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/alexellis/inlets/pkg/proxy"
)

// MakeProxyHandler makes a handler to process requests to the reverse proxy
func MakeProxyHandler(msg chan *http.Response, outgoing chan *http.Request, upstream string) func(w http.ResponseWriter, r *http.Request) {

	return func(w http.ResponseWriter, r *http.Request) {
		log.Println("Reverse proxy", r.Host, r.Method, r.URL.String())
		if r.Body != nil {
			defer r.Body.Close()
		}

		body, _ := ioutil.ReadAll(r.Body)

		req, _ := http.NewRequest(r.Method, fmt.Sprintf("%s%s", upstream, r.URL.Path),
			bytes.NewReader(body))
		proxy.CopyHeaders(req.Header, &r.Header)

		// log.Printf("Request to tunnel: %s\n", string(body))
		outgoing <- req

		log.Println("waiting for response")

		res := <-msg

		log.Println("writing response from tunnel", res.ContentLength)

		proxy.CopyHeaders(w.Header(), &res.Header)
		w.WriteHeader(res.StatusCode)

		innerBody, _ := ioutil.ReadAll(res.Body)

		w.Write(innerBody)
	}
}
