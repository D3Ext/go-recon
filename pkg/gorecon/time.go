package gorecon

import (
	"github.com/D3Ext/go-recon/core"
	"time"
)

func StartTimer() time.Time {
	return time.Now()
}

func TimerDiff(t1 time.Time) time.Duration {
	return core.TimerDiff(t1)
}
