package infrastructure

import (
	"context"
	"time"
)

type Clock interface {
	Now(ctx context.Context) time.Time
}

type SystemClock struct{}

func (SystemClock) Now(context.Context) time.Time {
	return time.Now().UTC()
}
