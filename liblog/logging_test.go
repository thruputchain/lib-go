package liblog

import (
	"testing"

	"github.com/rs/zerolog"
)

func TestLogger(t *testing.T) {
	SetLogLevel(zerolog.DebugLevel)
	Debugf("debug message")
}
