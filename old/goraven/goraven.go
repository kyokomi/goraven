package goraven

import (
	"context"
	"net/http"

	"github.com/getsentry/raven-go"
	oldcontext "golang.org/x/net/context"

	"github.com/kyokomi/goraven"
)

// Context sentry service context
type Context struct {
	goraven.Context
}

// IsValid return context valid
func (c Context) IsValid() bool {
	return c.Context.IsValid()
}

// NewContext return new context
func NewContext(dsn string) *Context {
	return &Context{
		Context: goraven.NewContext(dsn),
	}
}

// NewContextWithSetDSN return new context with setup dsn to default client
func NewContextWithSetDSN(dsn string) *Context {
	return &Context{
		Context: goraven.NewContextWithSetDSN(dsn),
	}
}

// Handler is old context middleware handler
type Handler func(oldcontext.Context, http.ResponseWriter, *http.Request) error

// Middleware is old context sentry middleware wrapper
type Middleware struct {
	goraven.Middleware
}

// NewOldMiddleware return new middleware (old context)
func NewMiddleware(ctx goraven.Context) *Middleware {
	return &Middleware{
		Middleware: goraven.NewMiddleware(ctx),
	}
}

// SetupHandler setup sentry client (old context)
func (m *Middleware) SetupHandler(h goraven.Handler, sentryCtx goraven.Context, defaultTags map[string]string) Handler {
	return Handler(func(ctx oldcontext.Context, rw http.ResponseWriter, req *http.Request) error {
		return m.Middleware.SetupHandler(h, sentryCtx, defaultTags)(context.Context(ctx), rw, req)
	})
}

// ReportHandler send report when an error is returned when handler is executing (old context)
func (m *Middleware) ReportHandler(h goraven.Handler) Handler {
	return Handler(func(ctx oldcontext.Context, rw http.ResponseWriter, req *http.Request) error {
		return m.Middleware.ReportHandler(h)(context.Context(ctx), rw, req)
	})
}

// RecoverHandler recover panic when executing handler and send report (old context)
func (m *Middleware) RecoverHandler(h goraven.Handler) Handler {
	return Handler(func(ctx oldcontext.Context, rw http.ResponseWriter, req *http.Request) error {
		return m.Middleware.RecoverHandler(h)(context.Context(ctx), rw, req)
	})
}

// GetClient return setup middleware sentry client (old context)
func GetClient(ctx oldcontext.Context) *raven.Client {
	return goraven.GetClient(context.Context(ctx))
}

// DefaultClient return don't setup default sentry client (old context)
func DefaultClient() *raven.Client {
	return goraven.DefaultClient()
}
