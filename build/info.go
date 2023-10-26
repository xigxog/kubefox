package build

import (
	"runtime/debug"
	"strings"
)

// Injected at build time.
var (
	BuildDate  string
	Component  string
	Commit     string
	RootCommit string
	HeadRef    string
	TagRef     string
)

// Set inside init.
var (
	Version    string
	VersionMap map[string]string
)

func init() {
	// Set Version.
	var modVersion string
	if info, ok := debug.ReadBuildInfo(); ok {
		if v := info.Main.Version; v != "" && v != "(devel)" {
			modVersion = v
		}
	}

	if p := strings.Split(TagRef, "/"); len(p) > 0 {
		Version = p[len(p)-1]
	} else if modVersion != "" {
		Version = modVersion
	} else if p := strings.Split(HeadRef, "/"); len(p) > 0 {
		Version = p[len(p)-1]
	} else {
		Version = RootCommit
	}

	// Set VersionMap.
	VersionMap = make(map[string]string)
	if BuildDate != "" {
		VersionMap["date"] = BuildDate
	}
	if Component != "" {
		VersionMap["component"] = Component
	}
	if Commit != "" {
		VersionMap["commit"] = Commit
	}
	if RootCommit != "" {
		VersionMap["rootCommit"] = RootCommit
	}
	if HeadRef != "" {
		VersionMap["branch"] = HeadRef
	}
	if TagRef != "" {
		VersionMap["tag"] = TagRef
	}
	if Version != "" {
		VersionMap["version"] = Version
	}
}
