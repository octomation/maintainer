package jitter

import (
	"math/rand"
	"time"
)

type Transformation = func(time.Duration) time.Duration

func Full(generator *rand.Rand) Transformation {
	return func(duration time.Duration) time.Duration {
		return time.Duration(generator.Int63n(int64(duration)))
	}
}

func FullRandom() Transformation {
	return Full(rand.New(rand.NewSource(time.Now().UnixNano())))
}
