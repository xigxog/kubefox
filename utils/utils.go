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

var RegexpSpecialChar = regexp.MustCompile(`[^a-z0-9]`)

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

// CleanName returns name with all special characters replaced with dashes and
// set to lowercase. If name is a path only the basename is used.
func CleanName(name string) string {
	cleaned := filepath.Base(name)
	cleaned = strings.ToLower(cleaned)
	cleaned = RegexpSpecialChar.ReplaceAllLiteralString(cleaned, "-")
	cleaned = strings.TrimPrefix(strings.TrimSuffix(cleaned, "-"), "-")
	return cleaned
}
