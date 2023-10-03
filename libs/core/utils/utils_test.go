package utils

import (
	"bytes"
	"testing"
)

func TestUIntToByteArray(t *testing.T) {
	i := uint64(1234567890)
	b := UIntToByteArray(i)

	exp := []byte{210, 2, 150, 73, 0, 0, 0, 0}
	if !bytes.Equal(b, exp) {
		t.Fail()
	}
}

func TestByteArrayToUInt(t *testing.T) {
	b := []byte{210, 2, 150, 73, 0, 0, 0, 0}
	i := ByteArrayToUInt(b)

	exp := uint64(1234567890)
	if i != exp {
		t.Fail()
	}
}
