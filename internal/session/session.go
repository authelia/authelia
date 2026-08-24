package session

import (
	"github.com/authelia/authelia/v4/internal/configuration/schema"
)

// NewEncapsulatedSession returns a new encapsulated session which binds a Strategy to the Context of a single request.
func NewEncapsulatedSession(base Strategy, ctx Context) *EncapsulatedSession {
	return &EncapsulatedSession{base: base, ctx: ctx}
}

// EncapsulatedSession encapsulates a Strategy and the Context of the request it's handling, which allows consumers to
// use the session without having to thread the Context through every call.
type EncapsulatedSession struct {
	base Strategy

	ctx Context
}

// NewDefaultUserSession returns a new default UserSession for this session provider.
func (p *EncapsulatedSession) NewDefaultUserSession() (userSession UserSession) {
	return p.base.NewDefault()
}

// GetSession return the user session from a request.
func (p *EncapsulatedSession) GetSession() (userSession UserSession, err error) {
	var session *UserSession

	if session, err = p.base.Get(p.ctx); err != nil {
		return p.base.NewDefault(), err
	}

	return *session, nil
}

// SaveSession save the user session.
func (p *EncapsulatedSession) SaveSession(userSession *UserSession) (err error) {
	return p.base.Save(p.ctx, userSession)
}

// RegenerateSession regenerate a session ID.
func (p *EncapsulatedSession) RegenerateSession() (err error) {
	return p.base.Regenerate(p.ctx)
}

// DestroySession destroy a session ID and delete the cookie.
func (p *EncapsulatedSession) DestroySession() (err error) {
	return p.base.Destroy(p.ctx)
}

// GetSessionConfig returns the session configuration.
func (p *EncapsulatedSession) GetSessionConfig() (config schema.SessionCookie) {
	return p.base.GetConfig()
}

// Manager is the interface that wraps the basic methods of a session provider bound to a specific request.
type Manager interface {
	NewDefaultUserSession() (userSession UserSession)
	GetSession() (userSession UserSession, err error)
	SaveSession(userSession *UserSession) (err error)
	RegenerateSession() (err error)
	DestroySession() (err error)
	GetSessionConfig() (config schema.SessionCookie)
}

var (
	_ Manager = (*EncapsulatedSession)(nil)
)
