package tdigest

import (
	"fmt"
	"math"
	"math/rand"
)

// centroid is a simple container for a mean,count pair.
type centroid struct {
	mean  float64
	count int64
}

func (c *centroid) String() string {
	return fmt.Sprintf("c{%f x%d}", c.mean, c.count)
}

// A TDigest is an efficient data structure for computing streaming
// approximate quantiles of a dataset.
type TDigest struct {
	centroids   []*centroid
	compression float64
	countTotal  int64
}

// New produces a new TDigest using the default compression level of
// 100.
func New() *TDigest {
	return NewWithCompression(100)
}

// NewWithCompression produces a new TDigest with a specific
// compression level. The input compression value, which should be >=
// 1.0, will control how aggressively the TDigest compresses data
// together.
//
// The original TDigest paper suggests using a value of 100 for a good
// balance between precision and efficiency. It will land at very
// small (think like 1e-6 percentile points) errors at extreme points
// in the distribution, and compression ratios of around 500 for large
// data sets (1 millionish datapoints).
func NewWithCompression(compression float64) *TDigest {
	return &TDigest{
		centroids:   make([]*centroid, 0),
		compression: compression,
		countTotal:  0,
	}
}

// Find the indexes of centroids which have the minimum distance to the
// input value.
//
// TODO: Use a better data structure to avoid this loop.
func (d *TDigest) nearest(val float64) []int {
	var (
		nearestDist float64 = math.Inf(+1)
		thisDist    float64
		delta       float64
		result      []int = make([]int, 0)
	)
	for i, c := range d.centroids {
		thisDist = val - c.mean
		if thisDist < 0 {
			thisDist *= -1
		}

		delta = thisDist - nearestDist
		switch {
		case delta < 0:
			// we have a new winner!
			nearestDist = thisDist
			result = result[0:0] // wipe result
			result = append(result, i)
		case delta == 0:
			// we have a tie
			result = append(result, i)
		default:
			// Since d.centroids is sorted by mean, this means we
			// have passed the best spot, so we may as well break
			break
		}
	}
	return result
}

// returns the maximum weight that can be placed at specified index
func (d *TDigest) weightLimit(idx int) int64 {
	ptile := d.quantileOf(idx)
	limit := float64(4 * d.compression * ptile * (1 - ptile) * float64(len(d.centroids)))
	return int64(limit)
}

// checks whether the centroid has room for more weight
func (d *TDigest) centroidHasRoom(idx int) bool {
	return d.centroids[idx].count < d.weightLimit(idx)
}

// find which centroid to add the value to (by index)
func (d *TDigest) findAddTarget(val float64) int {
	var (
		nearest  []int = d.nearest(val)
		eligible []int
	)

	for _, c := range nearest {
		// if there is room for more weight at this centroid...
		if d.centroidHasRoom(c) {
			eligible = append(eligible, c)
		}
	}

	if len(eligible) == 0 {
		return -1
	}
	if len(eligible) == 1 {
		return eligible[0]
	}

	// Multiple eligible centroids to add to. They must be equidistant
	// from this value. Four cases are possible:
	//
	//   1. All eligible centroids' means are less than val
	//   2. Some eligible centroids' means are less than val, some are greater
	//   3. All eligible centroids' means are exactly equal to val
	//   4. All eligible centroids' means are greater than val
	//
	// If 1, then we should take the highest indexed centroid to
	// preserve ordering.  If 4, we should take the lowest for the
	// same reason. If 2 or 3, we can pick randomly.

	var anyLesser, anyGreater bool
	for _, c := range eligible {
		m := d.centroids[c].mean
		if m < val {
			anyLesser = true
		} else if m > val {
			anyGreater = true
		}
	}

	// case 1: all are less, none are greater. Take highest one.
	if anyLesser && !anyGreater {
		greatest := eligible[0]
		for _, c := range eligible[1:] {
			if c > greatest {
				greatest = c
			}
		}
		return greatest
	}

	// case 4: all are greater, none are less. Take lowest one.
	if !anyLesser && anyGreater {
		least := eligible[0]
		for _, c := range eligible[1:] {
			if c < least {
				least = c
			}
		}
		return least
	}

	// case 2 or 3: Anything else. Pick randomly.
	return eligible[rand.Intn(len(eligible))]
}

func (d *TDigest) addNewCentroid(mean float64, weight int64) {
	var idx int = len(d.centroids)

	for i, c := range d.centroids {
		// add in sorted order
		if mean < c.mean {
			idx = i
			break
		}
	}

	d.centroids = append(d.centroids, nil)
	copy(d.centroids[idx+1:], d.centroids[idx:])
	d.centroids[idx] = &centroid{mean, weight}
}

