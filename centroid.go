package tdigest

import (
	"fmt"
	"log"
	"math"
	"math/rand"
)

type centroid struct {
	mean  float64
	count int
}

func (c *centroid) String() string {
	return fmt.Sprintf("c{%dx @ %f}", c.count, c.mean)
}

type centroidSet struct {
	centroids   []*centroid
	compression float64
	countTotal  int

	reclusterAt int
}

// create a new centroid set. 100.0 is a reasonable value for compression.
func newCentroidSet(compression float64) *centroidSet {
	return &centroidSet{
		centroids:   make([]*centroid, 0),
		compression: compression,
		countTotal:  0,
		reclusterAt: 4,
	}
}

func (cs *centroidSet) String() string {
	out := "["
	for i, c := range cs.centroids {
		out += fmt.Sprintf("c[%f x%d]  limit=%d  ptile=%f\n", c.mean, c.count, cs.weightLimit(i), cs.quantile(i))
	}
	out += "]"
	return out
}

// run all centroids through the adding algorithm, which tends to
// shrink the number of clusters without much data loss
func (cs *centroidSet) recluster() {
	if true {
		return
	}
	log.Printf("recluster start from %d", len(cs.centroids))
	cCopy := make([]*centroid, len(cs.centroids))
	copy(cCopy, cs.centroids)

	// shuffle the order of the copy
	for i := range cCopy {
		j := rand.Intn(i + 1)
		cCopy[i], cCopy[j] = cCopy[j], cCopy[i]
	}
	log.Printf("preshuffle from %d: %v", len(cs.centroids), cs)
	cs.centroids = cCopy
	log.Printf("shuffled from %d: %v", len(cs.centroids), cs)

	cs.centroids = make([]*centroid, 0)
	cs.countTotal = 0
	for _, c := range cCopy {
		cs.addValue(c.mean, c.count)
	}
	cs.reclusterAt *= 2
	log.Printf("recluster - size went from %d to %d, total weight=%d", len(cCopy), len(cs.centroids), cs.countTotal)
}

// Find the indexes of centroids which have the minimum distance to the
// input value.
//
// TODO: Use a better data structure to avoid this loop.
func (cs *centroidSet) nearest(val float64) []int {
	var (
		nearestDist float64
		thisDist    float64
		result      []int
	)
	for i, c := range cs.centroids {
		thisDist = math.Abs(val - c.mean)
		// on the first loop, the nearest hasn't yet been found, so
		// mark the first centroid as a 'winner.'
		if len(result) == 0 {
			nearestDist = thisDist
			result = []int{i}
			continue
		}
		if thisDist < nearestDist {
			// we have a new winner!
			nearestDist = thisDist
			result = []int{i}
			continue
		}
		if thisDist == nearestDist {
			// we have a tie
			result = append(result, i)
			continue
		}
		if thisDist > nearestDist {
			// Since cs.centroids is sorted by mean, this means we
			// have passed the best spot, so we may as well break
			break
		}
	}
	return result
}

// returns the maximum weight that can be placed at specified index
func (cs *centroidSet) weightLimit(idx int) int {
	ptile := cs.quantile(idx)
	limit := int(4 * cs.compression * ptile * (1 - ptile) * float64(len(cs.centroids)))
	return limit
}

// checks whether the centroid has room for more weight
func (cs *centroidSet) centroidHasRoom(idx int) bool {
	return cs.centroids[idx].count < cs.weightLimit(idx)
}

// find which centroid to add the value to (by index)
func (cs *centroidSet) findAddTarget(val float64) int {
	var (
		nMatch  int   = 0
		addTo   int   = -1
		nearest []int = cs.nearest(val)
	)
	for _, c := range nearest {
		// if there is room for more weight at this centroid...
		if cs.centroidHasRoom(c) {
			nMatch += 1
			// ... and it passes a random filter...
			r := rand.Intn(nMatch)

			if r == 0 {
				// ... then add it here
				addTo = c
			}
		}
	}
	return addTo
}

