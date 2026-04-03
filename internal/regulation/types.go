package regulation

import (
	"context"
	"net"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/authelia/authelia/v4/internal/clock"
	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/storage"
)

// Regulator an authentication regulator preventing attackers to brute force the service.
type Regulator struct {
	users bool
	ips   bool

	config schema.Regulation

	store storage.RegulatorProvider

	clock clock.Provider
}

// Context represents a regulator context.
type Context interface {
	context.Context
	MetricsRecorder

	GetLogger() *logrus.Entry
	RemoteIP() (ip net.IP)
}

// MetricsRecorder represents the methods used to record regulation.
type MetricsRecorder interface {
	RecordAuthn(success, banned bool, authType string)
}

// NewBan constructs a friendly version of ban information for easy formatting.
func NewBan(ban BanType, value string, expires *time.Time) *Ban {
	return &Ban{
		ban:     ban,
		value:   value,
		expires: expires,
	}
}

type Ban struct {
	ban     BanType
	value   string
	expires *time.Time
}

func (b *Ban) IsBanned() bool {
	return b.Type() != BanTypeNone
}

func (b *Ban) Value() string {
	if b == nil {
		return ""
	}

	return b.value
}

func (b *Ban) Type() BanType {
	if b == nil {
		return BanTypeNone
	}

	return b.ban
}

func (b *Ban) Expires() *time.Time {
	if b == nil {
		return nil
	}

	return b.expires
}

func (b *Ban) FormatExpires() string {
	if b == nil || b.expires == nil {
		return FormatExpiresLong(nil)
	}

	return FormatExpiresLong(b.expires)
}

type BanType int

const (
	BanTypeNone BanType = iota
	BanTypeIP
	BanTypeUser
)
