// Package main is the executable for the echo server.
package main

import (
	"bytes"
	"embed"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"sort"
	"strconv"
	"strings"
	"text/template"
	"time"

	"github.com/gorilla/websocket"
	"log"
	"sort"
	"strings"
)

func RunServer(addr string, sslAddr string, ssl map[string]string) chan error {
	// Thanks buddy guy: http://stackoverflow.com/a/29468115
	errs := make(chan error)

	go func() {
		fmt.Printf("Echo server listening on port %s.\n", addr)
		if err := http.ListenAndServe(addr, nil); err != nil {
			errs <- err
		}

	}()

	go func() {
		fmt.Printf("Echo server listening on ssl port %s.\n", sslAddr)
		if err := http.ListenAndServeTLS(sslAddr, ssl["cert"], ssl["key"], nil); err != nil {
			errs <- err
		}
	}()

	return errs
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	sslport := os.Getenv("SSLPORT")
	if sslport == "" {
		sslport = "8443"
	}

	http.HandleFunc("/", http.HandlerFunc(handler))

	errs := RunServer(":"+port, ":"+sslport, map[string]string{
		"cert": "cert.pem",
		"key":  "key.pem",
	})

	// This will run forever until channel receives error
	select {
	case err := <-errs:
		log.Printf("Could not start serving service due to (error: %s)", err)
	}
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(*http.Request) bool {
		return true
	},
}

func handler(wr http.ResponseWriter, req *http.Request) {
	defer req.Body.Close()

	if os.Getenv("LOG_HTTP_BODY") != "" || os.Getenv("LOG_HTTP_HEADERS") != "" {
		fmt.Printf("--------  %s | %s %s\n", req.RemoteAddr, req.Method, req.URL)
	} else {
		fmt.Printf("%s | %s %s\n", req.RemoteAddr, req.Method, req.URL)
	}

	if os.Getenv("LOG_HTTP_HEADERS") != "" {
		fmt.Printf("Headers\n")
		printHeaders(os.Stdout, req.Header)
	}

	if os.Getenv("LOG_HTTP_BODY") != "" {
		buf := &bytes.Buffer{}
		buf.ReadFrom(req.Body) // nolint:errcheck

		if buf.Len() != 0 {
			w := hex.Dumper(os.Stdout)
			w.Write(buf.Bytes()) // nolint:errcheck
			w.Close()
		}

		// Replace original body with buffered version so it's still sent to the
		// browser.
		req.Body.Close()
		req.Body = io.NopCloser(
			bytes.NewReader(buf.Bytes()),
		)
	}

	sendServerHostnameString := os.Getenv("SEND_SERVER_HOSTNAME")
	if v := req.Header.Get("X-Send-Server-Hostname"); v != "" {
		sendServerHostnameString = v
	}

	sendServerHostname := !strings.EqualFold(
		sendServerHostnameString,
		"false",
	)

	for _, line := range os.Environ() {
		parts := strings.SplitN(line, "=", 2)
		key, value := parts[0], parts[1]

		if name, ok := strings.CutPrefix(key, `SEND_HEADER_`); ok {
			wr.Header().Set(
				strings.ReplaceAll(name, "_", "-"),
				value,
			)
		}
	}

	if websocket.IsWebSocketUpgrade(req) {
		serveWebSocket(wr, req, sendServerHostname)
	} else if path.Base(req.URL.Path) == ".ws" {
		serveFrontend(wr, req)
	} else if path.Base(req.URL.Path) == ".sse" {
		serveSSE(wr, req, sendServerHostname)
	} else {
		serveHTTP(wr, req, sendServerHostname)
	}
}

