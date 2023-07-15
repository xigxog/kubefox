package main

import (
	"github.com/xigxog/kubefox/components/broker/cmd"
	"github.com/xigxog/kubefox/components/broker/config"
)

// Injected at build time
var (
	GitRef  string
	GitHash string
)

func main() {
	config.GitRef = GitRef
	config.GitHash = GitHash

	cmd.Execute()
}
