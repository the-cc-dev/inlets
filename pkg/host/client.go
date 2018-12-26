package host

import (
	"bufio"
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"

	"github.com/alexellis/inlets/pkg/proxy"
	"github.com/gorilla/websocket"
)

func RunClient(remote string) {

	client := http.DefaultClient
	client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	}

	u := url.URL{Scheme: "ws", Host: remote, Path: "/ws"}
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

			// proxyToUpstream
			log.Printf("recv: %d", len(message))
			buf := bytes.NewBuffer(message)
			bufReader := bufio.NewReader(buf)
			req, _ := http.ReadRequest(bufReader)
			fmt.Println("RequestURI", req.RequestURI)

			body, _ := ioutil.ReadAll(req.Body)

			newReq, _ := http.NewRequest(req.Method, fmt.Sprintf("http://%s%s", req.Host, req.URL.String()), bytes.NewReader(body))

			proxy.CopyHeaders(newReq.Header, &req.Header)

			res, resErr := client.Do(newReq)

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
