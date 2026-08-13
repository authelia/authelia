package model_test

import (
	"maps"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/authelia/authelia/v4/internal/authentication"
	"github.com/authelia/authelia/v4/internal/model"
)

func TestUserShouldNotDriftFromUserDetailsExtended(t *testing.T) {
	testCases := []struct {
		name     string
		expected reflect.Type
		actual   reflect.Type
	}{
		{
			"ShouldMatchUserDetailsExtended",
			reflect.TypeOf(authentication.UserDetailsExtended{}),
			reflect.TypeOf(model.User{}),
		},
		{
			"ShouldMatchUserDetailsAddress",
			reflect.TypeOf(authentication.UserDetailsAddress{}),
			reflect.TypeOf(model.UserAddress{}),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, flattenTypeFields(tc.expected), flattenTypeFields(tc.actual))
		})
	}
}

func flattenTypeFields(t reflect.Type) map[string]string {
	fields := map[string]string{}

	for i := range t.NumField() {
		field := t.Field(i)

		ft := field.Type

		if field.Anonymous {
			if ft.Kind() == reflect.Pointer {
				ft = ft.Elem()
			}

			if ft.Kind() == reflect.Struct {
				maps.Copy(fields, flattenTypeFields(ft))

				continue
			}
		}

		fields[field.Name] = normalizeTypeFieldName(ft)
	}

	return fields
}

func normalizeTypeFieldName(t reflect.Type) string {
	switch t.String() {
	case "*authentication.UserDetailsAddress", "*model.UserAddress":
		return "*address"
	default:
		return t.String()
	}
}
