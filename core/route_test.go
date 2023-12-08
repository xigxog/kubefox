package core

import (
	"testing"

	"github.com/xigxog/kubefox/api"
)

func TestRoute_Resolve(t *testing.T) {
	r := Route{
		RouteSpec: api.RouteSpec{
			Id:       0,
			Rule:     "Path(`/{{unique}}/{{testVar2 | unique}}/{{testVar}}/{{intVar}}/{{floatVar}}/{{arrNum}}/{{arrStr}}`)",
			Priority: 0,
		},
	}

	err := r.Resolve(map[string]*api.Val{
		"testVar":  api.ValString("testVarValue"),
		"testVar2": api.ValString("testVarValue2"),
		"intVar":   api.ValInt(123),
		"floatVar": api.ValFloat(1.1),
		"arrNum":   api.ValArrayFloat([]float64{1.1, 2}),
		"arrStr":   api.ValArrayString([]string{"a", "b"}),
		"unique":   api.ValString("imunique"),
	})
	if err != nil {
		t.Error(err)
	}

	t.Logf("resolved: %s", r.ResolvedRule)
	expected := "Path(`/imunique/testVarValue2/testVarValue/123/1.1/{^1\\.1$|^2$}/{^a$|^b$}`)"
	if r.ResolvedRule != expected {
		t.Fail()
	}
}
