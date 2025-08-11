package logging

import (
	"context"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestPrintfLoggers(t *testing.T) {
	x := LoggerPrintf(logrus.DebugLevel)
	assert.NotNil(t, x)

	x.Printf("abc %s", "123")

	y := LoggerCtxPrintf(logrus.DebugLevel)
	assert.NotNil(t, y)

	y.Printf(context.Background(), "abc %s", "123")
}
