package schema_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/clems4ever/authelia/configuration/schema"
)

type TestNestedStruct struct {
	MustBe5 int
}

func (tns *TestNestedStruct) Validate(validator *schema.StructValidator) {
	if tns.MustBe5 != 5 {
		validator.Push(fmt.Errorf("MustBe5 must be 5"))
	}
}

type TestStruct struct {
	MustBe10   int
	NotEmpty   string
	SetDefault string
	Nested     TestNestedStruct
	Nested2    TestNestedStruct
	NilPtr     *int
	NestedPtr  *TestNestedStruct
}

func (ts *TestStruct) Validate(validator *schema.StructValidator) {
	if ts.MustBe10 != 10 {
		validator.Push(fmt.Errorf("MustBe10 must be 10"))
	}

	if ts.NotEmpty == "" {
		validator.Push(fmt.Errorf("NotEmpty must not be empty"))
	}

	if ts.SetDefault == "" {
		ts.SetDefault = "xyz"
	}
}

func TestValidator(t *testing.T) {
	validator := schema.NewValidator()

	s := TestStruct{
		MustBe10:  5,
		NotEmpty:  "",
		NestedPtr: &TestNestedStruct{},
	}

	err := validator.Validate(&s)
	if err != nil {
		panic(err)
	}

	errs := validator.Errors()
	assert.Equal(t, 4, len(errs))

	assert.Equal(t, 2, len(errs["root"]))
	assert.ElementsMatch(t, []error{
		fmt.Errorf("MustBe10 must be 10"),
		fmt.Errorf("NotEmpty must not be empty")}, errs["root"])

	assert.Equal(t, 1, len(errs["root.Nested"]))
	assert.ElementsMatch(t, []error{
		fmt.Errorf("MustBe5 must be 5")}, errs["root.Nested"])

	assert.Equal(t, 1, len(errs["root.Nested2"]))
	assert.ElementsMatch(t, []error{
		fmt.Errorf("MustBe5 must be 5")}, errs["root.Nested2"])

	assert.Equal(t, 1, len(errs["root.NestedPtr"]))
	assert.ElementsMatch(t, []error{
		fmt.Errorf("MustBe5 must be 5")}, errs["root.NestedPtr"])

	assert.Equal(t, "xyz", s.SetDefault)
}
