package alerts

import (
	"gopkg.in/retry.v1"
	"time"
)

func getRetryStrategy() retry.Exponential {
	return retry.Exponential{
		Initial:  100 * time.Millisecond,
		Factor:   2,
		MaxDelay: 1 * time.Second,
		Jitter:   true,
	}
}

func shouldRetry(err error) bool {
	return err != nil
}
