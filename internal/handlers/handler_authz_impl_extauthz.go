package handlers

import (
	"github.com/authelia/authelia/v4/internal/authorization"
)

func handleAuthzGetObjectExtAuthz(ctx AuthzContext) (object authorization.Object, err error) {
	var requestedObject *authorization.Object

	host := ctx.Host()

	if requestedObject, err = authorization.NewObjectMethodSchemeHostPath(ctx.Method(), ctx.XForwardedProto(), host, ctx.AuthzPath()); err != nil {
		return object, err
	}

	return *requestedObject, nil
}
