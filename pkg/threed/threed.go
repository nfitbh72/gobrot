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
		solid := m.GetSolidFromLayer(layer, name)
		fmt.Println("Saving " + filename)
		err := solid.WriteFile(filename)
		if err != nil {
			fmt.Println(err)
		}
	}
}

// given i, j, the width and height of the layer and the depth of the pixel, calculate the vector 3d point
func (m *TThreeD) MakeVector(i int, j int, width int, height int, depth float64, inverseDepth bool) stl.Vec3 {
	if inverseDepth {
		depth = 1.0 - depth
	}
	v := stl.Vec3{
		float32(m.Scale) * float32(i) / float32(width-1),
		float32(m.Scale) * float32(j) / float32(height-1),
		float32(m.Scale) * float32(depth) * float32(m.DepthScale),
	}
	return v
}

// given three 3d points, create the triangle
func (m *TThreeD) MakeTriangle(p1 stl.Vec3, p2 stl.Vec3, p3 stl.Vec3) stl.Triangle {
	return stl.Triangle{
		Vertices:   [3]stl.Vec3{p1, p3, p2},
		Attributes: 0,
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

// create a solid from the layer values
func (m *TThreeD) GetSolidFromLayer(layer [][]float64, name string) stl.Solid {
	solid := stl.Solid{
		Name: name,
	}
	//the cuts are applied to a copy: layer is shared with the image pipeline and later runs
	layer = m.applyCuts(layer)
	width := len(layer)
	height := len(layer[0])

	//4 vectors
	p1 := stl.Vec3{0, 0, 0}
	p2 := stl.Vec3{0, 0, 0}
	p3 := stl.Vec3{0, 0, 0}
	p4 := stl.Vec3{0, 0, 0}

	numTriangles := 0

	//for all lines
	for i := 0; i < width; i++ {
		firstLine := (i == 0)
		lastLine := (i == (width - 1))
		//for all pixels
		for j := 0; j < height; j++ {
			firstPixel := (j == 0)
			lastPixel := (j == (height - 1))
			//fmt.Printf("pixel: %0.4f; cutabove: %0.4f\n", layer[i][j], m.CutAbove)

			//first/last lines need special treatment to tie them into the bottom of the solid
			if !firstPixel && (firstLine || lastLine) {
				p1 = m.MakeVector(i, j-1, width, height, layer[i][j-1], m.InverseDepth)
				p2 = m.MakeVector(i, j, width, height, layer[i][j], m.InverseDepth)
				p3 = m.MakeVector(i, j-1, width, height, 0.0, false)
				p4 = m.MakeVector(i, j, width, height, 0.0, false)
				if firstLine {
					solid.AppendTriangle(m.MakeTriangle(p1, p3, p2))
					solid.AppendTriangle(m.MakeTriangle(p3, p4, p2))
				} else {
					solid.AppendTriangle(m.MakeTriangle(p3, p1, p2))
					solid.AppendTriangle(m.MakeTriangle(p4, p3, p2))
				}
			}
			//first/last pixels need special treatment to tie them into the bottom of the solid
			if !firstLine && (firstPixel || lastPixel) {
				p1 = m.MakeVector(i-1, j, width, height, layer[i-1][j], m.InverseDepth)
				p2 = m.MakeVector(i, j, width, height, layer[i][j], m.InverseDepth)
				p3 = m.MakeVector(i-1, j, width, height, 0.0, false)
				p4 = m.MakeVector(i, j, width, height, 0.0, false)
				if firstPixel {
					solid.AppendTriangle(m.MakeTriangle(p3, p1, p2))
					solid.AppendTriangle(m.MakeTriangle(p4, p3, p2))
				} else {
					solid.AppendTriangle(m.MakeTriangle(p1, p3, p2))
					solid.AppendTriangle(m.MakeTriangle(p3, p4, p2))

				}
			}
			//only if not first row/column
			if !firstLine && !firstPixel {
				//these are the "roof" of our solid
				//get 4 points (i-1,j-1, i-1,j, i,j-1 and i,j)
				p1 = m.MakeVector(i-1, j-1, width, height, layer[i-1][j-1], m.InverseDepth)
				p2 = m.MakeVector(i, j-1, width, height, layer[i][j-1], m.InverseDepth)
				p3 = m.MakeVector(i-1, j, width, height, layer[i-1][j], m.InverseDepth)
				p4 = m.MakeVector(i, j, width, height, layer[i][j], m.InverseDepth)

				var t1 stl.Triangle
				var t2 stl.Triangle
				//triangle composition alternates between diagonals... as seen from above:
				// x-x-x
				// |/|\|
				// x-x-x
				// |\|/|
				// x-x-x
				if (i+j)%2 == 0 {
					t1 = m.MakeTriangle(p2, p1, p3)
					t2 = m.MakeTriangle(p3, p4, p2)
				} else {
					t1 = m.MakeTriangle(p4, p2, p1)
					t2 = m.MakeTriangle(p1, p3, p4)
				}
				drawT1 := false
				drawT2 := false

				//append the 2 triangles
				if p1[2] != 0.0 || p2[2] != 0.0 || p3[2] != 0.0 {
					drawT1 = true
					solid.AppendTriangle(t1)
					numTriangles++
				}
				if p4[2] != 0.0 || p2[2] != 0.0 || p3[2] != 0.0 {
					drawT2 = true
					solid.AppendTriangle(t2)
					numTriangles++
				}

				//these are the base of the solid - same coords as the roof triangles except with their depth set to 0
				p1 = m.MakeVector(i-1, j-1, width, height, 0.0, false)
				p2 = m.MakeVector(i, j-1, width, height, 0.0, false)
				p3 = m.MakeVector(i-1, j, width, height, 0.0, false)
				p4 = m.MakeVector(i, j, width, height, 0.0, false)

				if (i+j)%2 == 0 {
					t1 = m.MakeTriangle(p1, p2, p3)
					t2 = m.MakeTriangle(p4, p3, p2)
				} else {
					t1 = m.MakeTriangle(p2, p4, p1)
					t2 = m.MakeTriangle(p3, p1, p4)
				}
				//append the 2 triangles
				if drawT1 {
					solid.AppendTriangle(t1)
				}
				if drawT2 {
					solid.AppendTriangle(t2)
				}
			}
		}
	}
	fmt.Println("Created", numTriangles, " triangles")

	//we didnt add any normals to the triangles as we created them, so get the package to do the work
	solid.RecalculateNormals()

	fmt.Println("Added normals")

	//validate() takes 30 secs and doesn't fix anything - only worth validating if we're gonna do something about it
	//errors := solid.Validate()
	/*
		for key, error := range errors {
			fmt.Println(key, "has a problem", error.HasEqualVertices, error.NormalDoesNotMatch, len(error.EdgeErrors))
			if (len(error.EdgeErrors) > 0) {
				vs := solid.Triangles[key].Vertices
				fmt.Println(error.EdgeErrors[0], "0->1: ", vs[0], "to", vs[1])
				fmt.Println(error.EdgeErrors[1], "1->2: ", vs[1], "to", vs[2])
				fmt.Println(error.EdgeErrors[2], "2->0: ", vs[2], "to", vs[0])
			}
		}
	*/

	//fmt.Println("Solid has", len(errors), "validation errors")

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
