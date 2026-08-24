package handlers

import (
	"fmt"

	"github.com/authelia/authelia/v4/internal/authorization"
)

func handleAuthzGetObjectForwardAuth(ctx AuthzContext) (object authorization.Object, err error) {
	var (
		method          []byte
		requestedObject *authorization.Object
	)

	if method = ctx.XForwardedMethod(); len(method) == 0 {
		return object, fmt.Errorf("header 'X-Forwarded-Method' is empty")
	}

	if requestedObject, err = authorization.NewObjectMethodSchemeHostPath(method, ctx.XForwardedProto(), ctx.XForwardedHost(), ctx.XForwardedURI()); err != nil {
		return object, err
	}

	return *requestedObject, nil
}
