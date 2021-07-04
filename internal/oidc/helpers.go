package oidc

func scopeNamesToScopes(scopeSlice []string) (scopes []Scope) {
	for _, name := range scopeSlice {
		if val, ok := scopeDescriptions[name]; ok {
			scopes = append(scopes, Scope{name, val})
		} else {
			scopes = append(scopes, Scope{name, name})
		}
	}

	return scopes
}

func audienceNamesToAudience(scopeSlice []string) (audience []Audience) {
	for _, name := range scopeSlice {
		if val, ok := audienceDescriptions[name]; ok {
			audience = append(audience, Audience{name, val})
		} else {
			audience = append(audience, Audience{name, name})
		}
	}

	return audience
}
