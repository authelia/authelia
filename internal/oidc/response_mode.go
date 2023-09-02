package oidc

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/ory/fosite"
	"github.com/valyala/fasthttp"
)

// ResponseModeHandler returns the response mode handler.
func (p *OpenIDConnectProvider) ResponseModeHandler(ctx context.Context) fosite.ResponseModeHandler {
	if ext := p.Config.GetResponseModeHandlerExtension(ctx); ext != nil {
		return ext
	}

	return handlerDefaultResponseMode
}

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
	Config Configurator
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

// EncodeResponseForm encodes the response form if necessary.
func (h *ResponseModeHandler) EncodeResponseForm(ctx context.Context, rm fosite.ResponseModeType, client Client, session any, parameters url.Values) (form url.Values, err error) {
	switch rm {
	case ResponseModeFormPostJWT, ResponseModeQueryJWT, ResponseModeFragmentJWT:
		return EncodeJWTSecuredResponseParameters(GenerateJWTSecuredResponse(ctx, h.Config, client, session, parameters))
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

	h.doWriteAuthorizeResponse(ctx, rw, requester, responder.GetParameters())
}

// WriteAuthorizeError writes authorization errors.
func (h *ResponseModeHandler) WriteAuthorizeError(ctx context.Context, rw http.ResponseWriter, requester fosite.AuthorizeRequester, e error) {
	rfc := fosite.ErrorToRFC6749Error(e).
		WithLegacyFormat(h.Config.GetUseLegacyErrorFormat(ctx)).
		WithExposeDebug(h.Config.GetSendDebugMessagesToClients(ctx)).
		WithLocalizer(h.Config.GetMessageCatalog(ctx), GetLangFromRequester(requester))

	if !requester.IsRedirectURIValid() {
		h.doWriteAuthorizeErrorJSON(ctx, rw, rfc)

		return
	}

	parameters := rfc.ToValues()

	if state := requester.GetState(); len(state) != 0 {
		parameters.Set(FormParameterState, state)
	}

	switch requester.GetResponseMode() {
	case fosite.ResponseModeFormPost, fosite.ResponseModeQuery, fosite.ResponseModeFragment, fosite.ResponseModeDefault:
		if issuer := h.Config.GetAuthorizationServerIdentificationIssuer(ctx); len(issuer) != 0 {
			parameters.Set(FormParameterIssuer, issuer)
		}
	}

	h.doWriteAuthorizeResponse(ctx, rw, requester, parameters)
}

func (h *ResponseModeHandler) doWriteAuthorizeResponse(ctx context.Context, rw http.ResponseWriter, requester fosite.AuthorizeRequester, parameters url.Values) {
	redirectURI := requester.GetRedirectURI()
	redirectURI.Fragment = ""

	var (
		client Client
		ok     bool
	)

	if client, ok = requester.GetClient().(Client); !ok {
		h.doWriteAuthorizeErrorJSON(ctx, rw, fosite.ErrServerError.WithDebug("The client had an unexpected type."))

		return
	}

	rm := requester.GetResponseMode()

	if rm == ResponseModeJWT {
		if requester.GetResponseTypes().ExactOne(ResponseTypeAuthorizationCodeFlow) {
			rm = ResponseModeQueryJWT
		} else {
			rm = ResponseModeFragmentJWT
		}
	}

	var (
		form     url.Values
		err      error
		location string
	)

	switch rm {
	case fosite.ResponseModeFormPost, ResponseModeFormPostJWT:
		if form, err = h.EncodeResponseForm(ctx, rm, client, requester.GetSession(), parameters); err != nil {
			h.doWriteAuthorizeErrorJSON(ctx, rw, fosite.ErrServerError.WithWrap(err).WithDebug(err.Error()))

			return
		}

		rw.Header().Set(fasthttp.HeaderContentType, headerContentTypeTextHTML)
		fosite.WriteAuthorizeFormPostResponse(redirectURI.String(), form, h.Config.GetFormPostHTMLTemplate(ctx), rw)

		return
	case fosite.ResponseModeQuery, fosite.ResponseModeDefault, ResponseModeQueryJWT, ResponseModeJWT:
		for key, values := range redirectURI.Query() {
			for _, value := range values {
				parameters.Add(key, value)
			}
		}

		if form, err = h.EncodeResponseForm(ctx, rm, client, requester.GetSession(), parameters); err != nil {
			h.doWriteAuthorizeErrorJSON(ctx, rw, fosite.ErrServerError.WithWrap(err).WithDebug(err.Error()))

			return
		}

		redirectURI.RawQuery = form.Encode()

		location = redirectURI.String()
	case fosite.ResponseModeFragment, ResponseModeFragmentJWT:
		if form, err = h.EncodeResponseForm(ctx, rm, client, requester.GetSession(), parameters); err != nil {
			h.doWriteAuthorizeErrorJSON(ctx, rw, fosite.ErrServerError.WithWrap(err).WithDebug(err.Error()))

			return
		}

		location = redirectURI.String() + "#" + form.Encode()
	}

	rw.Header().Set(fasthttp.HeaderLocation, location)
	rw.WriteHeader(http.StatusSeeOther)
}

func (h *ResponseModeHandler) doWriteAuthorizeErrorJSON(ctx context.Context, rw http.ResponseWriter, rfc *fosite.RFC6749Error) {
	rw.Header().Set(fasthttp.HeaderContentType, headerContentTypeApplicationJSON)

	var (
		data []byte
		err  error
	)

	if data, err = json.Marshal(rfc); err != nil {
		if h.Config.GetSendDebugMessagesToClients(ctx) {
			http.Error(rw, fmt.Sprintf(`{"error":"server_error","error_description":"%s"}`, fosite.EscapeJSONString(err.Error())), http.StatusInternalServerError)
		} else {
			http.Error(rw, `{"error":"server_error"}`, http.StatusInternalServerError)
		}

		return
	}

	rw.WriteHeader(rfc.CodeField)
	_, _ = rw.Write(data)
}

var (
	_ fosite.ResponseModeHandler = (*ResponseModeHandler)(nil)

	handlerDefaultResponseMode = &fosite.DefaultResponseModeHandler{}
)
