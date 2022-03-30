package schema

// DefaultConfiguration represents some defaults users can tune.
type DefaultConfiguration struct {
	UserSecondFactorMethod string `koanf:"user_second_factor_method"`
}
