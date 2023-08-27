package oidc

import (
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"net/url"

	"github.com/ory/fosite"
	"github.com/valyala/fasthttp"
)

// WriteAuthorizeResponse decorates the fosite.WriteAuthorizeResponse so that we can ensure our response mode handler is used first.
func (p *OpenIDConnectProvider) WriteAuthorizeResponse(ctx context.Context, rw http.ResponseWriter, requester fosite.AuthorizeRequester, responder fosite.AuthorizeResponder) {
	if handler := p.ResponseModeHandler(ctx); handler.ResponseModes().Has(requester.GetResponseMode()) {
		handler.WriteAuthorizeResponse(ctx, rw, requester, responder)

		return
	}

	p.OAuth2Provider.WriteAuthorizeResponse(ctx, rw, requester, responder)
}

// ResponseModeHandler is the custom response mode handler for Authelia.
// Implements the fosite.ResponseModeHandler interface.
type ResponseModeHandler struct {
	Configurator
}

// ResponseModes returns the response modes this fosite.ResponseModeHandler is responsible for.
func (h *ResponseModeHandler) ResponseModes() fosite.ResponseModeTypes {
	return fosite.ResponseModeTypes{
		fosite.ResponseModeDefault,
		fosite.ResponseModeQuery,
		fosite.ResponseModeFragment,
		fosite.ResponseModeFormPost,
		ResponseModeJWT,
		ResponseModeQueryJWT,
		ResponseModeFragmentJWT,
		ResponseModeFormPostJWT,
	}
}

// GetPostFormHTMLTemplate returns the 'form_post' response mode template or returns the default.
func (h *ResponseModeHandler) GetPostFormHTMLTemplate(ctx context.Context) (t *template.Template) {
	if t = h.Configurator.GetFormPostHTMLTemplate(ctx); t != nil {
		return t
	}

	return fosite.DefaultFormPostTemplate
}

func (h *ResponseModeHandler) EncodeResponseForm(ctx context.Context, rm fosite.ResponseModeType, config JARMConfigurator, client Client, session any, parameters url.Values) (form url.Values, err error) {
	var issuer string

	if octx, ok := ctx.(Context); ok {
		issuer = octx.RootURL().String()

		parameters.Set(FormParameterIssuer, issuer)
	}

	switch rm {
	case ResponseModeFormPostJWT, ResponseModeQueryJWT, ResponseModeFragment:
		form, err = EncodeJWTSecuredResponseParameters(GenerateJWTSecuredResponse(ctx, config, client, session, parameters))

		if len(issuer) > 0 {
			form.Set(FormParameterIssuer, issuer)
		}

		return
	default:
		return parameters, nil
	}
}

// WriteAuthorizeResponse writes authorization responses.
func (h *ResponseModeHandler) WriteAuthorizeResponse(ctx context.Context, rw http.ResponseWriter, requester fosite.AuthorizeRequester, responder fosite.AuthorizeResponder) {
	wh := rw.Header()
	rh := responder.GetHeader()

	for k := range rh {
		wh.Set(k, rh.Get(k))
	}

	redirectURI := requester.GetRedirectURI()

	var (
		client Client
		ok     bool
	)

	if client, ok = requester.GetClient().(Client); !ok {
		handleWriteAuthorizeErrorJSON(ctx, h.Configurator, rw, fosite.ErrServerError.WithDebug("The client had an unexpected type."))

		return
	}

	var location string

	rm := requester.GetResponseMode()

	if rm == ResponseModeJWT {
		switch requester.GetDefaultResponseMode() {
		case fosite.ResponseModeFragment:
			rm = ResponseModeFragmentJWT
		case fosite.ResponseModeQuery:
			rm = ResponseModeQueryJWT
		}
	}

	switch rm {
	case fosite.ResponseModeFormPost:
		form, err := h.EncodeResponseForm(ctx, rm, h.Configurator, client, requester.GetSession(), responder.GetParameters())
		if err != nil {
			handleWriteAuthorizeErrorJSON(ctx, h.Configurator, rw, fosite.ErrServerError.WithWrap(err).WithDebug(err.Error()))

			return
		}

		rw.Header().Add(fasthttp.HeaderContentType, headerContentTypeTextHTML)
		fosite.WriteAuthorizeFormPostResponse(redirectURI.String(), form, h.GetPostFormHTMLTemplate(ctx), rw)

		return
	case fosite.ResponseModeFragment:
		redirectURI.Fragment = ""

		form, err := h.EncodeResponseForm(ctx, rm, h.Configurator, client, requester.GetSession(), responder.GetParameters())
		if err != nil {
			handleWriteAuthorizeErrorJSON(ctx, h.Configurator, rw, fosite.ErrServerError.WithWrap(err).WithDebug(err.Error()))

			return
		}

		if len(form) > 0 {
			location = redirectURI.String() + "#" + form.Encode()
		} else {
			location = redirectURI.String()
		}
	case fosite.ResponseModeQuery, fosite.ResponseModeDefault:
		query := redirectURI.Query()
		parameters := responder.GetParameters()

		for k := range parameters {
			query.Set(k, parameters.Get(k))
		}

		form, err := h.EncodeResponseForm(ctx, rm, h.Configurator, client, requester.GetSession(), query)
		if err != nil {
			handleWriteAuthorizeErrorJSON(ctx, h.Configurator, rw, fosite.ErrServerError.WithWrap(err).WithDebug(err.Error()))

			return
		}

		redirectURI.RawQuery = form.Encode()

		location = redirectURI.String()
	}

	rw.Header().Set(fasthttp.HeaderLocation, location)
	rw.WriteHeader(http.StatusSeeOther)
}

