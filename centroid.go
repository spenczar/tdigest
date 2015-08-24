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
	centroids  []*centroid
	accuracy   float64
	countTotal int
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
	log.Printf("ptile at %d is %f", idx, ptile)
	limit := int(4*cs.accuracy*ptile*(1-ptile)) * len(cs.centroids)
	log.Printf("weight limit at %d is %d", idx, limit)
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
			log.Printf("room found at %d for %f. nmatch", c, val)
			// ... and it passes a random filter...
			r := rand.Intn(nMatch)

			if r == 0 {
				// ... then add it here
				log.Printf("random draw success")
				addTo = c
			} else {
				log.Printf("random draw fail: r=%f", r)
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

func (cs *centroidSet) quantileValue(q float64) float64 {
	var (
		qtile        float64
		qtileIdx     int
		lastQtile    float64
		lastQtileIdx int
	)
	if q == 1 {
		c1 := cs.centroids[len(cs.centroids)-1]
		c0 := cs.centroids[len(cs.centroids)-2]
		return c1.mean + (c1.mean - c0.mean)
	}
	total := 0
	for i, c := range cs.centroids {
		total += c.count
		qtile = (float64(c.count-1)/2 + float64(total)) / float64(cs.countTotal)
		qtileIdx = i
		if qtile > q {
			log.Printf("believe qtile %f to be between %d and %d (qtiles %f and %f, valus %f and %f)", q, lastQtileIdx, qtileIdx, lastQtile, qtile, cs.centroids[lastQtileIdx].mean, cs.centroids[qtileIdx].mean)
			break
		}
		lastQtile = qtile
		lastQtileIdx = qtileIdx
	}
	// interpolate between the two values
	delta := cs.centroids[qtileIdx].mean - cs.centroids[lastQtileIdx].mean
	log.Printf("delta means: %f", delta)
	delta *= (q - lastQtile) / (qtile - lastQtile)
	log.Printf("delta qtiles: %f", delta)
	log.Printf("adding %f to %f", delta, cs.centroids[qtileIdx].mean)
	return cs.centroids[qtileIdx].mean + delta
}
