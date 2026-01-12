package jitter_test

import (
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	. "go.octolab.org/toolset/maintainer/internal/pkg/time/jitter"
)

func TestFullCustom(t *testing.T) {
	jitter := FullCustom(rand.New(rand.NewSource(1)))

	t.Run("positive duration", func(t *testing.T) {
		for range 1000 {
			d := jitter.Apply(time.Minute)
			assert.GreaterOrEqual(t, d, time.Duration(0))
			assert.Less(t, d, time.Minute)
		}
	})

	t.Run("no room", func(t *testing.T) {
		assert.Zero(t, jitter.Apply(0))
		assert.Zero(t, jitter.Apply(-time.Minute))
	})
}