func serveWebSocket(wr http.ResponseWriter, req *http.Request, sendServerHostname bool) {
	connection, err := upgrader.Upgrade(wr, req, nil)
	if err != nil {
		fmt.Printf("%s | %s\n", req.RemoteAddr, err)
		return
	}

	defer connection.Close()
	fmt.Printf("%s | upgraded to websocket\n", req.RemoteAddr)

	var message []byte

	if sendServerHostname {
		host, err := os.Hostname()
		if err == nil {
			message = []byte(fmt.Sprintf("Request served by %s", host))
		} else {
			message = []byte(fmt.Sprintf("Server hostname unknown: %s", err.Error()))
		}
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

//go:embed "html"
var files embed.FS

func serveFrontend(wr http.ResponseWriter, req *http.Request) {
	const templateName = "html/frontend.tmpl.html"
	tmpl, err := template.ParseFS(files, templateName)
	if err != nil {
		http.Error(wr, err.Error(), http.StatusInternalServerError)
		return
	}
	templateData := struct {
		Path string
	}{
		Path: path.Join(
			os.Getenv("WEBSOCKET_ROOT"),
			path.Dir(req.URL.Path),
		),
	}
	err = tmpl.Execute(wr, templateData)
	if err != nil {
		http.Error(wr, err.Error(), http.StatusInternalServerError)
		return
	}
	wr.Header().Add("Content-Type", "text/html")
	wr.WriteHeader(200)
}

func serveHTTP(wr http.ResponseWriter, req *http.Request, sendServerHostname bool) {
	// Take in extra headers to present
	addheaders := os.Getenv("ADD_HEADERS")
	if len(addheaders) > 0 {
		// rawjson := json.RawMessage(addheaders)
		var headerjson map[string]interface{}
		json.Unmarshal([]byte(addheaders), &headerjson)
		for k, v := range headerjson {
			//fmt.Printf("%s = %s", k, v)
			wr.Header().Add(k, v.(string))
		}
	}

	wr.Header().Add("Content-Type", "text/plain")
	wr.WriteHeader(200)

	// -> Intro

	fmt.Fprintln(wr, "Welcome to echo-server!  Here's what I know.")
	fmt.Fprintf(wr, "  > Head to /ws for interactive websocket echo!\n\n")
	// -> Hostname

	host, err := os.Hostname()
	if err == nil {
		fmt.Fprintf(wr, "-> My hostname is: %s\n\n", host)
	} else {
		fmt.Fprintf(wr, "-> Server hostname unknown: %s\n\n", err.Error())
	}

	// -> Pod Details

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

	// Requesting/Source IP

	fmt.Fprintf(wr, "-> Requesting IP: %s\n\n", req.RemoteAddr)

	// -> TLS Info
	if req.TLS != nil {
		fmt.Fprintln(wr, "-> TLS Connection Info | \n ")
		fmt.Fprintf(wr, "  %+v\n\n", req.TLS)
	}

	// -> Request Header

	fmt.Fprintln(wr, "-> Request Headers | \n ")
	fmt.Fprintf(wr, "  %s %s %s\n", req.Proto, req.Method, req.URL)
	fmt.Fprintf(wr, "\n")
	fmt.Fprintf(wr, "  Host: %s\n", req.Host)

	var reqheaders []string
	for k, vs := range req.Header {
		for _, v := range vs {
			reqheaders = append(reqheaders, (fmt.Sprintf("%s: %s", k, v)))
		}
	}

	if sendServerHostname {
		host, err := os.Hostname()
		if err == nil {
			fmt.Fprintf(wr, "Request served by %s\n\n", host)
		} else {
			fmt.Fprintf(wr, "Server hostname unknown: %s\n\n", err.Error())
		}
	}
	sort.Strings(reqheaders)
	for _, h := range reqheaders {
		fmt.Fprintf(wr, "  %s\n", h)
	}

	// -> Response Headers

	fmt.Fprintf(wr, "\n\n")
	fmt.Fprintln(wr, "-> Response Headers | \n ")
	var respheaders []string
	for k, vs := range wr.Header() {
		for _, v := range vs {
			respheaders = append(respheaders, (fmt.Sprintf("%s: %s", k, v)))
		}
	}
	sort.Strings(respheaders)
	for _, h := range respheaders {
		fmt.Fprintf(wr, "  %s\n", h)
	}
	fmt.Fprintln(wr, "\n  > Note that you may also see \"Transfer-Encoding\" and \"Date\"!")

	// -> Environment

	fmt.Fprintf(wr, "\n\n")
	fmt.Fprintln(wr, "-> My environment |")
	envs := os.Environ()
	sort.Strings(envs)
	for _, e := range envs {
		pair := strings.Split(e, "=")
		fmt.Fprintf(wr, "  %s=%s\n", pair[0], pair[1])
	}

	// Lets get resolv.conf
	fmt.Fprintf(wr, "\n\n")
	resolvfile, err := ioutil.ReadFile("/etc/resolv.conf") // just pass the file name
	if err != nil {
		fmt.Fprintf(wr, "%s", err)
	}

	str := string(resolvfile) // convert content to a 'string'

	fmt.Fprintf(wr, "-> Contents of /etc/resolv.conf | \n%s\n", str) // print the content as a 'string'

	// Lets get hosts
	fmt.Fprintf(wr, "\n")
	hostsfile, err := ioutil.ReadFile("/etc/hosts") // just pass the file name
	if err != nil {
		fmt.Fprintf(wr, "%s", err)
	}

	hostsstr := string(hostsfile) // convert content to a 'string'

	fmt.Fprintf(wr, "-> Contents of /etc/hosts | \n%s\n\n", hostsstr) // print the content as a 'string'

	fmt.Fprintln(wr, "")
	curtime := time.Now().UTC()
	fmt.Fprintln(wr, "-> And that's the way it is", curtime)
	fmt.Fprintln(wr, "\n// Thanks for using echo-server, a project by Mario Loria (InAnimaTe).")
	fmt.Fprintln(wr, "// https://github.com/inanimate/echo-server")
	fmt.Fprintln(wr, "// https://hub.docker.com/r/inanimate/echo-server")
	io.Copy(wr, req.Body)
	writeRequest(wr, req)
}

func serveSSE(wr http.ResponseWriter, req *http.Request, sendServerHostname bool) {
	if _, ok := wr.(http.Flusher); !ok {
		http.Error(wr, "Streaming unsupported!", http.StatusInternalServerError)
		return
	}

	var echo strings.Builder
	writeRequest(&echo, req)

	wr.Header().Set("Content-Type", "text/event-stream")
	wr.Header().Set("Cache-Control", "no-cache")
	wr.Header().Set("Connection", "keep-alive")
	wr.Header().Set("Access-Control-Allow-Origin", "*")

	var id int

	// Write an event about the server that is serving this request.
	if sendServerHostname {
		if host, err := os.Hostname(); err == nil {
			writeSSE(
				wr,
				req,
				&id,
				"server",
				host,
			)
		}
	}

	// Write an event that echoes back the request.
	writeSSE(
		wr,
		req,
		&id,
		"request",
		echo.String(),
	)

	// Then send a counter event every second.
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-req.Context().Done():
			return
		case t := <-ticker.C:
			writeSSE(
				wr,
				req,
				&id,
				"time",
				t.Format(time.RFC3339),
			)
		}
	}
}

