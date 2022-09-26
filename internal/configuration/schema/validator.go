package schema

// ErrorContainer represents a container where we can add errors and retrieve them.
type ErrorContainer interface {
	Push(err error)
	PushWarning(err error)
	HasErrors() bool
	HasWarnings() bool
	Errors() []error
	Warnings() []error
}

// StructValidator is a validator for structs.
type StructValidator struct {
	errors   []error
	warnings []error
}

// NewStructValidator is a constructor of struct validator.
func NewStructValidator() *StructValidator {
	val := new(StructValidator)
	val.errors = []error{}
	val.warnings = []error{}

	return val
}

// Push an error to the validator.
func (v *StructValidator) Push(err error) {
	v.errors = append(v.errors, err)
}

// PushWarning error to the validator.
func (v *StructValidator) PushWarning(err error) {
	v.warnings = append(v.warnings, err)
}

// HasErrors checks whether the validator contains errors.
func (v *StructValidator) HasErrors() bool {
	return len(v.errors) > 0
}

// HasWarnings checks whether the validator contains warning errors.
func (v *StructValidator) HasWarnings() bool {
	return len(v.warnings) > 0
}

// Errors returns the errors.
func (v *StructValidator) Errors() []error {
	return v.errors
}

// Warnings returns the warnings.
func (v *StructValidator) Warnings() []error {
	return v.warnings
}

// Clear errors and warnings.
func (v *StructValidator) Clear() {
	v.errors = []error{}
	v.warnings = []error{}
}
