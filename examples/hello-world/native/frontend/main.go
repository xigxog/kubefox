package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

var (
	addr string
)

func main() {
	flag.StringVar(&addr, "addr", "127.0.0.1:3333", "address http server should bind to")
	flag.Parse()

	subPath := os.Getenv("subPath")
	if subPath == "" {
		subPath = "unknown"
	}
	path := fmt.Sprintf("/%s/hello", subPath)

	fmt.Printf("starting http server on '%s' for path '%s'...\n", addr, path)

	http.HandleFunc(path, sayHello)
	err := http.ListenAndServe(addr, nil)
	if errors.Is(err, http.ErrServerClosed) {
		fmt.Println("server closed")
	} else if err != nil {
		fmt.Printf("error starting server: %s\n", err)
		os.Exit(1)
	}
}

func sayHello(respWriter http.ResponseWriter, req *http.Request) {
	resp, err := http.DefaultClient.Get("http://hello-world-backend/")
	if err != nil {
		writeErr(respWriter, err)
		return
	}
	defer resp.Body.Close()

	who, err := io.ReadAll(resp.Body)
	if err != nil {
		writeErr(respWriter, err)
		return
	}

	msg := fmt.Sprintf("👋 Hello %s!", who)
	fmt.Println(msg)

	var body []byte
	a := strings.ToLower(resp.Header.Get("accept"))
	switch {
	case strings.Contains(a, "application/json"):
		body, err = json.Marshal(map[string]any{"msg": msg})
		if err != nil {
			writeErr(respWriter, err)
			return
		}

	case strings.Contains(a, "text/html"):
		body = []byte(fmt.Sprintf(html, msg))

	default:
		body = []byte(msg)
	}

	respWriter.Write(body)
}

func writeErr(respWriter http.ResponseWriter, err error) {
	respWriter.WriteHeader(http.StatusInternalServerError)
	respWriter.Write([]byte(err.Error()))
}

const html = `
<!DOCTYPE html>
<html>
  <head>
    <meta charset="UTF-8" />
    <title>Hello KubeFox</title>
    <style>
      html,
      body,
      p {
        height: 100%%;
        margin: 0;
      }
      .container {
        display: flex;
        flex-direction: column;
        min-height: 80%%;
        align-items: center;
        justify-content: center;
      }
    </style>
  </head>
  <body>
    <main class="container">
      <h1>%s</h1>
    </main>
  </body>
</html>
`
