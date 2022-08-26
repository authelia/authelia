package templates

import (
	"path"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestShouldNotHaveDeprecatedPlaceholders(t *testing.T) {
	data, err := embedFS.ReadFile(path.Join("src", TemplateCategoryNotifications, TemplateNameEmailEnvelope))
	require.NoError(t, err)

	assert.False(t, tmplEnvelopeHasDeprecatedPlaceholders(data))
}
