package jitter

import (
	"math/rand"
	"time"
)

type Transformation func(time.Duration) time.Duration

func (fn Transformation) Apply(d time.Duration) time.Duration { return fn(d) }

// FullCustom returns a random duration in [0, duration),
// or zero if the duration is not positive.
func FullCustom(generator *rand.Rand) Transformation {
	return func(duration time.Duration) time.Duration {
		if duration <= 0 {
			return 0
		}
		return time.Duration(generator.Int63n(int64(duration)))
	}
}

func FullRandom() Transformation {
	return FullCustom(rand.New(rand.NewSource(time.Now().UnixNano())))
}
