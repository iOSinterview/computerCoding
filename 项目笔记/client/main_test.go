package main

import (
	"testing"
)

func BenchmarkStartClient(b *testing.B) {
	for i := 0; i < b.N; i++ {
		StartClient("gaocache", "scores")
	}
}
