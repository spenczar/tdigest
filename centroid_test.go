package tdigest

import (
	"math"
	"math/rand"
	"reflect"
	"testing"
)

const rngSeed = 1234567

func TestFindNearest(t *testing.T) {
	type testcase struct {
		centroids []*centroid
		val       float64
		want      []int
	}

	testcases := []testcase{
		{[]*centroid{{0, 1}, {1, 1}, {2, 1}}, -1, []int{0}},
		{[]*centroid{{0, 1}, {1, 1}, {2, 1}}, 0, []int{0}},
		{[]*centroid{{0, 1}, {1, 1}, {2, 1}}, 1, []int{1}},
		{[]*centroid{{0, 1}, {1, 1}, {2, 1}}, 2, []int{2}},
		{[]*centroid{{0, 1}, {1, 1}, {2, 1}}, 3, []int{2}},
		{[]*centroid{{0, 1}, {2, 1}}, 1, []int{0, 1}},
		{[]*centroid{}, 1, []int{}},
	}

	for i, tc := range testcases {
		cs := centroidSet{centroids: tc.centroids}
		have := cs.nearest(tc.val)
		if len(tc.want) == 0 {
			if len(have) != 0 {
				t.Errorf("centroidSet.nearest wrong test=%d, have=%v, want=%v", i, have, tc.want)
			}
		} else {
			if !reflect.DeepEqual(tc.want, have) {
				t.Errorf("centroidSet.nearest wrong test=%d, have=%v, want=%v", i, have, tc.want)
			}
		}
	}
}

func BenchmarkFindNearest(b *testing.B) {
	n := 500
	cset := simpleCentroidSet(n)

	b.ResetTimer()
	var val float64
	for i := 0; i < b.N; i++ {
		val = float64(i % cset.countTotal)
		_ = cset.nearest(val)
	}
}

func TestFindAddTarget(t *testing.T) {
	type testcase struct {
		centroids []*centroid
		val       float64
		want      int
	}

	testcases := []testcase{
		{[]*centroid{}, 1, -1},
	}
	for i, tc := range testcases {
		cs := centroidSet{centroids: tc.centroids, countTotal: len(tc.centroids)}
		have := cs.findAddTarget(tc.val)
		if have != tc.want {
			t.Errorf("centroidSet.findAddTarget wrong test=%d, have=%v, want=%v", i, have, tc.want)
		}
	}
}

// adding a new centroid should maintain sorted order
func TestAddNewCentroid(t *testing.T) {
	type testcase struct {
		centroidVals []float64
		add          float64
		want         []float64
	}
	testcases := []testcase{
		{[]float64{}, 1, []float64{1}},
		{[]float64{1}, 2, []float64{1, 2}},
		{[]float64{1, 2}, 1.5, []float64{1, 1.5, 2}},
		{[]float64{1, 1.5, 2}, -1, []float64{-1, 1, 1.5, 2}},
		{[]float64{1, 1.5, 2}, 3, []float64{1, 1.5, 2, 3}},
		{[]float64{1, 1.5, 2}, 1.6, []float64{1, 1.5, 1.6, 2}},
	}

	for i, tc := range testcases {
		cset := csetFromMeans(tc.centroidVals)
		cset.addNewCentroid(tc.add, 1)

		have := make([]float64, len(cset.centroids))
		for i, c := range cset.centroids {
			have[i] = c.mean
		}

		if !reflect.DeepEqual(tc.want, have) {
			t.Errorf("centroidSet.addNewCentroid wrong test=%d, have=%v, want=%v", i, have, tc.want)
		}
	}
}

func TestQuantile(t *testing.T) {
	type testcase struct {
		weights []int
		idx     int
		want    float64
	}
	testcases := []testcase{
		{[]int{1, 1, 1, 1}, 0, 0.0},
		{[]int{1, 1, 1, 1}, 1, 0.25},
		{[]int{1, 1, 1, 1}, 2, 0.5},
		{[]int{1, 1, 1, 1}, 3, 0.75},

		{[]int{5, 1, 1, 1}, 0, 0.250},
		{[]int{5, 1, 1, 1}, 1, 0.625},
		{[]int{5, 1, 1, 1}, 2, 0.750},
		{[]int{5, 1, 1, 1}, 3, 0.875},

		{[]int{1, 1, 1, 5}, 0, 0.0},
		{[]int{1, 1, 1, 5}, 1, 0.125},
		{[]int{1, 1, 1, 5}, 2, 0.250},
		{[]int{1, 1, 1, 5}, 3, 0.625},
	}

	for i, tc := range testcases {
		cset := csetFromWeights(tc.weights)
		have := cset.quantile(tc.idx)
		if have != tc.want {
			t.Errorf("centroidSet.quantile wrong test=%d, have=%.3f, want=%.3f", i, have, tc.want)
		}
	}
}

