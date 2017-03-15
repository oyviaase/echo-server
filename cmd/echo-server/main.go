package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/gorilla/websocket"
	"strings"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	fmt.Printf("Echo server listening on port %s.\n", port)
	err := http.ListenAndServe(":"+port, http.HandlerFunc(handler))
	if err != nil {
		panic(err)
	}
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(*http.Request) bool {
		return true
	},
}

func handler(wr http.ResponseWriter, req *http.Request) {
	fmt.Printf("%s | %s %s\n", req.RemoteAddr, req.Method, req.URL)
	if websocket.IsWebSocketUpgrade(req) {
		serveWebSocket(wr, req)
	} else if req.URL.Path == "/.ws" {
		wr.Header().Add("Content-Type", "text/html")
		wr.WriteHeader(200)
		io.WriteString(wr, websocketHTML)
	} else {
		serveHTTP(wr, req)
	}
}

func serveWebSocket(wr http.ResponseWriter, req *http.Request) {
	connection, err := upgrader.Upgrade(wr, req, nil)
	if err != nil {
		fmt.Printf("%s | %s\n", req.RemoteAddr, err)
		return
	}

	defer connection.Close()
	fmt.Printf("%s | upgraded to websocket\n", req.RemoteAddr)

	var message []byte

	host, err := os.Hostname()
	if err == nil {
		message = []byte(fmt.Sprintf("Request served by %s", host))
	} else {
		message = []byte(fmt.Sprintf("Server hostname unknown: %s", err.Error()))
	}

	err = connection.WriteMessage(websocket.TextMessage, message)
	if err == nil {
		var messageType int

		for {
			messageType, message, err = connection.ReadMessage()
			if err != nil {
				break
			}

			if messageType == websocket.TextMessage {
				fmt.Printf("%s | txt | %s\n", req.RemoteAddr, message)
			} else {
				fmt.Printf("%s | bin | %d byte(s)\n", req.RemoteAddr, len(message))
			}

			err = connection.WriteMessage(messageType, message)
			if err != nil {
				break
			}
		}
	}

	if err != nil {
		fmt.Printf("%s | %s\n", req.RemoteAddr, err)
	}
}

func serveHTTP(wr http.ResponseWriter, req *http.Request) {
	// Take in extra headers to present
	addheaders := os.Getenv("ADD_HEADERS")
	if len(addheaders) > 0 {
		rawjson := json.RawMessage(addheaders)
		var headerjson map[string]interface{}
		json.Unmarshal(rawjson, &headerjson)
		for k, v := range headerjson {
			//fmt.Printf("%s = %s", k, v)
			wr.Header().Add(k, v.(string))
	    	}
	}


	wr.Header().Add("Content-Type", "text/plain")
	wr.WriteHeader(200)

	host, err := os.Hostname()
	if err == nil {
		fmt.Fprintf(wr, "-> My hostname is: %s\n\n", host)
	} else {
		fmt.Fprintf(wr, "-> Server hostname unknown: %s\n\n", err.Error())
	}

	podname := os.Getenv("POD_NAME")
	if len(podname) > 0 {
		fmt.Fprintf(wr, "-> My Pod Name is: %s\n", podname)
	}

	podnamespace := os.Getenv("POD_NAMESPACE")
	if len(podname) > 0 {
		fmt.Fprintf(wr, "-> My Pod Namespace is: %s\n", podnamespace)
	}

	podip := os.Getenv("POD_IP")
	if len(podname) > 0 {
		fmt.Fprintf(wr, "-> My Pod IP is: %s\n\n", podip)
	}

	fmt.Fprintf(wr, "-> Requesting IP: %s\n\n", req.RemoteAddr)

	fmt.Fprintln(wr, "-> Request Headers: \n")
	fmt.Fprintf(wr, "  %s %s %s\n", req.Proto, req.Method, req.URL)
	fmt.Fprintf(wr, "\n")
	fmt.Fprintf(wr, "  Host: %s\n", req.Host)
	for key, values := range req.Header {
		for _, value := range values {
			fmt.Fprintf(wr, "  %s: %s\n", key, value)
		}
	}
	fmt.Fprintf(wr, "\n\n")
	fmt.Fprintln(wr, "-> Response Headers: \n")
	fmt.Fprintln(wr, " >> Note that you may also see `Transfer-Encoding` and `Date`! \n")
	for k,v := range wr.Header() {
		fmt.Fprintf(wr, "  %s:%s\n", k,v )
	}


	fmt.Fprintf(wr, "\n\n")
	fmt.Fprintln(wr, "-> My environment:")
	for _, e := range os.Environ() {
		pair := strings.Split(e, "=")
		fmt.Fprintf(wr, "  %s=%s\n", pair[0],pair[1])
	}

	// Lets get resolv.conf
	fmt.Fprintf(wr, "\n\n")
	resolvfile, err := ioutil.ReadFile("/etc/resolv.conf") // just pass the file name
	if err != nil {
		fmt.Fprint(wr, "%s", err)
	}

	str := string(resolvfile) // convert content to a 'string'

	fmt.Fprintf(wr, "-> Contents of /etc/resolv.conf: \n%s\n\n", str) // print the content as a 'string'

	// Lets get hosts
	fmt.Fprintf(wr, "\n\n")
	hostsfile, err := ioutil.ReadFile("/etc/hosts") // just pass the file name
	if err != nil {
		fmt.Fprint(wr, "%s", err)
	}

	hostsstr := string(hostsfile) // convert content to a 'string'

	fmt.Fprintf(wr, "-> Contents of /etc/hosts: \n%s\n\n", hostsstr) // print the content as a 'string'


	fmt.Fprintln(wr, "")
	io.Copy(wr, req.Body)
}
