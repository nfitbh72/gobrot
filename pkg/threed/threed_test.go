package threed

import (
	"reflect"
	"testing"
)

func testLayer() [][]float64 {
	return [][]float64{
		{0.1, 0.4, 0.9},
		{0.5, 0.2, 0.7},
		{0.3, 0.95, 0.6},
	}
}

func TestGetSolidFromLayerDoesNotMutateLayer(t *testing.T) {
	layer := testLayer()
	m := &TThreeD{Scale: 1, DepthScale: 1, InverseDepth: true, CutBelow: 0.3, CutAbove: 0.8}
	m.GetSolidFromLayer(layer, "a")
	if !reflect.DeepEqual(layer, testLayer()) {
		t.Errorf("layer was modified: %v", layer)
	}
}

func TestGetSolidFromLayerIsRepeatable(t *testing.T) {
	layer := testLayer()
	m := &TThreeD{Scale: 1, DepthScale: 1, InverseDepth: true, CutBelow: 0.3, CutAbove: 0.8}
	first := m.GetSolidFromLayer(layer, "a")
	// relaxing the cuts must see the original data, not the clipped one
	m.CutBelow, m.CutAbove = 0, 1
	relaxed := m.GetSolidFromLayer(layer, "b")
	fresh := (&TThreeD{Scale: 1, DepthScale: 1, InverseDepth: true, CutBelow: 0, CutAbove: 1}).GetSolidFromLayer(testLayer(), "c")
	if len(relaxed.Triangles) != len(fresh.Triangles) {
		t.Errorf("after clipped run: %d triangles, fresh data gives %d", len(relaxed.Triangles), len(fresh.Triangles))
	}
	if len(first.Triangles) == len(fresh.Triangles) {
		t.Errorf("cuts had no effect: %d triangles either way", len(first.Triangles))
	}
}

func TestGetSolidFromLayerIsDeterministicWithZeroAttributes(t *testing.T) {
	m := &TThreeD{Scale: 1, DepthScale: 1, CutAbove: 1}
	a := m.GetSolidFromLayer(testLayer(), "a")
	b := m.GetSolidFromLayer(testLayer(), "a")
	if !reflect.DeepEqual(a.Triangles, b.Triangles) {
		t.Error("identical input gave different triangles")
	}
	for i, tr := range a.Triangles {
		if tr.Attributes != 0 {
			t.Fatalf("triangle %d has attribute %d, want 0", i, tr.Attributes)
		}
	}
}
