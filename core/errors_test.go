package core

import (
	"encoding/json"
	"errors"
	"fmt"
	"testing"
)

func TestErrors_Error(t *testing.T) {
	err := ErrBrokerMismatch()
	errWC := ErrBrokerMismatch(errors.New("this is the cause"))

	wrapped := fmt.Errorf("wrapped: %w", ErrBrokerMismatch())
	unwrap := &Err{}
	if ok := errors.As(wrapped, &unwrap); !ok {
		t.FailNow()
	}

	if err.HTTPCode() != unwrap.HTTPCode() {
		t.Fail()
	}

	t.Logf("\nerr\n%s\n\nunwrap2\n%s\n\nerrWC\n%s", err, unwrap, errWC)

	if err == unwrap {
		t.Fail()
	}
}

func TestErrors_JSON(t *testing.T) {
	testErr := ErrBrokerMismatch(errors.New("this is the cause"))
	b, err := json.Marshal(testErr)
	if err != nil {
		t.Log(err)
		t.FailNow()
	}

	if string(b) != `{"grpcCode":9,"httpCode":502,"msg":"broker mismatch"}` {
		t.Fail()
	}

	t.Log(string(b))

	e := &Err{}
	if err := json.Unmarshal(b, e); err != nil {
		t.Log(err)
		t.FailNow()
	}
	if e.HTTPCode() != testErr.HTTPCode() {
		t.Fail()
	}
}
