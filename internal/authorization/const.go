package authorization

import "github.com/authelia/authelia/internal/configuration/schema"

// Level is the type representing an authorization level.
type Level int

const (
	// Bypass bypass level.
	Bypass Level = iota
	// OneFactor one factor level.
	OneFactor Level = iota
	// TwoFactor two factor level.
	TwoFactor Level = iota
	// Denied denied level.
	Denied Level = iota
)

var testACLNetwork = []schema.ACLNetwork{
	{
		Name:     []string{"localhost"},
		Networks: []string{"127.0.0.1"},
	},
	{
		Name:     []string{"internal"},
		Networks: []string{"10.0.0.0/8"},
	},
}
