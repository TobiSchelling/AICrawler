package cluster

import "math"

// merge records a single merge step in the dendrogram.
type merge struct {
	a, b     int     // indices merged (can be original points or previous clusters)
	distance float64 // merge distance
	size     int     // size of the new cluster
}

// pairwiseDistances computes the squared Euclidean distance matrix (condensed form).
// Returns a flat array of n*(n-1)/2 distances in row-major upper-triangle order.
func pairwiseDistances(embeddings [][]float64) []float64 {
	n := len(embeddings)
	dist := make([]float64, n*(n-1)/2)

	idx := 0
	for i := 0; i < n; i++ {
		for j := i + 1; j < n; j++ {
			var d float64
			for k := range embeddings[i] {
				diff := embeddings[i][k] - embeddings[j][k]
				d += diff * diff
			}
			dist[idx] = d
			idx++
		}
	}
	return dist
}

// condensedIndex returns the index in the condensed distance array for pair (i, j) where i < j.
func condensedIndex(n, i, j int) int {
	if i > j {
		i, j = j, i
	}
	return n*i - i*(i+1)/2 + j - i - 1
}

// wardLinkage performs Ward's agglomerative clustering using the Lance-Williams recurrence.
// Input: condensed squared Euclidean distance matrix, number of points.
// Returns merge history (n-1 merges).
func wardLinkage(dist []float64, n int) []merge {
	// Active cluster tracking
	active := make([]bool, 2*n-1)
	size := make([]int, 2*n-1)
	for i := 0; i < n; i++ {
		active[i] = true
		size[i] = 1
	}

	// Working distance matrix — copy so we can mutate
	d := make([]float64, len(dist))
	copy(d, dist)

	merges := make([]merge, 0, n-1)

	for step := 0; step < n-1; step++ {
		// Find the pair with minimum distance among active clusters
		minDist := math.MaxFloat64
		var minI, minJ int
		for i := 0; i < n+step; i++ {
			if !active[i] {
				continue
			}
			for j := i + 1; j < n+step; j++ {
				if !active[j] {
					continue
				}
				dij := getDist(d, n, i, j)
				if dij < minDist {
					minDist = dij
					minI = i
					minJ = j
				}
			}
		}

		newCluster := n + step
		newSize := size[minI] + size[minJ]
		active[minI] = false
		active[minJ] = false
		active = append(active[:newCluster+1], active[newCluster+1:]...)
		for len(active) <= newCluster {
			active = append(active, false)
		}
		active[newCluster] = true

		for len(size) <= newCluster {
			size = append(size, 0)
		}
		size[newCluster] = newSize

		merges = append(merges, merge{
			a:        minI,
			b:        minJ,
			distance: math.Sqrt(minDist), // scipy reports Euclidean distance, not squared
			size:     newSize,
		})

		// Lance-Williams update: compute distances from new cluster to all other active clusters
		// Ward's formula: d(new, k) = ((n_k + n_i) * d(i,k) + (n_k + n_j) * d(j,k) - n_k * d(i,j)) / (n_k + n_i + n_j)
		for k := 0; k < newCluster; k++ {
			if !active[k] {
				continue
			}
			ni := float64(size[minI])
			nj := float64(size[minJ])
			nk := float64(size[k])

			dik := getDist(d, n, minI, k)
			djk := getDist(d, n, minJ, k)
			dij := minDist // already the squared distance

			newDist := ((nk+ni)*dik + (nk+nj)*djk - nk*dij) / (nk + ni + nj)
			setDist(&d, n, newCluster, k, newDist)
		}
	}

	return merges
}

// cutDendrogram assigns cluster labels by cutting the dendrogram at a threshold.
// Returns cluster assignments (0-indexed) for each of the n original points.
func cutDendrogram(merges []merge, n int, threshold float64) []int {
	// Each point starts in its own cluster
	labels := make([]int, 2*n-1)
	for i := range labels {
		labels[i] = i
	}

	// Process merges — only merge if distance <= threshold
	for step, m := range merges {
		newCluster := n + step
		if m.distance <= threshold {
			// Merge: assign both subtrees to the same label
			labelA := find(labels, m.a)
			labels[newCluster] = labelA
			setLabel(labels, m.b, labelA)
		} else {
			// Don't merge — new cluster gets its own label
			labels[newCluster] = newCluster
		}
	}

	// Resolve final labels for original points and remap to sequential IDs
	finalLabels := make([]int, n)
	labelMap := make(map[int]int)
	nextID := 0

	for i := 0; i < n; i++ {
		root := find(labels, i)
		if _, ok := labelMap[root]; !ok {
			labelMap[root] = nextID
			nextID++
		}
		finalLabels[i] = labelMap[root]
	}

	return finalLabels
}

// find resolves the root label for a node.
func find(labels []int, i int) int {
	for labels[i] != i {
		labels[i] = labels[labels[i]] // path compression
		i = labels[i]
	}
	return i
}

// setLabel sets all nodes in the subtree of b to label.
func setLabel(labels []int, b, label int) {
	for labels[b] != b {
		next := labels[b]
		labels[b] = label
		b = next
	}
	labels[b] = label
}

// getDist reads from the condensed/extended distance structure.
func getDist(d []float64, n, i, j int) float64 {
	if i == j {
		return 0
	}
	if i > j {
		i, j = j, i
	}
	// For original points, use condensed index
	if i < n && j < n {
		idx := condensedIndex(n, i, j)
		if idx < len(d) {
			return d[idx]
		}
	}
	// For new clusters, we store in extended positions
	key := extendedKey(n, i, j)
	if key < len(d) {
		return d[key]
	}
	return 0
}

// setDist writes to the distance structure, extending if needed.
func setDist(d *[]float64, n, i, j int, val float64) {
	if i > j {
		i, j = j, i
	}
	if i < n && j < n {
		idx := condensedIndex(n, i, j)
		(*d)[idx] = val
		return
	}
	key := extendedKey(n, i, j)
	for len(*d) <= key {
		*d = append(*d, 0)
	}
	(*d)[key] = val
}

// extendedKey computes a storage key for distance involving new clusters.
func extendedKey(n, i, j int) int {
	if i > j {
		i, j = j, i
	}
	base := n * (n - 1) / 2
	// Use a simple mapping for extended clusters
	return base + i*(2*n-1) + j
}
