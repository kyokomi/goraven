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
	Context     Context
	defaultTags map[string]string
}

// NewMiddleware return new middleware
func NewMiddleware(ctx Context) *Middleware {
	return &Middleware{
		Context: ctx,
	}
}

// SetDefaultTags set default tags
func (m *Middleware) SetDefaultTags(tags map[string]string) {
	m.defaultTags = tags
}

// SetupHandler setup sentry client
func (m *Middleware) SetupHandler(h Handler) Handler {
	return func(ctx context.Context, rw http.ResponseWriter, req *http.Request) (err error) {
		if c, err := raven.New(m.Context.DSN); err != nil {
			log.Println(err)
		} else {
			// setup report info
			c.SetHttpContext(raven.NewHttp(req))
			c.SetTagsContext(m.defaultTags)
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

// Client is raven client wrapper
type Client struct {
	*raven.Client
}

// GetClient return setup middleware sentry client
func GetClient(ctx context.Context) Client {
	c, ok := ctx.Value(contextKey).(*raven.Client)
	if !ok {
		return DefaultClient()
	}
	return Client{Client: c}
}

// DefaultClient return don't setup default sentry client
func DefaultClient() Client {
	return Client{Client: raven.DefaultClient}
}

// CaptureErrorMessage formats and delivers an error to the Sentry server.
// Adds a stacktrace to the packet, excluding the call to this method.
// messageの部分を自分で指定したかったのでravenにあったCaptureErrorをコピペして改造しました
func (c *Client) CaptureErrorMessage(message string, err error, tags map[string]string, interfaces ...raven.Interface) string {
	packet := raven.NewPacket(message, append(interfaces, raven.NewException(err,
		raven.NewStacktrace(1, 3, c.IncludePaths())))...)
	eventID, _ := c.Capture(packet, tags)
	return eventID
}
