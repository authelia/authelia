// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package expression

import (
	"errors"
	"fmt"
	"reflect"
	"slices"
	"sort"

	"cel.dev/cel-go/cel"
	"cel.dev/cel-go/common/types"
)

const (
	// ClaimsMappingVariableClaims is the variable holding the claims the upstream provider returned.
	ClaimsMappingVariableClaims = "claims"

	// ClaimsMappingVariableAMR is the variable holding the Authentication Method Reference values resolved before the
	// mapping is applied.
	ClaimsMappingVariableAMR = "amr"

	// ClaimsMappingClaimPreferredUsername is the name of the mappable preferred username claim.
	ClaimsMappingClaimPreferredUsername = "preferred_username"

	// ClaimsMappingClaimName is the name of the mappable name claim.
	ClaimsMappingClaimName = "name"

	// ClaimsMappingClaimEmail is the name of the mappable email claim.
	ClaimsMappingClaimEmail = "email"

	// ClaimsMappingClaimAMR is the name of the mappable Authentication Method Reference claim.
	ClaimsMappingClaimAMR = "amr"

	// ClaimsMappingClaimDisabled is the name of the mapping which refuses the sign in when it evaluates to true.
	ClaimsMappingClaimDisabled = "disabled"
)

var claimsMappingStringClaims = []string{ClaimsMappingClaimEmail, ClaimsMappingClaimName, ClaimsMappingClaimPreferredUsername}

// ClaimsMapping is a compiled set of expressions which remap the claims of an upstream provider to result claims.
type ClaimsMapping struct {
	strings map[string]cel.Program
	amr     cel.Program

	disabled cel.Program
}

// NewClaimsMapping compiles and type checks the expressions of a mapping of result claim name to expression.
func NewClaimsMapping(mapping map[string]string) (claimsMapping *ClaimsMapping, err error) {
	var env *cel.Env

	if env, err = cel.NewEnv(withBaseCELEnvOpts(
		cel.Variable(ClaimsMappingVariableClaims, cel.MapType(cel.StringType, cel.DynType)),
		cel.Variable(ClaimsMappingVariableAMR, cel.ListType(cel.StringType)),
	)...); err != nil {
		return nil, fmt.Errorf("error creating the claims mapping environment: %w", err)
	}

	claimsMapping = &ClaimsMapping{strings: make(map[string]cel.Program, len(mapping))}

	names := make([]string, 0, len(mapping))

	for name := range mapping {
		names = append(names, name)
	}

	sort.Strings(names)

	for _, name := range names {
		var program cel.Program

		switch {
		case name == ClaimsMappingClaimAMR:
			program, err = newClaimsMappingProgram(env, mapping[name], cel.ListType(cel.StringType))
			claimsMapping.amr = program
		case name == ClaimsMappingClaimDisabled:
			program, err = newClaimsMappingProgram(env, mapping[name], cel.BoolType)
			claimsMapping.disabled = program
		case slices.Contains(claimsMappingStringClaims, name):
			program, err = newClaimsMappingProgram(env, mapping[name], cel.StringType)
			claimsMapping.strings[name] = program
		default:
			return nil, fmt.Errorf("the claim '%s' can't be mapped: it must be one of 'amr', 'disabled', 'email', 'name', or 'preferred_username'", name)
		}

		if err != nil {
			return nil, fmt.Errorf("the claim '%s' has an invalid expression: %w", name, err)
		}
	}

	return claimsMapping, nil
}

func newClaimsMappingProgram(env *cel.Env, expression string, expected *cel.Type) (program cel.Program, err error) {
	ast, issues := env.Compile(expression)
	if issues != nil && issues.Err() != nil {
		return nil, issues.Err()
	}

	output := ast.OutputType()

	if output.Kind() != types.DynKind && !expected.IsAssignableType(output) {
		return nil, fmt.Errorf("the result type must be '%s' but it is '%s'", expected.String(), output.String())
	}

	return env.Program(ast)
}

// MapsAMR returns true when the mapping includes the Authentication Method Reference claim.
func (m *ClaimsMapping) MapsAMR() bool {
	return m.amr != nil
}

// Disabled evaluates whether the sign in must be refused. A failure to evaluate the expression returns true with the
// error so that a caller can't accidentally allow the sign in.
func (m *ClaimsMapping) Disabled(claims map[string]any, amr []string) (disabled bool, err error) {
	if m.disabled == nil {
		return false, nil
	}

	if claims == nil {
		claims = map[string]any{}
	}

	if amr == nil {
		amr = []string{}
	}

	result, _, err := m.disabled.Eval(map[string]any{ClaimsMappingVariableClaims: claims, ClaimsMappingVariableAMR: amr})
	if err != nil {
		return true, fmt.Errorf("error evaluating the expression for the claim '%s': %w", ClaimsMappingClaimDisabled, err)
	}

	value, ok := result.Value().(bool)
	if !ok {
		return true, fmt.Errorf("error evaluating the expression for the claim '%s': the result is not a boolean", ClaimsMappingClaimDisabled)
	}

	return value, nil
}

// Evaluate evaluates the expressions against the upstream claims and the resolved Authentication Method Reference
// values. Only configured claims are present in the returned values, and amr is nil unless MapsAMR returns true. Any
// failure is returned as an error as the unmapped claims must not be used as a fallback.
func (m *ClaimsMapping) Evaluate(claims map[string]any, amr []string) (values map[string]string, mappedAMR []string, err error) {
	if claims == nil {
		claims = map[string]any{}
	}

	if amr == nil {
		amr = []string{}
	}

	vars := map[string]any{ClaimsMappingVariableClaims: claims, ClaimsMappingVariableAMR: amr}

	values = make(map[string]string, len(m.strings))

	for name, program := range m.strings {
		if values[name], err = evaluateClaimsMappingString(program, vars); err != nil {
			return nil, nil, fmt.Errorf("error evaluating the expression for the claim '%s': %w", name, err)
		}
	}

	if m.amr != nil {
		if mappedAMR, err = evaluateClaimsMappingAMR(m.amr, vars); err != nil {
			return nil, nil, fmt.Errorf("error evaluating the expression for the claim '%s': %w", ClaimsMappingClaimAMR, err)
		}
	}

	return values, mappedAMR, nil
}

func evaluateClaimsMappingString(program cel.Program, vars map[string]any) (value string, err error) {
	result, _, err := program.Eval(vars)
	if err != nil {
		return "", err
	}

	var ok bool

	if value, ok = result.Value().(string); !ok {
		return "", errors.New("the result is not a string")
	}

	return value, nil
}

func evaluateClaimsMappingAMR(program cel.Program, vars map[string]any) (values []string, err error) {
	result, _, err := program.Eval(vars)
	if err != nil {
		return nil, err
	}

	native, err := result.ConvertToNative(reflect.TypeFor[[]string]())
	if err != nil {
		return nil, errors.New("the result is not a list of strings")
	}

	values, ok := native.([]string)
	if !ok {
		return nil, errors.New("the result is not a list of strings")
	}

	if slices.Contains(values, "") {
		return nil, errors.New("the result contains an empty value")
	}

	return values, nil
}
