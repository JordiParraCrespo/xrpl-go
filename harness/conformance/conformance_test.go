package conformance

import (
	"reflect"
	"strings"
	"testing"
)

// fieldByJSONTag finds the struct field (searching embedded structs) whose json
// tag (the name before any comma) equals tag, and returns a settable reflect
// Value for it. The bool reports whether the field was found.
func fieldByJSONTag(v reflect.Value, tag string) (reflect.Value, bool) {
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	if v.Kind() != reflect.Struct {
		return reflect.Value{}, false
	}
	t := v.Type()
	for i := 0; i < t.NumField(); i++ {
		sf := t.Field(i)
		name := strings.Split(sf.Tag.Get("json"), ",")[0]
		if name == tag {
			return v.Field(i), true
		}
		// Recurse into embedded (anonymous) structs to reach promoted fields.
		if sf.Anonymous {
			if fv, ok := fieldByJSONTag(v.Field(i), tag); ok {
				return fv, true
			}
		}
	}
	return reflect.Value{}, false
}

// TestConformance is the core static conformance harness. For every registered
// case it verifies the valid instance passes Validate(), then for each field
// xrpl.js marks required it zeroes that field and asserts Validate() rejects it.
func TestConformance(t *testing.T) {
	if len(cases) == 0 {
		t.Fatal("no conformance cases registered")
	}

	for _, c := range cases {
		c := c
		t.Run(c.Name, func(t *testing.T) {
			// 1. The valid instance must validate cleanly.
			if err := c.Valid().Validate(); err != nil {
				t.Fatalf("valid %s instance failed Validate(): %v", c.Name, err)
			}

			if c.JSInterface == "" {
				return
			}

			spec, ok := jsSpec[c.JSInterface]
			if !ok {
				t.Fatalf("xrpl.js interface %q not found in jsspec.json", c.JSInterface)
			}

			skip := make(map[string]bool, len(c.SkipRequired))
			for _, s := range c.SkipRequired {
				skip[s] = true
			}

			// 2. Each xrpl.js-required field must be enforced by Go's Validate().
			for _, field := range spec.Required {
				if skip[field] {
					continue
				}
				t.Run("requires/"+field, func(t *testing.T) {
					inst := c.Valid()
					fv, ok := fieldByJSONTag(reflect.ValueOf(inst), field)
					if !ok {
						t.Fatalf("%s has no field with json tag %q (required by xrpl.js %s)",
							c.Name, field, c.JSInterface)
					}
					if !fv.CanSet() {
						t.Fatalf("field %q on %s is not settable", field, c.Name)
					}
					fv.Set(reflect.Zero(fv.Type()))
					if err := inst.Validate(); err == nil {
						t.Errorf("%s.Validate() accepted a request missing required field %q "+
							"(xrpl.js %s marks it required)", c.Name, field, c.JSInterface)
					}
				})
			}
		})
	}
}
