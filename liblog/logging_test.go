package liblog

import (
	"context"
	"testing"

	"github.com/rs/zerolog"
)

func TestLogger(t *testing.T) {
	SetLogLevel(zerolog.DebugLevel)
	Debugf(context.TODO(), "debug message")
}
