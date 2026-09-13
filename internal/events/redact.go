// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package events

import (
	"reflect"
)

// Redact returns a copy of the payload with every field tagged sensitive set to its zero value. Sensitive fields
// carry the omitempty JSON option, so a zeroed field is absent from the marshaled payload rather than present and
// empty. The input is never mutated.
//
// The copy is deep for every struct, pointer, slice, array, and map reachable from the payload: each is rebuilt (or,
// for pointers, reallocated) before being descended into and redacted, so mutating a nested field of the redacted
// copy, or of the caller's original, never affects the other. The one exception is data held behind an interface
// value (for example NotificationValues.Details, a map[string]any): the interface value itself is copied, but if
// its dynamic value is itself a pointer, slice, map, or channel, that referenced data is shared with the original,
// and a sensitive field nested inside it would not be redacted. No payload type currently stores sensitive data that
// way; the exhaustive test in redact_test.go covers only what is statically reachable through struct field tags. Map
// keys are likewise copied verbatim and never descended into, so a sensitive value used as a map key, rather than as a
// map value, would not be redacted either.
//
// It fails closed. A payload which is not a non-nil pointer to a struct, or whose rebuilt copy is somehow not a Data,
// cannot be redacted and returns nil rather than the original: the caller must treat that as a delivery failure, since
// the one thing a redaction primitive must never do when it fails is hand back the unredacted payload. Marshal
// refuses to encode a nil payload for that reason.
func Redact(data Data) (redacted Data) {
	v := reflect.ValueOf(data)

	if v.Kind() != reflect.Pointer || v.IsNil() || v.Elem().Kind() != reflect.Struct {
		return nil
	}

	out := reflect.New(v.Elem().Type())

	out.Elem().Set(v.Elem())

	redactValue(out.Elem())

	value, ok := out.Interface().(Data)
	if !ok {
		return nil
	}

	return value
}

func redactValue(v reflect.Value) {
	t := v.Type()

	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)

		if !f.IsExported() {
			continue
		}

		field := v.Field(i)

		if f.Tag.Get(TagSensitive) == "true" {
			field.Set(reflect.Zero(f.Type))

			continue
		}

		redactContainer(field)
	}
}

func redactContainer(v reflect.Value) {
	switch v.Kind() { //nolint:exhaustive // Only container kinds that can carry a nested struct need to descend; scalar and interface kinds fall through.
	case reflect.Struct:
		redactValue(v)
	case reflect.Pointer:
		if v.IsNil() {
			return
		}

		cp := reflect.New(v.Type().Elem())

		cp.Elem().Set(v.Elem())

		redactContainer(cp.Elem())

		v.Set(cp)
	case reflect.Slice:
		if v.IsNil() {
			return
		}

		cp := reflect.MakeSlice(v.Type(), v.Len(), v.Len())

		for i := 0; i < v.Len(); i++ {
			cp.Index(i).Set(v.Index(i))

			redactContainer(cp.Index(i))
		}

		v.Set(cp)
	case reflect.Array:
		for i := 0; i < v.Len(); i++ {
			redactContainer(v.Index(i))
		}
	case reflect.Map:
		if v.IsNil() {
			return
		}

		cp := reflect.MakeMapWithSize(v.Type(), v.Len())

		iter := v.MapRange()

		for iter.Next() {
			val := reflect.New(v.Type().Elem()).Elem()

			val.Set(iter.Value())

			redactContainer(val)

			cp.SetMapIndex(iter.Key(), val)
		}

		v.Set(cp)
	}
}
