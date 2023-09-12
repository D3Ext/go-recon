package core

import (
  "time"
)

func StartTimer() (time.Time) {
  return time.Now()
}

func TimerDiff(t1 time.Time) (time.Duration) {
  t2 := time.Now()
  diff := t2.Sub(t1)

  return diff.Round(10 * time.Millisecond)
}

