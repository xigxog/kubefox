package build

import (
	"runtime/debug"
	"strings"
)

// Injected at build time.
var (
	brokerCommit   string
	commit         string
	component      string
	date           string
	headRef        string
	httpsrvCommit  string
	operatorCommit string
	rootCommit     string
	tagRef         string
)

type BuildInfo struct {
	Branch         string `json:"branch,omitempty"`
	BrokerCommit   string `json:"brokerCommit,omitempty"`
	Commit         string `json:"commit,omitempty"`
	Component      string `json:"component,omitempty"`
	Date           string `json:"date,omitempty"`
	HTTPSrvCommit  string `json:"httpsrvCommit,omitempty"`
	OperatorCommit string `json:"operatorCommit,omitempty"`
	RootCommit     string `json:"rootCommit,omitempty"`
	Tag            string `json:"tag,omitempty"`
	Version        string `json:"version,omitempty"`
}

// Constructed inside init.
var Info BuildInfo

func init() {
	// Find Version
	var modVersion, version string
	if info, ok := debug.ReadBuildInfo(); ok {
		if v := info.Main.Version; v != "" && v != "(devel)" {
			modVersion = v
		}
	}

	if p := strings.Split(tagRef, "/"); tagRef != "" {
		version = p[len(p)-1]
	} else if modVersion != "" {
		version = modVersion
	} else if p := strings.Split(headRef, "/"); headRef != "" {
		version = p[len(p)-1]
	} else {
		version = rootCommit
	}

	// Construct Info
	Info = BuildInfo{
		Branch:         headRef,
		BrokerCommit:   brokerCommit,
		Commit:         commit,
		Component:      component,
		Date:           date,
		HTTPSrvCommit:  httpsrvCommit,
		OperatorCommit: operatorCommit,
		RootCommit:     rootCommit,
		Tag:            tagRef,
		Version:        version,
	}
}
