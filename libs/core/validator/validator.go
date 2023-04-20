package validator

import (
	"fmt"
	"reflect"
	"regexp"

	"strings"

	engine "github.com/go-playground/validator/v10"
	"github.com/xigxog/kubefox/libs/core/api/uri"
	"github.com/xigxog/kubefox/libs/core/logger"
	"github.com/xigxog/kubefox/libs/core/utils"
)

type NameRegexp interface {
	GetName() string
	GetNameRegExp() *regexp.Regexp
	IsNameValid() bool
}

type Validator struct {
	validate *engine.Validate
	log      *logger.Log
}

type ValidationError struct {
	Field    string `json:"field"`
	Tag      string `json:"tag"`
	Type     string `json:"type"`
	Value    any    `json:"value"`
	Expected string `json:"expected,omitempty"`
}

func New(log *logger.Log) *Validator {
	v := &Validator{
		validate: engine.New(),
		log:      log,
	}

	v.validate.RegisterValidation("objName", func(fl engine.FieldLevel) bool {
		return v.validateRegexp(fl, utils.NameRegexp)
	})
	v.validate.RegisterValidation("gitHash", func(fl engine.FieldLevel) bool {
		return v.validateRegexp(fl, utils.HashRegexp)
	})
	v.validate.RegisterValidation("componentImage", func(fl engine.FieldLevel) bool {
		return v.validateRegexp(fl, utils.ImageRegexp)
	})
	v.validate.RegisterValidation("configRef", func(fl engine.FieldLevel) bool {
		_, valid := v.validateURI(fl, uri.Config)
		return valid
	})
	v.validate.RegisterValidation("environmentRef", func(fl engine.FieldLevel) bool {
		_, valid := v.validateURI(fl, uri.Environment)
		return valid
	})
	v.validate.RegisterValidation("environmentIdRef", func(fl engine.FieldLevel) bool {
		u, valid := v.validateURI(fl, uri.Environment)
		if u != nil {
			valid = valid && u.SubKind() == uri.Id
		}
		return valid
	})
	v.validate.RegisterValidation("systemRef", func(fl engine.FieldLevel) bool {
		_, valid := v.validateURI(fl, uri.System)
		return valid
	})
	v.validate.RegisterValidation("systemIdRef", func(fl engine.FieldLevel) bool {
		u, valid := v.validateURI(fl, uri.System)
		if u != nil {
			valid = valid && u.SubKind() == uri.Id
		}
		return valid
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

	switch err.(type) {
	case *engine.InvalidValidationError:
		v.log.Errorf("error validating struct: %v", err)
		errs = append(errs, &ValidationError{
			Field: "_ERROR_",
			Type:  "_ERROR_",
			Value: err.Error(),
		})

	case engine.ValidationErrors:
		for _, e := range err.(engine.ValidationErrors) {
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

func (v *Validator) validateURI(fl engine.FieldLevel, kind uri.Kind) (uri.URI, bool) {
	field := fl.Field()
	if field.Kind() != reflect.String {
		panic(fmt.Sprintf("Bad field type %T", field.Interface()))
	}

	u, err := uri.New("validator", kind, field)
	if err != nil {
		v.log.Debugf("uri is invalid: %v", err)
		return nil, false
	}

	return u, true
}