// Add will add a value to the TDigest, updating all quantiles. A
// weight can be specified; use weight of 1 if you don't care about
// weighting your dataset.
func (d *TDigest) Add(val float64, weight int64) {
	d.countTotal += weight
	var idx = d.findAddTarget(val)

	if idx == -1 {
		d.addNewCentroid(val, weight)
		return
	}

	c := d.centroids[idx]

	limit := d.weightLimit(idx)
	// how much weight will we be adding?
	// if adding this node to this centroid would put it over the
	// weight limit, just add the most we can and recur with the remainder
	if c.count+weight > limit {
		add := limit - c.count
		if add < 0 {
			// this node was already overweight
			add = 0
		}
		remainder := weight - add

		c.count += add
		c.mean = c.mean + float64(add)*(val-c.mean)/float64(c.count)

		d.Add(val, remainder)
	} else {
		c.count += weight
		c.mean = c.mean + float64(weight)*(val-c.mean)/float64(c.count)
	}

}

// returns the approximate quantile that a particular centroid
// represents
func (d *TDigest) quantileOf(idx int) float64 {
	total := int64(0)
	for _, c := range d.centroids[:idx] {
		total += c.count
	}
	return (float64(d.centroids[idx].count/2) + float64(total)) / float64(d.countTotal)
}

// Quantile returns an estimate estimate the qth quantile value of the dataset.
// The input value of q should be in the range [0.0, 1.0]; if it is outside that

func (d *TDigest) Quantile(q float64) float64 {
	var n = len(d.centroids)
	if n == 0 {
		return math.NaN()
	}
	if n == 1 {
		return d.centroids[0].mean
	}

	if q < 0 {
		q = 0
	} else if q > 1 {
		q = 1
	}

	// rescale into count units instead of 0 to 1 units
	q = float64(d.countTotal) * q
	// find the first centroid which straddles q
	var (
		qTotal float64 = 0
		i      int
	)
	for i = 0; i < n && float64(d.centroids[i].count)/2+qTotal < q; i++ {
		qTotal += float64(d.centroids[i].count)
	}

	if i == 0 {
		// special case 1: the targeted quantile is before the
		// left-most centroid. extrapolate from the slope from
		// centroid0 to centroid1.
		c0 := d.centroids[0]
		c1 := d.centroids[1]
		slope := (c1.mean - c0.mean) / (float64(c1.count)/2 + float64(c0.count)/2)
		deltaQ := q - float64(c0.count)/2 // this is negative
		return c0.mean + slope*deltaQ
	}
	if i == n {
		// special case 2: the targeted quantile is from the
		// right-most centroid. extrapolate from the slope at the
		// right edge.
		c0 := d.centroids[n-2]
		c1 := d.centroids[n-1]
		slope := (c1.mean - c0.mean) / (float64(c1.count)/2 + float64(c0.count)/2)
		deltaQ := q - (qTotal - float64(c1.count)/2)
		return c1.mean + slope*deltaQ
	}
	// common case: targeted quantile is between 2 centroids
	c0 := d.centroids[i-1]
	c1 := d.centroids[i]
	slope := (c1.mean - c0.mean) / (float64(c1.count)/2 + float64(c0.count)/2)
	deltaQ := q - (float64(c1.count)/2 + qTotal)
	return c1.mean + slope*deltaQ
}

// MergeInto(other) will add all of the data within a TDigest into other,
// combining them into one larger TDigest.
func (d *TDigest) MergeInto(other *TDigest) {
	// Add each centroid in d into other. They should be added in
	// order of decreasing weight.
	addOrder := rand.Perm(len(d.centroids))
	for _, idx := range addOrder {
		c := d.centroids[idx]
		// gradually write up the volume written so that the tdigest doesnt overload
		// early
		added := int64(0)
		for i := int64(1); i < 10; i++ {
			toAdd := i * 2
			if added+toAdd > c.count {
				toAdd = c.count - added
			}
			other.Add(c.mean, toAdd)
			added += toAdd
			if added >= c.count {
				break
			}
		}
		if added < c.count {
			other.Add(c.mean, c.count-added)
		}
		other.Add(c.mean, c.count)
	}
}

// MarshalBinary encodes the TDigest in the same manner as the Java reference
// implementration at github.com/tdunning/t-digest. The encoded bytes can be
// read back out with UnmarshalBinary.
func (d *TDigest) MarshalBinary() ([]byte, error) {
	return marshalBinary(d)
}

// UnmarshalBinary unpacks the given bytes as a TDigest, using the same format
// as MarshalBinary.
func (d *TDigest) UnmarshalBinary(p []byte) error {
	return unmarshalBinary(d, p)
}
