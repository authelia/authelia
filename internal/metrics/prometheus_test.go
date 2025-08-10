package metrics

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewPrometheus(t *testing.T) {
	p := NewPrometheus()

	assert.NotNil(t, p)

	p.RecordRequest("400", "GET", time.Second)
	p.RecordAuthz("400")
	p.RecordAuthn(true, false, "WebAuthn")
	p.RecordAuthn(true, false, "1fa")
	p.RecordAuthn(true, false, "passkey")
	p.RecordAuthenticationDuration(true, time.Second)
	p.RecordRequestOpenIDConnect("example", "200", time.Second)
}
