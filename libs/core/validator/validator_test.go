package validator

import (
	"encoding/json"
	"testing"

	"github.com/xigxog/kubefox/libs/core/api/admin/v1alpha1"
	"github.com/xigxog/kubefox/libs/core/api/common"
	"github.com/xigxog/kubefox/libs/core/api/maker"
	"github.com/xigxog/kubefox/libs/core/logger"
)

var v = New(&logger.Log{})

func TestValidator_Validate(t *testing.T) {
	o := maker.Empty[v1alpha1.System]()
	o.Id = "aaaaaaa"
	o.GitRepo = "https://github.com/xigxog/kubefox/demo-system"
	o.GitHash = "aaaaaaa"
	o.GitRef = "branch/main"

	c := &common.AppComponent{}
	c.Type = "kubefox"
	c.GitHash = "aaaaaaa"
	c.Image = "ghcr.io/kubefox/comp1:aaaaaaa"

	a := &common.App{}
	a.Components = map[string]*common.AppComponent{"test": c}
	o.Apps = map[string]*common.App{"test": a}

	errs := v.Validate(o)

	if len(errs) != 0 {
		s, _ := json.MarshalIndent(errs, "", "  ")
		t.Log(string(s))

		t.Fatalf("%d unexpected error(s)", len(errs))
	}
}
