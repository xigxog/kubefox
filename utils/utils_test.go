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

func TestCleanLabel(t *testing.T) {
	if c := CleanLabel("aaa"); c != "aaa" {
		t.Logf("got '%s', expected '%s'", c, "aaa")
		t.Fail()
	}
	if c := CleanLabel("AABBCC"); c != "AABBCC" {
		t.Logf("got '%s', expected '%s'", c, "AABBCC")
		t.Fail()
	}
	if c := CleanLabel("-aaa"); c != "aaa" {
		t.Logf("got '%s', expected '%s'", c, "aaa")
		t.Fail()
	}
	if c := CleanLabel("aaa-"); c != "aaa" {
		t.Logf("got '%s', expected '%s'", c, "aaa")
		t.Fail()
	}
	if c := CleanLabel("a&a%a"); c != "a-a-a" {
		t.Logf("got '%s', expected '%s'", c, "a-a-a")
		t.Fail()
	}
	if c := CleanLabel("#$aa$a"); c != "aa-a" {
		t.Logf("got '%s', expected '%s'", c, "aa-a")
		t.Fail()
	}
	if c := CleanLabel("blah/#$aa$a"); c != "aa-a" {
		t.Logf("got '%s', expected '%s'", c, "aa-a")
		t.Fail()
	}
	if c := CleanLabel("//asd//asd/#$aa$a()"); c != "aa-a" {
		t.Logf("got '%s', expected '%s'", c, "aa-a")
		t.Fail()
	}
}
