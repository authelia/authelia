package oidc

import (
	"fmt"
	"log"
	"net/http"

	"github.com/authelia/authelia/internal/logging"
	"github.com/ory/fosite"
)

func AuthEndpointGet(oauth2 fosite.OAuth2Provider) http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		// Let's create an AuthorizeRequest object!
		// It will analyze the request and extract important information like scopes, response type and others.
		ar, err := oauth2.NewAuthorizeRequest(r.Context(), r)
		if err != nil {
			logging.Logger().Errorf("Error occurred in NewAuthorizeRequest: %+v", err)
			oauth2.WriteAuthorizeError(rw, ar, err)
			return
		}
		// You have now access to authorizeRequest, Code ResponseTypes, Scopes ...

		var requestedScopes string
		for _, this := range ar.GetRequestedScopes() {
			requestedScopes += fmt.Sprintf(`<li><input type="checkbox" name="scopes" value="%s">%s</li>`, this, this)
		}

		// Normally, this would be the place where you would check if the user is logged in and gives his consent.
		// We're simplifying things and just checking if the request includes a valid username and password
		/* req.ParseForm()
		if req.PostForm.Get("username") != "peter" {
			rw.Header().Set("Content-Type", "text/html; charset=utf-8")
			rw.Write([]byte(`<h1>Login page</h1>`))
			rw.Write([]byte(fmt.Sprintf(`
				<p>Howdy! This is the log in page. For this example, it is enough to supply the username.</p>
				<form method="post">
					<p>
						By logging in, you consent to grant these scopes:
						<ul>%s</ul>
					</p>
					<input type="text" name="username" /> <small>try peter</small><br>
					<input type="submit">
				</form>
			`, requestedScopes)))
			return
		}
		*/

		// let's see what scopes the user gave consent to
		for _, scope := range ar.GetRequestedScopes() {
			ar.GrantScope(scope)
		}

		// Now that the user is authorized, we set up a session:
		session := newSession("peter")

		// When using the HMACSHA strategy you must use something that implements the HMACSessionContainer.
		// It brings you the power of overriding the default values.
		//
		// mySessionData.HMACSession = &strategy.HMACSession{
		//	AccessTokenExpiry: time.Now().Add(time.Day),
		//	AuthorizeCodeExpiry: time.Now().Add(time.Day),
		// }
		//

		// If you're using the JWT strategy, there's currently no distinction between access token and authorize code claims.
		// Therefore, you both access token and authorize code will have the same "exp" claim. If this is something you
		// need let us know on github.
		//
		// mySessionData.JWTClaims.ExpiresAt = time.Now().Add(time.Day)

		// It's also wise to check the requested scopes, e.g.:
		// if ar.GetRequestedScopes().Has("admin") {
		//     http.Error(rw, "you're not allowed to do that", http.StatusForbidden)
		//     return
		// }

		// Now we need to get a response. This is the place where the AuthorizeEndpointHandlers kick in and start processing the request.
		// NewAuthorizeResponse is capable of running multiple response type handlers which in turn enables this library
		// to support open id connect.
		response, err := oauth2.NewAuthorizeResponse(r.Context(), ar, session)

		// Catch any errors, e.g.:
		// * unknown client
		// * invalid redirect
		// * ...
		if err != nil {
			log.Printf("Error occurred in NewAuthorizeResponse: %+v", err)
			oauth2.WriteAuthorizeError(rw, ar, err)
			return
		}

		// Last but not least, send the response!
		oauth2.WriteAuthorizeResponse(rw, ar, response)
	}
}
