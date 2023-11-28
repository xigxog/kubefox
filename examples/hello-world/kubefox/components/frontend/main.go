package main

import (
	"fmt"
	"strings"

	"github.com/xigxog/kubefox/kit"
)

var (
	backend kit.Dependency
)

func main() {
	k := kit.New()

	backend = k.Component("backend")
	// TODO mark env var as unique
	k.Route("Path(`/{{.Env.subPath}}/hello`)", sayHello)

	k.Start()
}

func sayHello(k kit.Kontext) error {
	r, err := k.Req(backend).Send()
	if err != nil {
		return err
	}

	msg := fmt.Sprintf("👋 Hello %s!", r.Str())
	k.Log().Debug(msg)

	a := strings.ToLower(k.Header("accept"))
	switch {
	case strings.Contains(a, "application/json"):
		return k.Resp().SendJSON(map[string]any{"msg": msg})

	case strings.Contains(a, "text/html"):
		return k.Resp().SendHTML(fmt.Sprintf(html, msg))

	default:
		return k.Resp().SendStr(msg)
	}
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
