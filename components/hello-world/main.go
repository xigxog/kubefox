package main

import (
	"fmt"

	"github.com/xigxog/kubefox/libs/core/kit"
)

func main() {
	kit := kit.New()

	kit.Route("Path(`/{{.Env.host}}/hello`)", hello)

	kit.Start()
}

func hello(k kit.Kontext) error {
	k.Log().Debug("Hey I got a request!")

	who := k.ParamOrDefault("who", "world")
	host := k.Env("host")
	s := fmt.Sprintf("Hello %s on host %s! 👋", who, host)

	return k.String(s)
}
