package core

import "io"

// Copyright 2009 The Go Authors. All rights reserved.
// https://cs.opensource.google/go/go/+/refs/tags/go1.20:src/io/io.go;l=456
// https://cs.opensource.google/go/go/+/refs/tags/go1.20:LICENSE
//
// LimitedReader reads from R but limits the amount of data returned to just N
// bytes. Each call to Read updates N to reflect the new amount remaining. Read
// returns ErrContentTooLarge when N <= 0.
//
// This is a variation of io.LimitedReader that returns a different err for
// exceeding limit and EOF.
type LimitedReader struct {
	R io.Reader // underlying reader
	N int64     // max bytes remaining
}

func (l *LimitedReader) Read(p []byte) (n int, err error) {
	if l.N <= 0 {
		return 0, ErrContentTooLarge()
	}
	if int64(len(p)) > l.N {
		p = p[0:l.N]
	}
	n, err = l.R.Read(p)
	l.N -= int64(n)
	return
}
