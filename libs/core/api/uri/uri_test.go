package uri

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"
)

const (
	testPath = "system/test/id/aaaaaaa"
	testURI  = "kubefox://acme/" + testPath
	testJSON = `{"uri":"` + testURI + `"}`
)

type testStruct struct {
	U URI `json:"uri"`
}

func TestURI_UnmarshalJSON(t *testing.T) {
	ts := &testStruct{U: &URIType{}}
	if err := json.Unmarshal([]byte(testJSON), ts); err != nil {
		t.Fatalf("%v", err)
	}

	s := fmt.Sprintf("%v", ts)
	t.Log(s)
	if s != "&{"+testPath+"}" {
		t.Fail()
	}
}

func TestURI_MarshalJSON(t *testing.T) {
	uri, err := Parse(testURI)
	if err != nil {
		t.Fatalf("%v", err)
	}

	b, err := json.Marshal(&testStruct{U: uri})
	if err != nil {
		t.Fatalf("%v", err)
	}

	s := string(b)
	t.Log(s)
	if s != testJSON {
		t.Fail()
	}
}

func TestParse(t *testing.T) {
	testParse(t, testURI, true)
	testParse(t, "kubefox://acme/system/demo/tag/v1", true)
	testParse(t, "kubefox://acme/system/demo/tag", true)
	testParse(t, "kubefox://acme/system/demo/branch/v1", true)
	testParse(t, "kubefox://acme/system/demo/branch", true)
	testParse(t, "kubefox://acme/system/demo/metadata", true)
	testParse(t, "kubefox://acme/system/demo", true)
	testParse(t, "kubefox://acme/system/", true)
	testParse(t, "kubefox://acme/system", true)

	testParse(t, "", false)
	testParse(t, "::", false)
	testParse(t, "kubefox", false)                                     // invalid scheme
	testParse(t, "kubefox:", false)                                    // missing org
	testParse(t, "kubefox://", false)                                  // missing org
	testParse(t, "kubefox:///", false)                                 // missing org
	testParse(t, "kubefox://BOOGERS/", false)                          // invalid name
	testParse(t, "kubefox://acme/", false)                             // missing kind
	testParse(t, "kubefox://acme/boogers", false)                      // invalid kind
	testParse(t, "kubefox://acme//name", false)                        // missing kind
	testParse(t, "kubefox://acme/system/NAME", false)                  // invalid name
	testParse(t, "kubefox://acme/system/name/boogers", false)          // invalid subKind
	testParse(t, "kubefox://acme/system/name//subpath", false)         // missing subKind
	testParse(t, "kubefox://acme/system/name/id/boogers", false)       // invalid id
	testParse(t, "kubefox://acme/environment/name/id/boogers", false)  // invalid id
	testParse(t, "kubefox://acme/system/name/tag/Boogers", false)      // invalid name
	testParse(t, "kubefox://acme/system/name/tag/-boogers", false)     // invalid name
	testParse(t, "kubefox://acme/system/name/tag/v1/alpha", false)     // extra part
	testParse(t, "kubefox://acme/system/demo/metadata/boogers", false) // extra part
}

func TestParse_Platform(t *testing.T) {
	testParse(t, "kubefox://acme/platform/dev/deployment/demo/tag/v1", true)
	testParse(t, "kubefox://acme/platform/dev/deployment/demo", true)
	testParse(t, "kubefox://acme/platform/dev/deployment", true)

	testParse(t, "kubefox://acme/platform/dev/release/demo/dev", true)
	testParse(t, "kubefox://acme/platform/dev/release/demo", true)
	testParse(t, "kubefox://acme/platform/dev/release", true)

	testParse(t, "kubefox://acme/platform/dev", true)
	testParse(t, "kubefox://acme/platform", true)

	testParse(t, "kubefox://acme/platform/localplatform/deployment/demo/tag/v1/boogers", false)
	testParse(t, "kubefox://acme/platform/localplatform/deployment/demo/tag", false)
	testParse(t, "kubefox://acme/platform/dev/release/demo/dev/boogers", false)
}

func testParse(t *testing.T, u string, shouldPass bool) {
	t.Logf("🧪 testing '%s'", u)
	parsed, err := Parse(u)
	if err != nil {
		if shouldPass {
			t.Logf("😞 failed: unexpected error: %v", err)
			t.FailNow()
		} else {
			t.Logf("💥 expected error: %v", err)
		}
	}

	u = strings.Trim(u, PathSeparator)
	passed := parsed != nil && u == parsed.URL()
	if passed && !shouldPass {
		t.Logf("😞 failed: expected error but got '%s", parsed)
		t.FailNow()
	} else if !passed && shouldPass {
		t.Logf("😞 failed: expected '%s' but got '%s'", u, parsed)
		t.FailNow()
	}
}
