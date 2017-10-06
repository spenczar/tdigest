package tdigest

import (
	"fmt"
	"math"
	"math/rand"
	"reflect"
	"testing"
)

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
		d := TDigest{centroids: tc.centroids}
		have := d.nearest(tc.val)
		if len(tc.want) == 0 {
			if len(have) != 0 {
				t.Errorf("TDigest.nearest wrong test=%d, have=%v, want=%v", i, have, tc.want)
			}
		} else {
			if !reflect.DeepEqual(tc.want, have) {
				t.Errorf("TDigest.nearest wrong test=%d, have=%v, want=%v", i, have, tc.want)
			}
		}
	}
}

func BenchmarkFindNearest(b *testing.B) {
	n := 500
	d := simpleTDigest(n)

	b.ResetTimer()
	var val float64
	for i := int64(0); i < int64(b.N); i++ {
		val = float64(i % d.countTotal)
		_ = d.nearest(val)
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
		d := TDigest{centroids: tc.centroids, countTotal: int64(len(tc.centroids))}
		have := d.findAddTarget(tc.val)
		if have != tc.want {
			t.Errorf("TDigest.findAddTarget wrong test=%d, have=%v, want=%v", i, have, tc.want)
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
		d := tdFromMeans(tc.centroidVals)
		d.addNewCentroid(tc.add, 1)

		have := make([]float64, len(d.centroids))
		for i, c := range d.centroids {
			have[i] = c.mean
		}

		if !reflect.DeepEqual(tc.want, have) {
			t.Errorf("TDigest.addNewCentroid wrong test=%d, have=%v, want=%v", i, have, tc.want)
		}
	}
}

func verifyCentroidOrder(t *testing.T, cs *TDigest) {
	if len(cs.centroids) < 2 {
		return
	}
	last := cs.centroids[0]
	for i, c := range cs.centroids[1:] {
		if c.mean < last.mean {
			t.Errorf("centroid %d lt %d: %v < %v", i+1, i, c.mean, last.mean)
		}
		last = c
	}
}

func TestQuantileOrder(t *testing.T) {
	// stumbled upon in real world application: adding a 1 to this
	// resulted in the 6th centroid getting incremented instead of the
	// 7th.
	d := &TDigest{
		countTotal:  14182,
		compression: 100,
		centroids: []*centroid{
			&centroid{0.000000, 1},
			&centroid{0.000000, 564},
			&centroid{0.000000, 1140},
			&centroid{0.000000, 1713},
			&centroid{0.000000, 2380},
			&centroid{0.000000, 2688},
			&centroid{0.000000, 1262},
			&centroid{2.005758, 1563},
			&centroid{30.499251, 1336},
			&centroid{381.533509, 761},
			&centroid{529.600000, 5},
			&centroid{1065.294118, 17},
			&centroid{2266.444444, 36},
			&centroid{4268.809783, 368},
			&centroid{14964.148148, 27},
			&centroid{41024.579618, 157},
			&centroid{124311.192308, 52},
			&centroid{219674.636364, 22},
			&centroid{310172.775000, 40},
			&centroid{412388.642857, 14},
			&centroid{582867.000000, 16},
			&centroid{701434.777778, 9},
			&centroid{869363.800000, 5},
			&centroid{968264.000000, 1},
			&centroid{987100.666667, 3},
			&centroid{1029895.000000, 1},
			&centroid{1034640.000000, 1},
		},
	}
	d.Add(1.0, 1)
	verifyCentroidOrder(t, d)
}

func TestQuantile(t *testing.T) {
	type testcase struct {
		weights []int64
		idx     int
		want    float64
	}
	testcases := []testcase{
		{[]int64{1, 1, 1, 1}, 0, 0.0},
		{[]int64{1, 1, 1, 1}, 1, 0.25},
		{[]int64{1, 1, 1, 1}, 2, 0.5},
		{[]int64{1, 1, 1, 1}, 3, 0.75},

		{[]int64{5, 1, 1, 1}, 0, 0.250},
		{[]int64{5, 1, 1, 1}, 1, 0.625},
		{[]int64{5, 1, 1, 1}, 2, 0.750},
		{[]int64{5, 1, 1, 1}, 3, 0.875},

		{[]int64{1, 1, 1, 5}, 0, 0.0},
		{[]int64{1, 1, 1, 5}, 1, 0.125},
		{[]int64{1, 1, 1, 5}, 2, 0.250},
		{[]int64{1, 1, 1, 5}, 3, 0.625},
	}

	for i, tc := range testcases {
		d := tdFromWeights(tc.weights)
		have := d.quantileOf(tc.idx)
		if have != tc.want {
			t.Errorf("TDigest.quantile wrong test=%d, have=%.3f, want=%.3f", i, have, tc.want)
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

	d := NewWithCompression(1)
	for i, tc := range testcases {
		d.Add(tc.value, tc.weight)
		if !reflect.DeepEqual(d.centroids, tc.want) {
			t.Fatalf("TDigest.addValue unexpected state step=%d, have=%v, want=%v", i, d.centroids, tc.want)
		}
	}
}

func TestQuantileValue(t *testing.T) {
	d := NewWithCompression(1)
	d.countTotal = 8
	d.centroids = []*centroid{{0.5, 3}, {1, 1}, {2, 2}, {3, 1}, {8, 1}}

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
		have := d.Quantile(tc.q)
		if math.Abs(have-tc.want) > epsilon {
			t.Errorf("TDigest.Quantile wrong step=%d, have=%v, want=%v",
				i, have, tc.want)
		}
	}
}

func BenchmarkFindAddTarget(b *testing.B) {
	n := 500
	d := simpleTDigest(n)

	b.ResetTimer()
	var val float64
	for i := int64(0); i < int64(b.N); i++ {
		val = float64(i % d.countTotal)
		_ = d.findAddTarget(val)
	}
}

// add the values [0,n) to a centroid set, equal weights
func simpleTDigest(n int) *TDigest {
	d := NewWithCompression(1.0)
	for i := 0; i < n; i++ {
		d.Add(float64(i), 1)
	}
	return d
}

func tdFromMeans(means []float64) *TDigest {
	centroids := make([]*centroid, len(means))
	for i, m := range means {
		centroids[i] = &centroid{m, 1}
	}
	d := NewWithCompression(1.0)
	d.centroids = centroids
	d.countTotal = int64(len(centroids))
	return d
}

func tdFromWeights(weights []int64) *TDigest {
	centroids := make([]*centroid, len(weights))
	countTotal := int64(0)
	for i, w := range weights {
		centroids[i] = &centroid{float64(i), w}
		countTotal += w
	}
	d := NewWithCompression(1.0)
	d.centroids = centroids
	d.countTotal = countTotal
	return d
}

func ExampleTDigest() {
	rand.Seed(5678)
	values := make(chan float64)

	// Generate 100k uniform random data between 0 and 100
	var (
		n        int     = 100000
		min, max float64 = 0, 100
	)
	go func() {
		for i := 0; i < n; i++ {
			values <- min + rand.Float64()*(max-min)
		}
		close(values)
	}()

	// Pass the values through a TDigest, compression parameter 100
	td := New()

	for val := range values {
		// Add the value with weight 1
		td.Add(val, 1)
	}

	// Print the 50th, 90th, 99th, 99.9th, and 99.99th percentiles
	fmt.Printf("50th: %.5f\n", td.Quantile(0.5))
	fmt.Printf("90th: %.5f\n", td.Quantile(0.9))
	fmt.Printf("99th: %.5f\n", td.Quantile(0.99))
	fmt.Printf("99.9th: %.5f\n", td.Quantile(0.999))
	fmt.Printf("99.99th: %.5f\n", td.Quantile(0.9999))
}

func TestMerge(t *testing.T) {
	values := make(chan float64)

	// Generate 100k uniform random data between 0 and 100
	var (
		n        int     = 100000
		min, max float64 = 0, 100
	)
	go func() {
		for i := 0; i < n; i++ {
			values <- min + rand.Float64()*(max-min)
		}
		close(values)
	}()

	// Pass the values through two TDigests
	td1 := New()
	td2 := New()

	i := 0
	for val := range values {
		// Add the value with weight 1. Alternate between the digests.
		if i%2 == 0 {
			td1.Add(val, 1)
		} else {
			td2.Add(val, 1)
		}
		i += 1
	}

	rand.Seed(2)
	// merge both into a third tdigest.
	td := New()
	td1.MergeInto(td)
	td2.MergeInto(td)
	fmt.Printf("10th: %.5f\n", td1.Quantile(0.1))
	fmt.Printf("50th: %.5f\n", td1.Quantile(0.5))
	fmt.Printf("90th: %.5f\n", td1.Quantile(0.9))
	fmt.Printf("99th: %.5f\n", td1.Quantile(0.99))
	fmt.Printf("99.9th: %.5f\n", td1.Quantile(0.999))
	fmt.Printf("99.99th: %.5f\n", td1.Quantile(0.9999))

	fmt.Printf("10th: %.5f\n", td2.Quantile(0.1))
	fmt.Printf("50th: %.5f\n", td2.Quantile(0.5))
	fmt.Printf("90th: %.5f\n", td2.Quantile(0.9))
	fmt.Printf("99th: %.5f\n", td2.Quantile(0.99))
	fmt.Printf("99.9th: %.5f\n", td2.Quantile(0.999))
	fmt.Printf("99.99th: %.5f\n", td2.Quantile(0.9999))

	fmt.Printf("10th: %.5f\n", td.Quantile(0.1))
	fmt.Printf("50th: %.5f\n", td.Quantile(0.5))
	fmt.Printf("90th: %.5f\n", td.Quantile(0.9))
	fmt.Printf("99th: %.5f\n", td.Quantile(0.99))
	fmt.Printf("99.9th: %.5f\n", td.Quantile(0.999))
	fmt.Printf("99.99th: %.5f\n", td.Quantile(0.9999))
}
