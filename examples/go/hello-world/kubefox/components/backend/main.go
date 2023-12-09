package main

import (
	"github.com/xigxog/kubefox/kit"
	"github.com/xigxog/kubefox/kit/env"
)

var (
	who kit.EnvVarDep
)

func main() {
	k := kit.New()

	who = k.EnvVar("who", env.Required)
	k.Default(sayWho)

	k.Start()
}

func sayWho(k kit.Kontext) error {
	who := k.EnvDef(who, "World")
	k.Log().Debugf("The who is '%s'!", who)

	return k.Resp().SendStr(who)
}
