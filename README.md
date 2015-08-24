# tdigest #

This is a Go implementation of Ted Dunning's
[t-digest](https://github.com/tdunning/t-digest). It's pretty
barebones, but fast enough to get the job done. It is not currently
safe for concurrent access across goroutines.

The current implementation is backed by slices, which is not as
efficient as using some kind of balanced tree. I plan on adding a tree
implementation when I have time.

## Benchmarks ##

```
BenchmarkAdd_1k_Ordered	 1000000	      3274 ns/op	    1015 B/op	      75 allocs/op
--- BENCH: BenchmarkAdd_1k_Ordered
	centroid_test.go:238: 1001 data points compresses to 7 centroids
	centroid_test.go:238: 1100 data points compresses to 7 centroids
	centroid_test.go:238: 11000 data points compresses to 31 centroids
	centroid_test.go:238: 1001000 data points compresses to 289 centroids
BenchmarkAdd_10k_Ordered	 1000000	      3033 ns/op	    1084 B/op	      96 allocs/op
--- BENCH: BenchmarkAdd_10k_Ordered
	centroid_test.go:238: 10001 data points compresses to 52 centroids
	centroid_test.go:238: 10100 data points compresses to 52 centroids
	centroid_test.go:238: 20000 data points compresses to 76 centroids
	centroid_test.go:238: 1010000 data points compresses to 316 centroids
BenchmarkAdd_100k_Ordered	 1000000	      6379 ns/op	    2405 B/op	     288 allocs/op
--- BENCH: BenchmarkAdd_100k_Ordered
	centroid_test.go:238: 100001 data points compresses to 500 centroids
	centroid_test.go:238: 100100 data points compresses to 501 centroids
	centroid_test.go:238: 110000 data points compresses to 501 centroids
	centroid_test.go:238: 1100000 data points compresses to 649 centroids
BenchmarkAdd_1k_Zipfian	 1000000	      1764 ns/op	     624 B/op	      26 allocs/op
--- BENCH: BenchmarkAdd_1k_Zipfian
	centroid_test.go:238: 1001 data points compresses to 9 centroids
	centroid_test.go:238: 1100 data points compresses to 9 centroids
	centroid_test.go:238: 11000 data points compresses to 23 centroids
	centroid_test.go:238: 1001000 data points compresses to 241 centroids
BenchmarkAdd_10k_Zipfian	 1000000	      1407 ns/op	     603 B/op	      24 allocs/op
--- BENCH: BenchmarkAdd_10k_Zipfian
	centroid_test.go:238: 10001 data points compresses to 23 centroids
	centroid_test.go:238: 10100 data points compresses to 23 centroids
	centroid_test.go:238: 20000 data points compresses to 35 centroids
	centroid_test.go:238: 1010000 data points compresses to 250 centroids
BenchmarkAdd_100k_Zipfian	 1000000	      1263 ns/op	     652 B/op	      26 allocs/op
--- BENCH: BenchmarkAdd_100k_Zipfian
	centroid_test.go:238: 100001 data points compresses to 83 centroids
	centroid_test.go:238: 100100 data points compresses to 83 centroids
	centroid_test.go:238: 110000 data points compresses to 88 centroids
	centroid_test.go:238: 1100000 data points compresses to 274 centroids
BenchmarkAdd_1k_Uniform	 1000000	      3158 ns/op	    1004 B/op	      76 allocs/op
--- BENCH: BenchmarkAdd_1k_Uniform
	centroid_test.go:238: 1001 data points compresses to 13 centroids
	centroid_test.go:238: 1100 data points compresses to 13 centroids
	centroid_test.go:238: 11000 data points compresses to 31 centroids
	centroid_test.go:238: 1001000 data points compresses to 289 centroids
BenchmarkAdd_10k_Uniform	 1000000	      3062 ns/op	    1113 B/op	      91 allocs/op
--- BENCH: BenchmarkAdd_10k_Uniform
	centroid_test.go:238: 10001 data points compresses to 28 centroids
	centroid_test.go:238: 10100 data points compresses to 28 centroids
	centroid_test.go:238: 20000 data points compresses to 42 centroids
	centroid_test.go:238: 1010000 data points compresses to 314 centroids
BenchmarkAdd_100k_Uniform	 1000000	      3180 ns/op	    1197 B/op	     123 allocs/op
--- BENCH: BenchmarkAdd_100k_Uniform
	centroid_test.go:238: 100001 data points compresses to 104 centroids
	centroid_test.go:238: 100100 data points compresses to 104 centroids
	centroid_test.go:238: 110000 data points compresses to 109 centroids
	centroid_test.go:238: 1100000 data points compresses to 361 centroids
BenchmarkAdd_1k_Normal	 1000000	      3312 ns/op	    1004 B/op	      76 allocs/op
--- BENCH: BenchmarkAdd_1k_Normal
	centroid_test.go:238: 1001 data points compresses to 7 centroids
	centroid_test.go:238: 1100 data points compresses to 8 centroids
	centroid_test.go:238: 11000 data points compresses to 30 centroids
	centroid_test.go:238: 1001000 data points compresses to 290 centroids
BenchmarkAdd_10k_Normal	 1000000	      3167 ns/op	    1139 B/op	      95 allocs/op
--- BENCH: BenchmarkAdd_10k_Normal
	centroid_test.go:238: 10001 data points compresses to 28 centroids
	centroid_test.go:238: 10100 data points compresses to 28 centroids
	centroid_test.go:238: 20000 data points compresses to 43 centroids
	centroid_test.go:238: 1010000 data points compresses to 316 centroids
BenchmarkAdd_100k_Normal	 1000000	      3220 ns/op	    1196 B/op	     121 allocs/op
--- BENCH: BenchmarkAdd_100k_Normal
	centroid_test.go:238: 100001 data points compresses to 103 centroids
	centroid_test.go:238: 100100 data points compresses to 103 centroids
	centroid_test.go:238: 110000 data points compresses to 108 centroids
	centroid_test.go:238: 1100000 data points compresses to 352 centroids
```
