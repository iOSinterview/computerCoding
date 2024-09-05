package gaocache

// func BenchmarkStartClient(b *testing.B) {
// 	for i := 0; i < b.N; i++ {
// 		StartClient("gaocache", "scores")
// 	}
// }

// func BenchmarkStartClientParallel(b *testing.B) {
// 	// 每个CPU核心启动2个goroutine
// 	b.SetParallelism(5)
// 	b.RunParallel(func(pb *testing.PB) {
// 		for pb.Next() {
// 			StartClient("gaocache", "scores") // 并行测试
// 		}
// 	})
// }
