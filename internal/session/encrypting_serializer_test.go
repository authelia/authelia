package session

import (
	"testing"

	"github.com/fasthttp/session/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestShouldEncryptAndDecrypt(t *testing.T) {
	payload := session.Dict{KV: map[string]any{"key": "value"}}

	dst, err := payload.MarshalMsg(nil)
	require.NoError(t, err)

	serializer := NewEncryptingSerializer("asecret")
	encryptedDst, err := serializer.Encode(payload)
	require.NoError(t, err)

	assert.NotEqual(t, dst, encryptedDst)

	decodedPayload := session.Dict{}
	err = serializer.Decode(&decodedPayload, encryptedDst)
	require.NoError(t, err)

	assert.Equal(t, "value", decodedPayload.KV["key"])
}

func TestShouldNotSupportUnencryptedSessionForBackwardCompatibility(t *testing.T) {
	payload := session.Dict{KV: map[string]any{"key": "value"}}

	dst, err := payload.MarshalMsg(nil)
	require.NoError(t, err)

	serializer := NewEncryptingSerializer("asecret")

	decodedPayload := session.Dict{}
	err = serializer.Decode(&decodedPayload, dst)
	assert.EqualError(t, err, "unable to decrypt session: cipher: message authentication failed")
}
