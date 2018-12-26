package handler

import (
	"bufio"
	"bytes"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

func MakeServeWebsocketHandler(msg chan *http.Response, outgoing chan *http.Request) func(w http.ResponseWriter, r *http.Request) {

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
