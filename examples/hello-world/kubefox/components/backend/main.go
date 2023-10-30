package main

import (
	"github.com/xigxog/kubefox/kit"
)

func main() {
	k := kit.New()
	k.Default(sayWho)
	k.Start()
}

func sayWho(k kit.Kontext) error {
	who := k.EnvDef("who", "World")
	k.Log().Debugf("The who is '%s'!", who)

	return k.Resp().SendStr(who)
}
