package goraven

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/getsentry/raven-go"
	"github.com/pkg/errors"
)

type ctxKey string

var (
	contextKey = ctxKey("sentry_client")
)

// Context sentry service context
type Context struct {
	DSN string
}

// IsValid return context valid
func (c Context) IsValid() bool {
	return len(c.DSN) > 0
}

// NewContext return new context
func NewContext(dsn string) *Context {
	return &Context{
		DSN: dsn,
	}
}

// NewContextWithSetDSN return new context with setup dsn to default client
func NewContextWithSetDSN(dsn string) *Context {
	raven.SetDSN(dsn)
	return NewContext(dsn)
}

// Handler middleware handler
type Handler func(context.Context, http.ResponseWriter, *http.Request) error

// Middleware sentry middleware
type Middleware struct {
	Context Context
}

// NewMiddleware return new middleware
func NewMiddleware(ctx Context) *Middleware {
	return &Middleware{
		Context: ctx,
	}
}

// SetupHandler setup sentry client
func (m *Middleware) SetupHandler(h Handler, sentryCtx Context, defaultTags map[string]string) Handler {
	return func(ctx context.Context, rw http.ResponseWriter, req *http.Request) (err error) {
		if c, err := raven.New(sentryCtx.DSN); err != nil {
			log.Println(err)
		} else {
			// setup report info
			c.SetHttpContext(raven.NewHttp(req))
			c.SetTagsContext(defaultTags)
			ctx = context.WithValue(ctx, contextKey, c)
		}
		return h(ctx, rw, req)
	}
}

// ReportHandler send report when an error is returned when handler is executing
func (m *Middleware) ReportHandler(h Handler) Handler {
	return func(ctx context.Context, rw http.ResponseWriter, req *http.Request) (err error) {
		e := h(ctx, rw, req)
		if e == nil {
			return nil
		}
		log.Println(errors.Cause(e).Error())
		log.Println(GetClient(ctx).CaptureError(e, nil))
		return e
	}
}

// RecoverHandler recover panic when executing handler and send report
func (m *Middleware) RecoverHandler(h Handler) Handler {
	return func(ctx context.Context, rw http.ResponseWriter, req *http.Request) (err error) {
		exception, _ := GetClient(ctx).CapturePanic(func() { err = h(ctx, rw, req) }, nil)
		switch rval := exception.(type) {
		case nil:
		case error:
			err = rval
		default:
			err = fmt.Errorf("%v", rval)
		}
		return err
	}
}

// GetClient return setup middleware sentry client
func GetClient(ctx context.Context) *raven.Client {
	c, ok := ctx.Value(contextKey).(*raven.Client)
	if !ok {
		return raven.DefaultClient
	}
	return c
}

// DefaultClient return don't setup default sentry client
func DefaultClient() *raven.Client {
	return raven.DefaultClient
}