// WriteAuthorizeError writes authorization errors.
func (h *ResponseModeHandler) WriteAuthorizeError(ctx context.Context, rw http.ResponseWriter, requester fosite.AuthorizeRequester, err error) {
	rfc := fosite.ErrorToRFC6749Error(err).
		WithLegacyFormat(h.GetUseLegacyErrorFormat(ctx)).
		WithExposeDebug(h.GetSendDebugMessagesToClients(ctx)).
		WithLocalizer(h.GetMessageCatalog(ctx), GetLangFromRequester(requester))

	if !requester.IsRedirectURIValid() {
		rw.Header().Set(fasthttp.HeaderContentType, headerContentTypeApplicationJSON)

		var data []byte

		if data, err = json.Marshal(rfc); err != nil {
			if h.GetSendDebugMessagesToClients(ctx) {
				http.Error(rw, fmt.Sprintf(`{"error":"server_error","error_description":"%s"}`, fosite.EscapeJSONString(err.Error())), http.StatusInternalServerError)
			} else {
				http.Error(rw, `{"error":"server_error"}`, http.StatusInternalServerError)
			}

			return
		}

		rw.WriteHeader(rfc.CodeField)
		_, _ = rw.Write(data)

		return
	}

	redirectURI := requester.GetRedirectURI()
	redirectURI.Fragment = ""

	response := rfc.ToValues()

	if octx, ok := ctx.(Context); ok {
		response.Set(FormParameterIssuer, octx.RootURL().String())
	}

	response.Set(FormParameterState, requester.GetState())

	var location string

	switch requester.GetResponseMode() {
	case fosite.ResponseModeFormPost:
		rw.Header().Set(fasthttp.HeaderContentType, headerContentTypeTextHTML)
		fosite.WriteAuthorizeFormPostResponse(redirectURI.String(), response, h.GetFormPostHTMLTemplate(ctx), rw)

		return
	case fosite.ResponseModeFragment:
		location = redirectURI.String() + "#" + response.Encode()
	case fosite.ResponseModeQuery, fosite.ResponseModeDefault:
		for key, values := range redirectURI.Query() {
			for _, value := range values {
				response.Add(key, value)
			}
		}

		redirectURI.RawQuery = response.Encode()

		location = redirectURI.String()
	}

	rw.Header().Set(fasthttp.HeaderLocation, location)
	rw.WriteHeader(http.StatusSeeOther)
}

// ResponseModeHandler returns the response mode handler.
func (p *OpenIDConnectProvider) ResponseModeHandler(ctx context.Context) fosite.ResponseModeHandler {
	if ext := p.Config.GetResponseModeHandlerExtension(ctx); ext != nil {
		return ext
	}

	return handlerDefaultResponseMode
}

var (
	_ fosite.ResponseModeHandler = (*ResponseModeHandler)(nil)

	handlerDefaultResponseMode = &fosite.DefaultResponseModeHandler{}
)
