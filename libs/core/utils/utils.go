package utils

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"unsafe"
)

type Comparable[T any] interface {
	Equals(T) bool
}

func Contains[T Comparable[T]](s []T, e T) bool {
	for _, v := range s {
		if v.Equals(e) {
			return true
		}
	}
	return false
}

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

func UIntToByteArray(i uint64) []byte {
	data := *(*[unsafe.Sizeof(i)]byte)(unsafe.Pointer(&i))
	return data[:]
}

func ByteArrayToUInt(b []byte) uint64 {
	return *(*uint64)(unsafe.Pointer(&b[0]))
}
