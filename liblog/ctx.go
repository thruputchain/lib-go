package liblog

import (
	"context"

	"github.com/rs/zerolog/log"
)

func WithLabeledLogger(ctx context.Context, label string) context.Context {
	withLogger := log.Logger.With().Str("LL", label).Logger()
	return withLogger.WithContext(ctx)
}
