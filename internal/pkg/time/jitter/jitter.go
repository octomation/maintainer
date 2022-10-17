package jitter

import (
	"math/rand"
	"time"
)

type Transformation func(time.Duration) time.Duration

func (fn Transformation) Apply(d time.Duration) time.Duration { return fn(d) }

func FullCustom(generator *rand.Rand) Transformation {
	return func(duration time.Duration) time.Duration {
		return time.Duration(generator.Int63n(int64(duration)))
	}
}

func FullRandom() Transformation {
	return FullCustom(rand.New(rand.NewSource(time.Now().UnixNano())))
}
