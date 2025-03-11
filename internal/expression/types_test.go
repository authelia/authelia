package expression

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestTypeMethods(t *testing.T) {
	detailer := &UserAttributeResolverDetailer{
		updated: time.Unix(1000000000, 0).UTC(),
	}

	assert.Equal(t, time.Unix(1000000000, 0).UTC(), detailer.GetUpdatedAt())
}
