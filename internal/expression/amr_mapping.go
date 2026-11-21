// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package expression

import (
	"fmt"
	"reflect"

	"cel.dev/cel-go/cel"
)

const celVariableClaims = "claims"

// AuthenticationMethodsReferencesMapping is a compiled expression which maps the verified claims of an external
// assertion to a list of RFC8176 authentication method references.
type AuthenticationMethodsReferencesMapping struct {
	program cel.Program
}

// NewAuthenticationMethodsReferencesMapping compiles the expression. An empty expression returns a nil mapping.
func NewAuthenticationMethodsReferencesMapping(expression string) (mapping *AuthenticationMethodsReferencesMapping, err error) {
	if expression == "" {
		return nil, nil
	}

	var (
		env *cel.Env
		ast *cel.Ast
		iss *cel.Issues
	)

	if env, err = cel.NewEnv(cel.Variable(celVariableClaims, cel.MapType(cel.StringType, cel.DynType))); err != nil {
		return nil, fmt.Errorf("error creating the expression environment: %w", err)
	}

	if ast, iss = env.Compile(expression); iss != nil && iss.Err() != nil {
		return nil, fmt.Errorf("error compiling the expression: %w", iss.Err())
	}

	if !ast.OutputType().IsExactType(cel.ListType(cel.StringType)) {
		return nil, fmt.Errorf("expected the output type to be a list of strings but it was '%s'", ast.OutputType())
	}

	var program cel.Program

	if program, err = env.Program(ast); err != nil {
		return nil, fmt.Errorf("error creating the expression program: %w", err)
	}

	return &AuthenticationMethodsReferencesMapping{program: program}, nil
}

// Evaluate evaluates the mapping against the claims. A nil mapping returns nil.
func (m *AuthenticationMethodsReferencesMapping) Evaluate(claims map[string]any) (amr []string, err error) {
	if m == nil {
		return nil, nil
	}

	var out any

	if out, _, err = m.program.Eval(map[string]any{celVariableClaims: claims}); err != nil {
		return nil, fmt.Errorf("error evaluating the expression: %w", err)
	}

	value, ok := out.(interface {
		ConvertToNative(typeDesc reflect.Type) (any, error)
	})
	if !ok {
		return nil, fmt.Errorf("error evaluating the expression: unexpected result type '%T'", out)
	}

	var native any

	if native, err = value.ConvertToNative(reflect.TypeOf([]string{})); err != nil {
		return nil, fmt.Errorf("error converting the expression result: %w", err)
	}

	if amr, ok = native.([]string); !ok {
		return nil, fmt.Errorf("error converting the expression result: unexpected type '%T'", native)
	}

	return amr, nil
}
