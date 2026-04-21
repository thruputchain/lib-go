package liblog

import (
	"context"
	"fmt"
	"math/rand"
	"runtime"

	"github.com/rs/zerolog"
)

type Span struct {
	parent *Span
	label  string
	id     string
	msg    string
}

func (x *Span) Trace() string {
	if x == nil {
		return ""
	}
	if x.parent == nil {
		return fmt.Sprintf("%s", x.id)
	}
	return fmt.Sprintf("%s:%s", x.parent.Trace(), x.id)
}

type spanKey struct{}

func ctxSpan(ctx context.Context) *Span {
	span := ctx.Value(spanKey{})
	if span == nil {
		return nil
	} else {
		return span.(*Span)
	}
}

func Enterf(ctx context.Context, label string, template string, args ...any) context.Context {
	_, file, line, _ := runtime.Caller(1)
	src := fmt.Sprintf("%s:%d ", file, line)
	msg := fmt.Sprintf(template, args...)
	//
	span := &Span{
		parent: ctxSpan(ctx),
		label:  label,
		id:     chooseSpanID(),
		msg:    msg,
	}
	ctx = context.WithValue(ctx, spanKey{}, span)
	zerolog.Ctx(ctx).Debug().Str("trace", span.Trace()).Str("span", span.label+":enter").Msg(src + msg)
	return ctx
}

func Leavef(ctx context.Context, template string, args ...any) {
	_, file, line, _ := runtime.Caller(1)
	src := fmt.Sprintf("%s:%d ", file, line)
	msg := fmt.Sprintf(template, args...)
	//
	span := ctxSpan(ctx)
	zerolog.Ctx(ctx).Debug().Str("trace", span.Trace()).Str("span", span.label+":leave").Msg(src + msg)
}

// generates a random 5-character string from the alphabet [a-zA-Z0-9]
func chooseSpanID() string {
	buf := make([]byte, 5)
	for i := range buf {
		buf[i] = byte(rand.Intn(26) + 'a')
	}
	return string(buf)
}
