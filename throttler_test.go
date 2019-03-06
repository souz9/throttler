package throttler

import (
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func BenchmarkThrottler(b *testing.B) {
	tr := New(time.Second)

	b.Run("Allow (no delay)", func(b *testing.B) {
		for n := 0; n < b.N; n++ {
			tr.Allow()
		}
	})

	b.Run("Allow (no delay, parallel)", func(b *testing.B) {
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				tr.Allow()
			}
		})
	})

	tr.Down()
	require.True(b, tr.delay > 0)

	b.Run("Allow", func(b *testing.B) {
		for n := 0; n < b.N; n++ {
			tr.Allow()
		}
	})

	b.Run("Allow (parallel)", func(b *testing.B) {
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				tr.Allow()
			}
		})
	})
}
