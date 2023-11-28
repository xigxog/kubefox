package utils

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"unsafe"
)

var (
	RegexpNameSpecialChar  = regexp.MustCompile(`[^a-z0-9]`)
	RegexpLabelSpecialChar = regexp.MustCompile(`[^a-z0-9A-Z-_\.]`)
	RegexpLabelPrefix      = regexp.MustCompile(`^[^a-z0-9A-Z]*`)
	RegexpLabelSuffix      = regexp.MustCompile(`[^a-z0-9A-Z-_\.]*[^a-z0-9A-Z]*$`)
)

func ResolveFlag(curr, envVar, def string) string {
	if curr != "" {
		return curr
	}
	if e := os.Getenv(envVar); e != "" {
		return e
	} else {
		return def
	}
}

func ResolveFlagBool(curr bool, envVar string, def bool) bool {
	if curr != def {
		return curr
	}
	if e, err := strconv.ParseBool(os.Getenv(envVar)); err == nil {
		return e
	} else {
		return def
	}
}

func ResolveFlagInt(curr int, envVar string, def int) int {
	if curr != def {
		return curr
	}
	if e, err := strconv.ParseInt(os.Getenv(envVar), 10, 0); err == nil {
		return int(e)
	} else {
		return def
	}
}

func CheckRequiredFlag(n, p string) {
	if p == "" {
		fmt.Fprintf(os.Stderr, "The flag \"%s\" is required.\n\n", n)
		flag.Usage()
		os.Exit(1)
	}
}

func EnvDef(name, def string) string {
	e, _ := os.LookupEnv(name)
	if e == "" {
		return def
	}
	return e
}

func UIntToByteArray(i uint64) []byte {
	data := *(*[unsafe.Sizeof(i)]byte)(unsafe.Pointer(&i))
	return data[:]
}

func ByteArrayToUInt(b []byte) uint64 {
	return *(*uint64)(unsafe.Pointer(&b[0]))
}

func ShortCommit(commit string) string {
	if len(commit) < 7 {
		return ""
	}

	return commit[0:7]
}

// First returns the first non-empty string. If all strings are empty then empty
// string is returned.
func First(strs ...string) string {
	for _, s := range strs {
		if s != "" {
			return s
		}
	}

	return ""
}

// CleanName returns name with all special characters replaced with dashes and
// set to lowercase. If name is a path only the basename is used.
//
// https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#dns-subdomain-names
func CleanName(name string) string {
	cleaned := filepath.Base(name)
	cleaned = strings.ToLower(cleaned)
	cleaned = RegexpNameSpecialChar.ReplaceAllLiteralString(cleaned, "-")
	cleaned = strings.TrimPrefix(strings.TrimSuffix(cleaned, "-"), "-")
	return cleaned
}

func IsValidName(name string) bool {
	return name == CleanName(name)
}

// CleanLabel returns the label value with all special characters replaced with
// dashes and any character that is not [a-z0-9A-Z] trimmed from start and end.
// If name is a path only the basename is used.
//
// https://kubernetes.io/docs/concepts/overview/working-with-objects/labels/#syntax-and-character-set
func CleanLabel(value string) string {
	cleaned := filepath.Base(value)
	// Remove special chars.
	cleaned = RegexpLabelSpecialChar.ReplaceAllLiteralString(cleaned, "-")
	// Ensure value begins and ends with [a-z0-9A-Z].
	cleaned = RegexpLabelPrefix.ReplaceAllLiteralString(cleaned, "")
	cleaned = RegexpLabelSuffix.ReplaceAllLiteralString(cleaned, "")
	return cleaned
}
