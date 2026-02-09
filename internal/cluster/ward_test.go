package cluster

import (
	"math"
	"testing"
)

func TestPairwiseDistances(t *testing.T) {
	embeddings := [][]float64{
		{1.0, 0.0},
		{0.0, 1.0},
		{1.0, 1.0},
	}
	dist := pairwiseDistances(embeddings)

	// d(0,1) = (1-0)^2 + (0-1)^2 = 2
	// d(0,2) = (1-1)^2 + (0-1)^2 = 1
	// d(1,2) = (0-1)^2 + (1-1)^2 = 1
	expected := []float64{2.0, 1.0, 1.0}

	if len(dist) != len(expected) {
		t.Fatalf("expected %d distances, got %d", len(expected), len(dist))
	}
	for i, d := range dist {
		if math.Abs(d-expected[i]) > 1e-10 {
			t.Errorf("dist[%d] = %f, expected %f", i, d, expected[i])
		}
	}
}

func TestWardLinkageSimple(t *testing.T) {
	// 4 points: 3 similar + 1 outlier
	embeddings := [][]float64{
		{1.0, 0.0, 0.0},
		{0.95, 0.05, 0.0},
		{0.9, 0.1, 0.0},
		{0.0, 0.0, 1.0},
	}

	dist := pairwiseDistances(embeddings)
	merges := wardLinkage(dist, 4)

	if len(merges) != 3 {
		t.Fatalf("expected 3 merges, got %d", len(merges))
	}

	// First merge should be between two of the close points (0,1 or 1,2 — both have d²=0.005)
	m0 := merges[0]
	closePoints := (m0.a == 0 && m0.b == 1) || (m0.a == 1 && m0.b == 0) ||
		(m0.a == 1 && m0.b == 2) || (m0.a == 2 && m0.b == 1) ||
		(m0.a == 0 && m0.b == 2) || (m0.a == 2 && m0.b == 0)
	if !closePoints {
		t.Errorf("expected first merge between close points, got %d and %d", m0.a, m0.b)
	}

	// Distances should be increasing
	for i := 1; i < len(merges); i++ {
		if merges[i].distance < merges[i-1].distance-1e-10 {
			t.Errorf("merge distances should be non-decreasing: %f < %f", merges[i].distance, merges[i-1].distance)
		}
	}
}

func TestCutDendrogramThreshold(t *testing.T) {
	// 4 points: 3 close together + 1 far away
	embeddings := [][]float64{
		{1.0, 0.0, 0.0},
		{0.95, 0.05, 0.0},
		{0.9, 0.1, 0.0},
		{0.0, 0.0, 1.0},
	}

	dist := pairwiseDistances(embeddings)
	merges := wardLinkage(dist, 4)
	labels := cutDendrogram(merges, 4, 1.0)

	// Points 0, 1, 2 should be in the same cluster
	if labels[0] != labels[1] || labels[1] != labels[2] {
		t.Errorf("expected points 0,1,2 in same cluster, got labels %v", labels)
	}

	// Point 3 should be in a different cluster
	if labels[3] == labels[0] {
		t.Errorf("expected point 3 in different cluster, got labels %v", labels)
	}
}

func TestCutDendrogramAllSeparate(t *testing.T) {
	embeddings := [][]float64{
		{1.0, 0.0},
		{0.0, 1.0},
		{-1.0, 0.0},
	}

	dist := pairwiseDistances(embeddings)
	merges := wardLinkage(dist, 3)
	labels := cutDendrogram(merges, 3, 0.001) // very small threshold

	// Each point should be in its own cluster
	if labels[0] == labels[1] || labels[1] == labels[2] || labels[0] == labels[2] {
		t.Errorf("expected all separate clusters with tiny threshold, got labels %v", labels)
	}
}

func TestCutDendrogramAllMerged(t *testing.T) {
	embeddings := [][]float64{
		{1.0, 0.0},
		{0.0, 1.0},
		{-1.0, 0.0},
	}

	dist := pairwiseDistances(embeddings)
	merges := wardLinkage(dist, 3)
	labels := cutDendrogram(merges, 3, 100.0) // very large threshold

	// All points should be in the same cluster
	if labels[0] != labels[1] || labels[1] != labels[2] {
		t.Errorf("expected all in same cluster with large threshold, got labels %v", labels)
	}
}
