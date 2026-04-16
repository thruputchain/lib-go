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
	return nil
}

// global logger

func Writer(ctx context.Context) io.Writer {
	return log.Ctx(ctx)
}

func Debugf(ctx context.Context, template string, args ...any) {
	_, file, line, _ := runtime.Caller(1)
	src := fmt.Sprintf("%s:%d ", file, line)
	msg := fmt.Sprintf(template, args...)
	log.Ctx(ctx).Debug().Msg(src + msg)
}

func Infof(ctx context.Context, template string, args ...any) {
	_, file, line, _ := runtime.Caller(1)
	src := fmt.Sprintf("%s:%d ", file, line)
	msg := fmt.Sprintf(template, args...)
	log.Ctx(ctx).Info().Msg(src + msg)
}

func Fatalf(ctx context.Context, template string, args ...any) {
	_, file, line, _ := runtime.Caller(1)
	src := fmt.Sprintf("%s:%d ", file, line)
	msg := fmt.Sprintf(template, args...)
	log.Ctx(ctx).Fatal().Msg(src + msg)
}

func Errorf(ctx context.Context, template string, args ...any) {
	_, file, line, _ := runtime.Caller(1)
	src := fmt.Sprintf("%s:%d ", file, line)
	msg := fmt.Sprintf(template, args...)
	log.Ctx(ctx).Error().Msg(src + msg)
}

func Tracef(ctx context.Context, template string, args ...any) {
	_, file, line, _ := runtime.Caller(1)
	src := fmt.Sprintf("%s:%d ", file, line)
	msg := fmt.Sprintf(template, args...)
	log.Ctx(ctx).Trace().Msg(src + msg)
}

func Warnf(ctx context.Context, template string, args ...any) {
	_, file, line, _ := runtime.Caller(1)
	src := fmt.Sprintf("%s:%d ", file, line)
	msg := fmt.Sprintf(template, args...)
	log.Ctx(ctx).Warn().Msg(src + msg)
}

func Info(ctx context.Context) *zerolog.Event {
	return log.Ctx(ctx).Info()
}

func Debug(ctx context.Context) *zerolog.Event {
	return log.Ctx(ctx).Debug()
}

func Fatal(ctx context.Context) *zerolog.Event {
	return log.Ctx(ctx).Fatal()
}

func Error(ctx context.Context) *zerolog.Event {
	return log.Ctx(ctx).Error()
}

func Warn(ctx context.Context) *zerolog.Event {
	return log.Ctx(ctx).Warn()
}

func Trace(ctx context.Context) *zerolog.Event {
	return log.Ctx(ctx).Trace()
}
