package configuration

var defaults = map[string]any{}

// Defaults returns a copy of the defaults.
func Defaults() map[string]any {
	values := map[string]any{}

	for k, v := range defaults {
		values[k] = v
	}

	return values
}
