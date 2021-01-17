package authorization

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
