package liblog

import (
	"context"
	"fmt"
	"io"
	"os"
	"runtime"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func init() {
	SetLogLevel(zerolog.ErrorLevel)
}

func SetLogLevelDebug() error {
	return SetLogLevel(zerolog.DebugLevel)
}

func GetLogLevel() zerolog.Level {
	return zerolog.GlobalLevel()
}

func SetLogLevel(level zerolog.Level) error {
	zerolog.SetGlobalLevel(level)
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnixNano
	log.Logger = log.Output(
		zerolog.ConsoleWriter{
			Out:        os.Stderr,
			NoColor:    false,
			TimeFormat: time.RFC3339Nano,
		},
	)

	// Make zerolog.Ctx(ctx) fall back to the global logger.
	zerolog.DefaultContextLogger = &log.Logger

	return nil
}

// global logger

func Writer(ctx context.Context) io.Writer {
	return zerolog.Ctx(ctx)
}

func Debugf(ctx context.Context, template string, args ...any) {
	_, file, line, _ := runtime.Caller(1)
	src := fmt.Sprintf("%s:%d ", file, line)
	msg := fmt.Sprintf(template, args...)
	if span := ctxSpan(ctx); span != nil {
		zerolog.Ctx(ctx).Debug().Str("trace", span.Trace()).Msg(src + msg)
	} else {
		zerolog.Ctx(ctx).Debug().Msg(src + msg)
	}
}

func Infof(ctx context.Context, template string, args ...any) {
	_, file, line, _ := runtime.Caller(1)
	src := fmt.Sprintf("%s:%d ", file, line)
	msg := fmt.Sprintf(template, args...)
	if span := ctxSpan(ctx); span != nil {
		zerolog.Ctx(ctx).Info().Str("trace", span.Trace()).Msg(src + msg)
	} else {
		zerolog.Ctx(ctx).Info().Msg(src + msg)
	}
}

func Fatalf(ctx context.Context, template string, args ...any) {
	_, file, line, _ := runtime.Caller(1)
	src := fmt.Sprintf("%s:%d ", file, line)
	msg := fmt.Sprintf(template, args...)
	if span := ctxSpan(ctx); span != nil {
		zerolog.Ctx(ctx).Fatal().Str("trace", span.Trace()).Msg(src + msg)
	} else {
		zerolog.Ctx(ctx).Fatal().Msg(src + msg)
	}
}

func Errorf(ctx context.Context, template string, args ...any) {
	_, file, line, _ := runtime.Caller(1)
	src := fmt.Sprintf("%s:%d ", file, line)
	msg := fmt.Sprintf(template, args...)
	if span := ctxSpan(ctx); span != nil {
		zerolog.Ctx(ctx).Error().Str("trace", span.Trace()).Msg(src + msg)
	} else {
		zerolog.Ctx(ctx).Error().Msg(src + msg)
	}
}

func Tracef(ctx context.Context, template string, args ...any) {
	_, file, line, _ := runtime.Caller(1)
	src := fmt.Sprintf("%s:%d ", file, line)
	msg := fmt.Sprintf(template, args...)
	if span := ctxSpan(ctx); span != nil {
		zerolog.Ctx(ctx).Trace().Str("trace", span.Trace()).Msg(src + msg)
	} else {
		zerolog.Ctx(ctx).Trace().Msg(src + msg)
	}
}

func Warnf(ctx context.Context, template string, args ...any) {
	_, file, line, _ := runtime.Caller(1)
	src := fmt.Sprintf("%s:%d ", file, line)
	msg := fmt.Sprintf(template, args...)
	if span := ctxSpan(ctx); span != nil {
		zerolog.Ctx(ctx).Warn().Str("trace", span.Trace()).Msg(src + msg)
	} else {
		zerolog.Ctx(ctx).Warn().Msg(src + msg)
	}
}

func Info(ctx context.Context) *zerolog.Event {
	if span := ctxSpan(ctx); span != nil {
		return zerolog.Ctx(ctx).Info().Str("trace", span.Trace())
	} else {
		return zerolog.Ctx(ctx).Info()
	}
}

func Debug(ctx context.Context) *zerolog.Event {
	if span := ctxSpan(ctx); span != nil {
		return zerolog.Ctx(ctx).Debug().Str("trace", span.Trace())
	} else {
		return zerolog.Ctx(ctx).Debug()
	}
}

func Fatal(ctx context.Context) *zerolog.Event {
	if span := ctxSpan(ctx); span != nil {
		return zerolog.Ctx(ctx).Fatal().Str("trace", span.Trace())
	} else {
		return zerolog.Ctx(ctx).Fatal()
	}
}

func Error(ctx context.Context) *zerolog.Event {
	if span := ctxSpan(ctx); span != nil {
		return zerolog.Ctx(ctx).Error().Str("trace", span.Trace())
	} else {
		return zerolog.Ctx(ctx).Error()
	}
}

func Warn(ctx context.Context) *zerolog.Event {
	if span := ctxSpan(ctx); span != nil {
		return zerolog.Ctx(ctx).Warn().Str("trace", span.Trace())
	} else {
		return zerolog.Ctx(ctx).Warn()
	}
}

func Trace(ctx context.Context) *zerolog.Event {
	if span := ctxSpan(ctx); span != nil {
		return zerolog.Ctx(ctx).Trace().Str("trace", span.Trace())
	} else {
		return zerolog.Ctx(ctx).Trace()
	}
}
