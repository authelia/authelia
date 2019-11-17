package schema

import (
	"fmt"
	"reflect"

	"github.com/Workiva/go-datastructures/queue"
)

// ErrorContainer represents a container where we can add errors and retrieve them
type ErrorContainer interface {
	Push(err error)
	HasErrors() bool
	Errors() []error
}

// Validator represents the validator interface
type Validator struct {
	errors map[string][]error
}

// NewValidator create a validator
func NewValidator() *Validator {
	validator := new(Validator)
	validator.errors = make(map[string][]error)
	return validator
}

// QueueItem an item representing a struct field and its path.
type QueueItem struct {
	value reflect.Value
	path  string
}

func (v *Validator) validateOne(item QueueItem, q *queue.Queue) error {
	if item.value.Type().Kind() == reflect.Ptr {
		if item.value.IsNil() {
			return nil
		}

		elem := item.value.Elem()
		q.Put(QueueItem{
			value: elem,
			path:  item.path,
		})
	} else if item.value.Kind() == reflect.Struct {
		numFields := item.value.Type().NumField()

		validateFn := item.value.Addr().MethodByName("Validate")

		if validateFn.IsValid() {
			structValidator := NewStructValidator()
			validateFn.Call([]reflect.Value{reflect.ValueOf(structValidator)})
			v.errors[item.path] = structValidator.Errors()
		}

		for i := 0; i < numFields; i++ {
			field := item.value.Type().Field(i)
			value := item.value.Field(i)

			q.Put(QueueItem{
				value: value,
				path:  item.path + "." + field.Name,
			})
		}
	}
	return nil
}

// Validate validate a struct
func (v *Validator) Validate(s interface{}) error {
	q := queue.New(40)
	q.Put(QueueItem{value: reflect.ValueOf(s), path: "root"})

	for !q.Empty() {
		val, err := q.Get(1)
		if err != nil {
			return err
		}
		item, ok := val[0].(QueueItem)
		if !ok {
			return fmt.Errorf("Cannot convert item into QueueItem")
		}
		v.validateOne(item, q)
	}
	return nil
}

// PrintErrors display the errors thrown during validation
func (v *Validator) PrintErrors() {
	for path, errs := range v.errors {
		fmt.Printf("Errors at %s:\n", path)
		for _, err := range errs {
			fmt.Printf("--> %s\n", err)
		}
	}
}

// Errors return the errors thrown during validation
func (v *Validator) Errors() map[string][]error {
	return v.errors
}

// StructValidator is a validator for structs
type StructValidator struct {
	errors []error
}

// NewStructValidator is a constructor of struct validator
func NewStructValidator() *StructValidator {
	val := new(StructValidator)
	val.errors = make([]error, 0)
	return val
}

// Push an error in the validator.
func (v *StructValidator) Push(err error) {
	v.errors = append(v.errors, err)
}

// HasErrors checks whether the validator contains errors.
func (v *StructValidator) HasErrors() bool {
	return len(v.errors) > 0
}

// Errors returns the errors.
func (v *StructValidator) Errors() []error {
	return v.errors
}
