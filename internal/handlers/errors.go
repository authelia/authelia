package handlers

import "errors"

var errPasswordPolicyNoMet = errors.New("the supplied password does not met the security policy")