func TestAddValue(t *testing.T) {
	type testcase struct {
		value  float64
		weight int
		want   []*centroid
	}

	testcases := []testcase{
		{1.0, 1, []*centroid{{1, 1}}},
		{0.0, 1, []*centroid{{0, 1}, {1, 1}}},
		{2.0, 1, []*centroid{{0, 1}, {1, 1}, {2, 1}}},
		{3.0, 1, []*centroid{{0, 1}, {1, 1}, {2.5, 2}}},
		{4.0, 1, []*centroid{{0, 1}, {1, 1}, {2.5, 2}, {4, 1}}},
	}

	cset := newCentroidSet(1)
	for i, tc := range testcases {
		cset.addValue(tc.value, tc.weight)
		if !reflect.DeepEqual(cset.centroids, tc.want) {
			t.Fatalf("centroidSet.addValue unexpected state step=%d, have=%v, want=%v", i, cset.centroids, tc.want)
		}
	}
}

func TestQuantileValue(t *testing.T) {
	cset := newCentroidSet(1)
	cset.countTotal = 8
	cset.centroids = []*centroid{{0.5, 3}, {1, 1}, {2, 2}, {3, 1}, {8, 1}}

	type testcase struct {
		q    float64
		want float64
	}

	// correct values, determined by hand with pen and paper for this set of centroids
	testcases := []testcase{
		{0.0, 5.0 / 40.0},
		{0.1, 13.0 / 40.0},
		{0.2, 21.0 / 40.0},
		{0.3, 29.0 / 40.0},
		{0.4, 37.0 / 40.0},
		{0.5, 20.0 / 15.0},
		{0.6, 28.0 / 15.0},
		{0.7, 36.0 / 15.0},
		{0.8, 44.0 / 15.0},
		{0.9, 13.0 / 2.0},
		{1.0, 21.0 / 2.0},
	}

	var epsilon = 1e-8

	for i, tc := range testcases {
		have := cset.quantileValue(tc.q)
		if math.Abs(have-tc.want) > epsilon {
			t.Errorf("centroidSet.quantileValue wrong step=%d, have=%v, want=%v",
				i, have, tc.want)
		}
	}
}

func BenchmarkFindAddTarget(b *testing.B) {
	n := 500
	cset := simpleCentroidSet(n)

	b.ResetTimer()
	var val float64
	for i := 0; i < b.N; i++ {
		val = float64(i % cset.countTotal)
		_ = cset.findAddTarget(val)
	}
}

type valueSource interface {
	Next() float64
}

type orderedValues struct {
	last float64
}

func (ov *orderedValues) Next() float64 {
	ov.last += 1
	return ov.last
}

func benchmarkAdd(b *testing.B, n int, src valueSource) {
	valsToAdd := make([]float64, n)

	cset := newCentroidSet(100)
	for i := 0; i < n; i++ {
		v := src.Next()
		valsToAdd[i] = v
		cset.addValue(v, 1)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cset.addValue(valsToAdd[i%n], 1)
	}
	b.StopTimer()
	b.Logf("%d data points compresses to %d centroids", cset.countTotal, len(cset.centroids))
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

// add the values [0,n) to a centroid set, equal weights
func simpleCentroidSet(n int) *centroidSet {
	cset := newCentroidSet(1.0)
	for i := 0; i < n; i++ {
		cset.addValue(float64(i), 1)
	}
	return cset
}

func csetFromMeans(means []float64) *centroidSet {
	centroids := make([]*centroid, len(means))
	for i, m := range means {
		centroids[i] = &centroid{m, 1}
	}
	cset := newCentroidSet(1.0)
	cset.centroids = centroids
	cset.countTotal = len(centroids)
	return cset
}

func csetFromWeights(weights []int) *centroidSet {
	centroids := make([]*centroid, len(weights))
	countTotal := 0
	for i, w := range weights {
		centroids[i] = &centroid{float64(i), w}
		countTotal += w
	}
	cset := newCentroidSet(1.0)
	cset.centroids = centroids
	cset.countTotal = countTotal
	return cset
}
