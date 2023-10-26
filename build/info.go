package build

import (
	"runtime/debug"
	"strings"
)

// Injected at build time.
var (
	date       string
	component  string
	commit     string
	rootCommit string
	headRef    string
	tagRef     string
)

type BuildInfo struct {
	Date       string `json:"date,omitempty"`
	Component  string `json:"component,omitempty"`
	Commit     string `json:"commit,omitempty"`
	RootCommit string `json:"rootCommit,omitempty"`
	HeadRef    string `json:"headRef,omitempty"`
	TagRef     string `json:"tagRef,omitempty"`
	Version    string `json:"version,omitempty"`
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
		Date:       date,
		Component:  component,
		Commit:     commit,
		RootCommit: rootCommit,
		HeadRef:    headRef,
		TagRef:     tagRef,
		Version:    version,
	}
}
