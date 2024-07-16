package source_code

import (
	"context"
	"errors"
	"time"
)

var PingURL = "https://www.google.com/"
var HasChallengeErr = errors.New("has challenge")

type SourceOptions struct {
	MaxDuration time.Duration
	BackOff     *BackOff
}

type SourceGetter interface {
	Get(ctx context.Context, endpoint string, options ...SourceOptions) ([]byte, int, error)
	Ping(ctx context.Context) bool
	Enabled() bool
	Name() string
}
