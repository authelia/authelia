package handlers

import (
	"context"
	"time"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/middlewares"
	"github.com/sirupsen/logrus"
	"github.com/valyala/fasthttp"
)

type contextAdapter struct {
	*middlewares.AutheliaCtx
	stdCtx context.Context
}

func (c *contextAdapter) GetConfiguration() *schema.Configuration {
	config := c.AutheliaCtx.GetConfiguration()
	return &config
}

func (c *contextAdapter) GetLogger() *logrus.Entry {
	return c.AutheliaCtx.GetLogger()
}

func (c *contextAdapter) GetProviders() middlewares.Providers {
	return c.AutheliaCtx.GetProviders()
}

func (c *contextAdapter) Deadline() (deadline time.Time, ok bool) {
	return c.stdCtx.Deadline()
}

func (c *contextAdapter) Done() <-chan struct{} {
	return c.stdCtx.Done()
}

func (c *contextAdapter) Err() error {
	return c.stdCtx.Err()
}

func (c *contextAdapter) Value(key interface{}) interface{} {
	return c.stdCtx.Value(key)
}

// ReadyGET is used to check the readiness of authelia.
func ReadyGET(ctx *middlewares.AutheliaCtx) {
	healthCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	adapter := &contextAdapter{
		AutheliaCtx: ctx,
		stdCtx:      healthCtx,
	}

	err := ctx.Providers.HealthChecks(adapter, true)
	if err != nil {
		ctx.Logger.WithError(err).Error("Startup check failed")
		ctx.ReplyStatusCode(fasthttp.StatusServiceUnavailable)
		return
	}

	ctx.ReplyOK()
}