func (cs *centroidSet) addNewCentroid(mean float64, weight int) {
	var idx int = len(cs.centroids)

	for i, c := range cs.centroids {
		// add in sorted order
		if mean < c.mean {
			idx = i
			break
		}
	}

	cs.centroids = append(cs.centroids, nil)
	copy(cs.centroids[idx+1:], cs.centroids[idx:])
	cs.centroids[idx] = &centroid{mean, weight}
}

func (cs *centroidSet) addValue(val float64, weight int) {
	cs.countTotal += weight
	var idx = cs.findAddTarget(val)

	if idx == -1 {
		cs.addNewCentroid(val, weight)
		return
	}

	c := cs.centroids[idx]

	limit := cs.weightLimit(idx)
	// how much weight will we be adding?
	// if adding this node to this centroid would put it over the
	// weight limit, just add the most we can and recur with the remainder
	if c.count+weight > limit {
		log.Printf("splitting value %f weight=%d", val, weight)
		add := limit - c.count
		if add < 0 {
			// this node was already overweight
			add = 0
		}
		remainder := weight - add

		c.count += add
		c.mean = c.mean + float64(add)*(val-c.mean)/float64(c.count)

		cs.addValue(val, remainder)
	} else {
		c.count += weight
		c.mean = c.mean + float64(weight)*(val-c.mean)/float64(c.count)
	}

	if len(cs.centroids) >= cs.reclusterAt {
		cs.recluster()
	}
}

// returns the approximate quantile that a particular centroid
// represents
func (cs *centroidSet) quantile(idx int) float64 {
	total := 0
	for _, c := range cs.centroids[:idx] {
		total += c.count
	}
	return (float64(cs.centroids[idx].count/2) + float64(total)) / float64(cs.countTotal)
}

// returns the approximate value of a quantile (eg the 99th percentile
// value would be centroidSet.quantileValue(0.99))
func (cs *centroidSet) quantileValue(q float64) float64 {
	var n = len(cs.centroids)
	if n == 0 {
		return math.NaN()
	}
	if n == 1 {
		return cs.centroids[0].mean
	}

	if q < 0 {
		q = 0
	} else if q > 1 {
		q = 1
	}

	// rescale into count units instead of 0 to 1 units
	q = float64(cs.countTotal) * q
	// find the first centroid which straddles q
	var (
		qTotal float64 = 0
		i      int
	)
	for i = 0; i < n && float64(cs.centroids[i].count)/2+qTotal < q; i++ {
		qTotal += float64(cs.centroids[i].count)
	}

	if i == 0 {
		// special case 1: the targeted quantile is before the
		// left-most centroid. extrapolate from the slope from
		// centroid0 to centroid1.
		c0 := cs.centroids[0]
		c1 := cs.centroids[1]
		slope := (c1.mean - c0.mean) / (float64(c1.count)/2 + float64(c0.count)/2)
		deltaQ := q - float64(c0.count)/2 // this is negative
		return c0.mean + slope*deltaQ
	}
	if i == n {
		// special case 2: the targeted quantile is from the
		// right-most centroid. extrapolate from the slope at the
		// right edge.
		c0 := cs.centroids[n-2]
		c1 := cs.centroids[n-1]
		slope := (c1.mean - c0.mean) / (float64(c1.count)/2 + float64(c0.count)/2)
		deltaQ := q - (qTotal - float64(c1.count)/2)
		return c1.mean + slope*deltaQ
	}
	// common case: targeted quantile is between 2 centroids
	c0 := cs.centroids[i-1]
	c1 := cs.centroids[i]
	slope := (c1.mean - c0.mean) / (float64(c1.count)/2 + float64(c0.count)/2)
	deltaQ := q - (float64(c1.count)/2 + qTotal)
	return c1.mean + slope*deltaQ
}
