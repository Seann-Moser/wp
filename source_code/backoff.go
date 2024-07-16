package source_code

import (
	"context"
	"fmt"
	"github.com/Seann-Moser/cutil/logc"
	bf "github.com/cenkalti/backoff/v4"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"time"
)

func GetFlagWithPrefix(prefix, flag string) string {
	if prefix == "" {
		return flag
	}
	return fmt.Sprintf("%s-%s", prefix, flag)
}

type BackOff struct {
	maxRetry        uint64
	maxInterval     time.Duration
	maxElapsedTime  time.Duration
	initialInterval time.Duration
}

func BackOffFlags(prefix string) *pflag.FlagSet {
	fs := pflag.NewFlagSet(GetFlagWithPrefix("backoff", prefix), pflag.ExitOnError)
	fs.Uint64(GetFlagWithPrefix("max-retry", prefix), 5, "")
	fs.Duration(GetFlagWithPrefix("max-interval", prefix), 15*time.Second, "")
	fs.Duration(GetFlagWithPrefix("max-elapsed-time", prefix), 45*time.Second, "")
	fs.Duration(GetFlagWithPrefix("max-initial-interval", prefix), 100*time.Millisecond, "")
	return fs
}
func NewBackoffWithFlags(prefix string) *BackOff {
	return &BackOff{
		maxRetry:        viper.GetUint64(GetFlagWithPrefix("max-retry", prefix)),
		maxInterval:     viper.GetDuration(GetFlagWithPrefix("max-interval", prefix)),
		maxElapsedTime:  viper.GetDuration(GetFlagWithPrefix("max-elapsed-time", prefix)),
		initialInterval: viper.GetDuration(GetFlagWithPrefix("max-initial-interval", prefix)),
	}
}
func (b *BackOff) Retry(ctx context.Context, operation bf.Operation) error {
	notify := func(err error, backoffDuration time.Duration) {
		logc.Warn(ctx, "retrying", zap.Error(err), zap.Duration("backoff_duration", backoffDuration))
	}
	if err := bf.RetryNotify(operation, b.getBackoff(), notify); err != nil {
		return err
	}
	return nil

}
func (b *BackOff) getBackoff() bf.BackOff {
	requestExpBackOff := bf.NewExponentialBackOff()
	requestExpBackOff.InitialInterval = b.initialInterval
	requestExpBackOff.RandomizationFactor = 0.5
	requestExpBackOff.Multiplier = 1.5
	requestExpBackOff.MaxInterval = b.maxInterval
	requestExpBackOff.MaxElapsedTime = b.maxElapsedTime
	return bf.WithMaxRetries(requestExpBackOff, b.maxRetry)
}
