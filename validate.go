package gwf

import (
	"fmt"
	"reflect"
	"strings"
	"sync"

	"github.com/go-playground/validator/v10"
)

var Validator StructValidator = &defaultValidator{}

type StructValidator interface {
	ValidateStruct(any) error
	Engine() any
}

type defaultValidator struct {
	once     sync.Once
	validate *validator.Validate
}

type SliceValidationErr []error

func (errs SliceValidationErr) Error() string {
	n := len(errs)
	switch n {
	case 0:
		return ""
	default:
		var b strings.Builder
		if errs[0] != nil {
			fmt.Fprintf(&b, "[d%]: %s", 0, errs[0].Error())
		}
		if n > 1 {
			for i := 1; i < n; i++ {
				if errs[i] != nil {
					b.WriteString("\n")
					fmt.Fprintf(&b, "[d%]: %s", i, errs[i].Error())
				}
			}
		}
		return b.String()
	}
}

func (d *defaultValidator) ValidateStruct(obj any) error {
	vf := reflect.ValueOf(obj)

	switch vf.Kind() {
	case reflect.Pointer:
		return d.ValidateStruct(vf.Elem().Interface())
	case reflect.Struct:
		return d.validateStruct(obj)
	case reflect.Slice, reflect.Array:
		count := vf.Len()
		sliceValidationErr := make(SliceValidationErr, 0)
		for i := 0; i < count; i++ {
			if err := d.validateStruct(vf.Index(i).Interface()); err != nil {
				sliceValidationErr = append(sliceValidationErr, err)
			}
		}
		return sliceValidationErr
	}
	return nil
}

func (d *defaultValidator) validateStruct(obj any) error {
	d.lazyInit()
	return d.validate.Struct(obj)
}

func (d *defaultValidator) Engine() any {
	d.lazyInit()
	return d.validate
}

func (d *defaultValidator) lazyInit() {
	d.once.Do(func() {
		d.validate = validator.New()
	})
}
