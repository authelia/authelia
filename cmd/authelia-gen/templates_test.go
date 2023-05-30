package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestShouldFailToLoadBadTemplate(t *testing.T) {
	assert.Panics(t, func() {
		mustLoadTmplFS("bad tmpl")
	})
}
