package threed

import (
	"fmt"
	"github.com/hschendel/stl"
	"strconv"
)

type TThreeD struct {
	FilenameBase string
	InverseDepth bool
	Scale        float64
	DepthScale   float64
	CutBelow     float64
	CutAbove     float64
	// Tolerance is the furthest, in output units, the surface may move from the layer's heights when
	// flat and gently sloping areas are covered with fewer triangles. 0 keeps every height exactly.
	Tolerance float64
}

func (m *TThreeD) Initialise() {
	m.InverseDepth = false
	m.DepthScale = 1.0
	m.CutBelow = 0.0
	m.CutAbove = 1.0
	m.Scale = 1.0
}

// saves a new stl file for each layer
func (m *TThreeD) Save(layers [][][]float64) {
	for n, layer := range layers {
		filename := m.FilenameBase + strconv.Itoa(n) + ".stl"
		name := "out" + strconv.Itoa(n)
		if err := checkSize(layer); err != nil {
			fmt.Println("Not saving "+filename+":", err)
			continue
		}
		solid := m.GetSolidFromLayer(layer, name)
		fmt.Println("Saving " + filename)
		err := solid.WriteFile(filename)
		if err != nil {
			fmt.Println(err)
		}
	}
}

// applyCuts returns a copy of the layer with values at or below CutBelow, or at or above CutAbove,
// set to 1.0. 1.0 is the "no material" value of the layer pipeline (see layers.GetLayers), so both
// cuts deliberately remove material in the same way.
func (m *TThreeD) applyCuts(layer [][]float64) [][]float64 {
	out := make([][]float64, len(layer))
	for i := range layer {
		out[i] = make([]float64, len(layer[i]))
		for j, v := range layer[i] {
			if v <= m.CutBelow || v >= m.CutAbove {
				v = 1.0
			}
			out[i][j] = v
		}
	}
	return out
}

// heights converts a layer to model heights, after the cuts. A height of 0 means no material.
func (m *TThreeD) heights(layer [][]float64) [][]float32 {
	//the cuts are applied to a copy: layer is shared with the image pipeline and later runs
	layer = m.applyCuts(layer)
	out := make([][]float32, len(layer))
	for i := range layer {
		out[i] = make([]float32, len(layer[i]))
		for j, depth := range layer[i] {
			if m.InverseDepth {
				depth = 1.0 - depth
			}
			out[i][j] = float32(m.Scale) * float32(depth) * float32(m.DepthScale)
		}
	}
	return out
}

// checkSize reports whether a layer is small enough to turn into a solid. The mesher works on a
// square grid as wide as the layer's longer side rounded up to a power of two, so the limit is on
// the longer side.
func checkSize(layer [][]float64) error {
	w, h := len(layer), 0
	if w > 0 {
		h = len(layer[0])
	}
	if gridSize(w, h) > maxGridSize {
		return fmt.Errorf("layer of %d x %d points is too large for 3d output, which takes up to %d points a side", w, h, maxGridSize+1)
	}
	return nil
}

// create a solid from the layer values.
//
// The solid is a heightmap slab: a roof following the layer, a flat base at z = 0 and walls joining
// them. Grid points with a height of 0 (empty, e.g. cut material with InverseDepth) carry no
// material and are left out, with walls around the hole. Flat and gently sloping areas are covered
// by fewer, larger triangles, as far as Tolerance allows. See mesh.go and docs/stl-mesh.md.
//
// The result has no zero-area triangles and every edge is shared by exactly two triangles with
// opposite winding. A layer that checkSize refuses gives a solid with no triangles.
func (m *TThreeD) GetSolidFromLayer(layer [][]float64, name string) stl.Solid {
	solid := stl.Solid{
		Name: name,
	}
	if checkSize(layer) != nil {
		return solid
	}
	mesh := newMesher(m.heights(layer))
	mesh.build(float32(m.Tolerance))

	//the layer is stretched to Scale in both directions, whatever its shape
	position := func(v vertex) stl.Vec3 {
		x, y, z := mesh.point(v)
		return stl.Vec3{
			float32(m.Scale) * float32(x) / float32(mesh.w-1),
			float32(m.Scale) * float32(y) / float32(mesh.h-1),
			z,
		}
	}
	solid.Triangles = make([]stl.Triangle, len(mesh.triangles))
	for n, t := range mesh.triangles {
		solid.Triangles[n].Vertices = [3]stl.Vec3{position(t[0]), position(t[1]), position(t[2])}
	}
	fmt.Println("Created", len(solid.Triangles), "triangles")

	//we didnt add any normals to the triangles as we created them, so get the package to do the work
	solid.RecalculateNormals()

	//return the solid
	return solid
}

// allows a test on a known and simple structure
func (m *TThreeD) Test() {
	layers := [][][]float64{
		{
			{0.4, 0.4, 0.1, 0.1},
			{0.9, 0.7, 1.0, 1.0},
			{0.1, 1.0, 1.0, 1.0},
			{0.5, 0.5, 0.5, 0.5},
		},
	}
	m.Save(layers)
}
