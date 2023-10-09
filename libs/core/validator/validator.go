package validator

import (
	"fmt"
	"reflect"
	"regexp"

	"strings"

	engine "github.com/go-playground/validator/v10"
	"github.com/xigxog/kubefox/libs/core/kubefox"
	"github.com/xigxog/kubefox/libs/core/logkf"
)

type NameRegexp interface {
	GetName() string
	GetNameRegExp() *regexp.Regexp
	IsNameValid() bool
}

type Validator struct {
	validate *engine.Validate
	log      *logkf.Logger
}

type ValidationError struct {
	Field    string `json:"field"`
	Tag      string `json:"tag"`
	Type     string `json:"type"`
	Value    any    `json:"value"`
	Expected string `json:"expected,omitempty"`
}

func New(log *logkf.Logger) *Validator {
	v := &Validator{
		validate: engine.New(),
		log:      log,
	}

	v.validate.RegisterValidation("name", func(fl engine.FieldLevel) bool {
		return v.validateRegexp(fl, kubefox.NameRegexp)
	})
	v.validate.RegisterValidation("commit", func(fl engine.FieldLevel) bool {
		return v.validateRegexp(fl, kubefox.CommitRegexp)
	})
	v.validate.RegisterValidation("componentImage", func(fl engine.FieldLevel) bool {
		return v.validateRegexp(fl, kubefox.ImageRegexp)
	})

	return v
}

func (v *Validator) Validate(s any) (errs []*ValidationError) {
	if obj, ok := s.(NameRegexp); ok {
		if !obj.IsNameValid() {
			errs = append(errs, &ValidationError{
				Field: "name",
				Tag:   "regexp: " + obj.GetNameRegExp().String(),
				Type:  "string",
				Value: obj.GetName(),
			})
		}
	}

	err := v.validate.Struct(s)

	switch errv := err.(type) {
	case *engine.InvalidValidationError:
		v.log.Errorf("error validating struct: %v", err)
		errs = append(errs, &ValidationError{
			Field: "_ERROR_",
			Type:  "_ERROR_",
			Value: err.Error(),
		})

	case engine.ValidationErrors:
		for _, e := range errv {
			// ensure field name starts with lowercase
			parts := []string{}
			for _, p := range strings.Split(e.Namespace(), ".") {
				// remove inlined props
				if strings.HasSuffix(p, "Prop") || strings.HasSuffix(p, "Props") {
					continue
				}
				parts = append(parts, strings.ToLower(string(p[0]))+p[1:])
			}

			errs = append(errs, &ValidationError{
				Field:    strings.Join(parts, "."),
				Tag:      e.Tag(),
				Type:     e.Type().Name(),
				Value:    e.Value(),
				Expected: e.Param(),
			})
		}
	}

	return
}

func (v *Validator) validateRegexp(fl engine.FieldLevel, r *regexp.Regexp) bool {
	field := fl.Field()
	if field.Kind() != reflect.String {
		panic(fmt.Sprintf("Bad field type %T", field.Interface()))
	}

	return r.MatchString(field.String())
}
