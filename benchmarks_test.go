package tdigest

import (
	"math/rand"
	"testing"
)

const rngSeed = 1234567

type valueSource interface {
	Next() float64
}

func benchmarkAdd(b *testing.B, n int, src valueSource) {
	valsToAdd := make([]float64, n)

	cset := newCentroidSet(100)
	for i := 0; i < n; i++ {
		v := src.Next()
		valsToAdd[i] = v
		cset.Add(v, 1)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cset.Add(valsToAdd[i%n], 1)
	}
	b.StopTimer()
}

func benchmarkQuantile(b *testing.B, n int, src valueSource) {
	quantilesToCheck := make([]float64, n)

	cset := newCentroidSet(100)
	for i := 0; i < n; i++ {
		v := src.Next()
		quantilesToCheck[i] = v
		cset.Add(v, 1)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = cset.Quantile(quantilesToCheck[i%n])
	}
	b.StopTimer()
}

type orderedValues struct {
	last float64
}

func (ov *orderedValues) Next() float64 {
	ov.last += 1
	return ov.last
}

func BenchmarkAdd_1k_Ordered(b *testing.B) {
	benchmarkAdd(b, 1000, &orderedValues{})
}

func BenchmarkAdd_10k_Ordered(b *testing.B) {
	benchmarkAdd(b, 10000, &orderedValues{})
}

func BenchmarkAdd_100k_Ordered(b *testing.B) {
	benchmarkAdd(b, 100000, &orderedValues{})
}

func BenchmarkQuantile_1k_Ordered(b *testing.B) {
	benchmarkQuantile(b, 1000, &orderedValues{})
}

func BenchmarkQuantile_10k_Ordered(b *testing.B) {
	benchmarkQuantile(b, 10000, &orderedValues{})
}

func BenchmarkQuantile_100k_Ordered(b *testing.B) {
	benchmarkQuantile(b, 100000, &orderedValues{})
}

type zipfValues struct {
	z *rand.Zipf
}

func newZipfValues() *zipfValues {
	r := rand.New(rand.NewSource(rngSeed))
	z := rand.NewZipf(r, 1.2, 1, 1024*1024)
	return &zipfValues{
		z: z,
	}
}

func (zv *zipfValues) Next() float64 {
	return float64(zv.z.Uint64())
}

func BenchmarkAdd_1k_Zipfian(b *testing.B) {
	benchmarkAdd(b, 1000, newZipfValues())
}

func BenchmarkAdd_10k_Zipfian(b *testing.B) {
	benchmarkAdd(b, 10000, newZipfValues())
}

func BenchmarkAdd_100k_Zipfian(b *testing.B) {
	benchmarkAdd(b, 100000, newZipfValues())
}

func BenchmarkQuantile_1k_Zipfian(b *testing.B) {
	benchmarkQuantile(b, 1000, newZipfValues())
}

func BenchmarkQuantile_10k_Zipfian(b *testing.B) {
	benchmarkQuantile(b, 10000, newZipfValues())
}

func BenchmarkQuantile_100k_Zipfian(b *testing.B) {
	benchmarkQuantile(b, 100000, newZipfValues())
}

type uniformValues struct {
	r *rand.Rand
}

func newUniformValues() *uniformValues {
	return &uniformValues{rand.New(rand.NewSource(rngSeed))}
}

func (uv *uniformValues) Next() float64 {
	return uv.r.Float64()
}

func BenchmarkAdd_1k_Uniform(b *testing.B) {
	benchmarkAdd(b, 1000, newUniformValues())
}

func BenchmarkAdd_10k_Uniform(b *testing.B) {
	benchmarkAdd(b, 10000, newUniformValues())
}

func BenchmarkAdd_100k_Uniform(b *testing.B) {
	benchmarkAdd(b, 100000, newUniformValues())
}

func BenchmarkQuantile_1k_Uniform(b *testing.B) {
	benchmarkQuantile(b, 1000, newUniformValues())
}

func BenchmarkQuantile_10k_Uniform(b *testing.B) {
	benchmarkQuantile(b, 10000, newUniformValues())
}

func BenchmarkQuantile_100k_Uniform(b *testing.B) {
	benchmarkQuantile(b, 100000, newUniformValues())
}

type normalValues struct {
	r *rand.Rand
}

func newNormalValues() *normalValues {
	return &normalValues{rand.New(rand.NewSource(rngSeed))}
}

func (uv *normalValues) Next() float64 {
	return uv.r.NormFloat64()
}

func BenchmarkAdd_1k_Normal(b *testing.B) {
	benchmarkAdd(b, 1000, newNormalValues())
}

func BenchmarkAdd_10k_Normal(b *testing.B) {
	benchmarkAdd(b, 10000, newNormalValues())
}

func BenchmarkAdd_100k_Normal(b *testing.B) {
	benchmarkAdd(b, 100000, newNormalValues())
}

func BenchmarkQuantile_1k_Normal(b *testing.B) {
	benchmarkQuantile(b, 1000, newNormalValues())
}

func BenchmarkQuantile_10k_Normal(b *testing.B) {
	benchmarkQuantile(b, 10000, newNormalValues())
}

func BenchmarkQuantile_100k_Normal(b *testing.B) {
	benchmarkQuantile(b, 100000, newNormalValues())
}