// writeSSE sends a server-sent event and logs it to the console.
func writeSSE(
	wr http.ResponseWriter,
	req *http.Request,
	id *int,
	event, data string,
) {
	*id++
	writeSSEField(wr, req, "event", event)
	writeSSEField(wr, req, "data", data)
	writeSSEField(wr, req, "id", strconv.Itoa(*id))
	fmt.Fprintf(wr, "\n")
	wr.(http.Flusher).Flush()
}

// writeSSEField sends a single field within an event.
func writeSSEField(
	wr http.ResponseWriter,
	req *http.Request,
	k, v string,
) {
	for _, line := range strings.Split(v, "\n") {
		fmt.Fprintf(wr, "%s: %s\n", k, line)
		fmt.Printf("%s | sse | %s: %s\n", req.RemoteAddr, k, line)
	}
}

// writeRequest writes request headers to w.
func writeRequest(w io.Writer, req *http.Request) {
	fmt.Fprintf(w, "%s %s %s\n", req.Method, req.URL, req.Proto)
	fmt.Fprintln(w, "")

	fmt.Fprintf(w, "Host: %s\n", req.Host)
	printHeaders(w, req.Header)

	var body bytes.Buffer
	io.Copy(&body, req.Body) // nolint:errcheck

	if body.Len() > 0 {
		fmt.Fprintln(w, "")
		body.WriteTo(w) // nolint:errcheck
	}
}

func printHeaders(w io.Writer, h http.Header) {
	sortedKeys := make([]string, 0, len(h))

	for key := range h {
		sortedKeys = append(sortedKeys, key)
	}

	sort.Strings(sortedKeys)

	for _, key := range sortedKeys {
		for _, value := range h[key] {
			fmt.Fprintf(w, "%s: %s\n", key, value)
		}
	}
}
