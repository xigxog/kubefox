package main

import (
	"errors"
	"flag"
	"fmt"
	"net/http"
	"os"
)

var (
	who  string
	addr string
)

func main() {
	who = os.Getenv("who")
	if who == "" {
		who = "World"
	}

	flag.StringVar(&addr, "addr", "127.0.0.1:3333", "address http server should bind to")
	flag.Parse()

	http.HandleFunc("/", sayWho)

	fmt.Printf("starting http server on '%s'...\n", addr)
	err := http.ListenAndServe(addr, nil)
	if errors.Is(err, http.ErrServerClosed) {
		fmt.Println("server closed")
	} else if err != nil {
		fmt.Printf("error starting server: %s\n", err)
		os.Exit(1)
	}
}

func sayWho(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("The who is '%s'", who)
	w.Write([]byte(who))
}
