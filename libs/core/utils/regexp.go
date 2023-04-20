package utils

import (
	"regexp"
)

var (
	HashRegexp  = regexp.MustCompile(`^[0-9a-f]{7}$`)
	ImageRegexp = regexp.MustCompile(`^.*:[a-z0-9-]{7}$`)
	// TODO use SHA256, switch pattern to ^.*@sha256:[a-z0-9]{64}$
	NameRegexp        = regexp.MustCompile(`^[a-z0-9][a-z0-9-]{0,28}[a-z0-9]$`)
	TagOrBranchRegexp = regexp.MustCompile(`^[a-z0-9][a-z0-9-\\.]{0,28}[a-z0-9]$`)
	UUIDRegexp        = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)
)
