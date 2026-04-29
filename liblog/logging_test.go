package liblog

import (
	"context"
	"testing"

	"github.com/rs/zerolog"
)

func TestLogger(t *testing.T) {
	SetLogLevel(zerolog.DebugLevel, true)
	Debugf(context.TODO(), "debug message")
}
