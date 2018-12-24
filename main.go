package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"

	"github.com/gorilla/websocket"
)

type Args struct {
	Port   int
	Server bool
	Remote string
}

func main() {
	args := Args{}
	flag.IntVar(&args.Port, "port", 8000, "port for server")
	flag.BoolVar(&args.Server, "server", true, "server or client")
	flag.StringVar(&args.Remote, "remote", "127.0.0.1:8000", " server address i.e. 127.0.0.1:8000")

	flag.Parse()

	if args.Server {
		startServer(args)
	} else {
		runClient(args)
	}
}

func runClient(args Args) {

	u := url.URL{Scheme: "ws", Host: args.Remote, Path: "/ws"}
	log.Printf("connecting to %s", u.String())

	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)

	if err != nil {
		panic(err)
	}
	fmt.Println(c.LocalAddr())

	defer c.Close()

	done := make(chan struct{})

	go func() {
		defer close(done)
		for {
			_, message, err := c.ReadMessage()
			if err != nil {
				log.Println("read:", err)
				return
			}
			log.Printf("recv: %d", len(message))
			buf := bytes.NewBuffer(message)
			bufReader := bufio.NewReader(buf)
			req, _ := http.ReadRequest(bufReader)
			fmt.Println("RequestURI", req.RequestURI)
			// fmt.Println("req", req)

			body, _ := ioutil.ReadAll(req.Body)

			newReq, _ := http.NewRequest(req.Method, fmt.Sprintf("http://%s%s", req.Host, req.URL.String()), bytes.NewReader(body))

			res, resErr := http.DefaultClient.Do(newReq)
			if resErr != nil {
				log.Println(resErr)
			} else {
				log.Printf("Upstream tunnel res: %s\n", res.Status)

				buf2 := new(bytes.Buffer)

				res.Write(buf2)

				fmt.Println("Whole response", buf2.Len())

				c.WriteMessage(websocket.TextMessage, buf2.Bytes())
			}
		}
	}()

	<-done
}

func proxyHandler(msg chan *http.Response, outgoing chan *http.Request) func(w http.ResponseWriter, r *http.Request) {

	return func(w http.ResponseWriter, r *http.Request) {
		log.Println("Reverse proxy", r.Host, r.Method, r.URL.String())
		if r.Body != nil {
			defer r.Body.Close()
		}

		body, _ := ioutil.ReadAll(r.Body)

		req, _ := http.NewRequest(r.Method, fmt.Sprintf("http://localhost:3000%s", r.URL.Path),
			bytes.NewReader(body))

		// log.Printf("Request to tunnel: %s\n", string(body))
		outgoing <- req

		log.Println("waiting for response")

		res := <-msg

		log.Println("writing response from tunnel", res.ContentLength)

		for k, v := range res.Header {
			res.Header.Set(k, v[0])
		}

		w.WriteHeader(res.StatusCode)

		innerBody, _ := ioutil.ReadAll(res.Body)

		w.Write(innerBody)
	}
}

func startServer(args Args) {
	ch := make(chan *http.Response)
	outgoing := make(chan *http.Request)
	http.HandleFunc("/ws", serveWs(ch, outgoing))
	http.HandleFunc("/", proxyHandler(ch, outgoing))
	if err := http.ListenAndServe(fmt.Sprintf(":%d", args.Port), nil); err != nil {
		log.Fatal(err)
	}
}

func serveWs(msg chan *http.Response, outgoing chan *http.Request) func(w http.ResponseWriter, r *http.Request) {

	var upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}

	return func(w http.ResponseWriter, r *http.Request) {
		ws, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			if _, ok := err.(websocket.HandshakeError); !ok {
				log.Println(err)
			}
			return
		}

		fmt.Println(ws.RemoteAddr())

		done := make(chan struct{})

		go func() {
			defer close(done)
			for {
				msgType, message, err := ws.ReadMessage()
				if err != nil {
					log.Println("read:", err)
					return
				}

				if msgType == websocket.TextMessage {
					// log.Printf("Server recv: %s", message)

					reader := bytes.NewReader(message)
					scanner := bufio.NewReader(reader)
					res, _ := http.ReadResponse(scanner, nil)
					// log.Println(res, resErr)

					// body, _ := ioutil.ReadAll(res.Body)
					msg <- res
				}
			}
		}()

		go func() {
			defer close(done)
			for {
				fmt.Println("wait for outboundRequest")
				outboundRequest := <-outgoing
				// fmt.Println("outboundRequest", outboundRequest)
				buf := new(bytes.Buffer)

				outboundRequest.Write(buf)

				ws.WriteMessage(websocket.TextMessage, buf.Bytes())
			}

		}()

		<-done
	}
}
