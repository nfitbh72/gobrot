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

// A cell is the square of grid points (i-1,j-1) (i,j-1) (i-1,j) (i,j), numbered 0..3 in that order.
// Each cell is split into two roof triangles along one diagonal, alternating with (i+j) parity, as
// seen from above:
//
//	x-x-x
//	|/|\|
//	x-x-x
//	|\|/|
//	x-x-x
//
// cellTriangles lists each roof triangle's corners in counter-clockwise order (normal pointing up)
// for [parity][triangle].
var cellTriangles = [2][2][3]int{
	{{1, 2, 0}, {2, 1, 3}},
	{{3, 0, 1}, {0, 3, 2}},
}

// cellCorner is the (i, j) offset of corner n from the cell's first corner (i-1, j-1).
var cellCorner = [4][2]int{{0, 0}, {1, 0}, {0, 1}, {1, 1}}

// cellSide identifies which side of a cell an edge between two corners lies on, or sideDiagonal
// for the diagonal shared by the cell's own two triangles.
const (
	sideJMinus = iota // corners 0-1, towards j-1
	sideJPlus         // corners 2-3, towards j+1
	sideIMinus        // corners 0-2, towards i-1
	sideIPlus         // corners 1-3, towards i+1
	sideDiagonal
)

func cellSide(a, b int) int {
	if a > b {
		a, b = b, a
	}
	switch [2]int{a, b} {
	case [2]int{0, 1}:
		return sideJMinus
	case [2]int{2, 3}:
		return sideJPlus
	case [2]int{0, 2}:
		return sideIMinus
	case [2]int{1, 3}:
		return sideIPlus
	}
	return sideDiagonal
}

// oppositeSide is the side of the neighbouring cell that touches a given side.
var oppositeSide = [4]int{sideJPlus, sideJMinus, sideIPlus, sideIMinus}

// sideOwner[parity][side] is which of a cell's two triangles (0 or 1) owns the given outer side.
var sideOwner = func() (owner [2][4]int) {
	for parity := 0; parity < 2; parity++ {
		for t := 0; t < 2; t++ {
			c := cellTriangles[parity][t]
			for k := 0; k < 3; k++ {
				if side := cellSide(c[k], c[(k+1)%3]); side != sideDiagonal {
					owner[parity][side] = t
				}
			}
		}
	}
	return
}()

// create a solid from the layer values.
//
// The solid is a heightmap slab: a roof triangulating the layer, a mirrored base at z = 0 and walls
// joining them. Grid points with a depth of 0 (empty, e.g. cut material with InverseDepth) carry no
// material, so a roof triangle is only solid if all three of its points are non-empty. Points where
// two filled regions would touch at a single point are emptied as well, so the result is a valid solid. Solid
// triangles get a base triangle, and a wall wherever a solid triangle borders an empty one or the
// edge of the layer. The result has no zero-area triangles and every edge is shared by exactly two
// triangles with opposite winding.
func (m *TThreeD) GetSolidFromLayer(layer [][]float64, name string) stl.Solid {
	solid := stl.Solid{
		Name: name,
	}
	//the cuts are applied to a copy: layer is shared with the image pipeline and later runs
	layer = m.applyCuts(layer)
	width := len(layer)
	height := len(layer[0])

	roof := func(i, j int) stl.Vec3 {
		return m.MakeVector(i, j, width, height, layer[i][j], m.InverseDepth)
	}
	floor := func(i, j int) stl.Vec3 {
		return m.MakeVector(i, j, width, height, 0.0, false)
	}
	corner := func(i, j, n int) (int, int) {
		return i - 1 + cellCorner[n][0], j - 1 + cellCorner[n][1]
	}

	//which triangles hold material? Triangles with an empty point do not
	empty := make([][]bool, width)
	filled := make([][][2]bool, width)
	for i := range empty {
		empty[i] = make([]bool, height)
		filled[i] = make([][2]bool, height)
		for j := range empty[i] {
			empty[i][j] = roof(i, j)[2] == 0.0
		}
	}
	refill := func() {
		for i := 1; i < width; i++ {
			for j := 1; j < height; j++ {
				for t := 0; t < 2; t++ {
					filled[i][j][t] = true
					for _, n := range cellTriangles[(i+j)%2][t] {
						ci, cj := corner(i, j, n)
						if empty[ci][cj] {
							filled[i][j][t] = false
						}
					}
				}
			}
		}
	}

	//is the triangle on the far side of an edge filled?
	neighbourFilled := func(i, j, t, side int) bool {
		switch side {
		case sideDiagonal:
			return filled[i][j][1-t]
		case sideJMinus:
			j--
		case sideJPlus:
			j++
		case sideIMinus:
			i--
		case sideIPlus:
			i++
		}
		if i < 1 || i >= width || j < 1 || j >= height {
			return false
		}
		return filled[i][j][sideOwner[(i+j)%2][oppositeSide[side]]]
	}

	//two filled regions touching at a single point would put four walls on one vertical edge, which
	//is not a valid solid. Such a point has more than 2 wall edges meeting at it (a normal point has
	//0 or 2), so empty those points too until none are left.
	refill()
	for {
		walls := make([][]uint8, width)
		for i := range walls {
			walls[i] = make([]uint8, height)
		}
		for i := 1; i < width; i++ {
			for j := 1; j < height; j++ {
				for t := 0; t < 2; t++ {
					if !filled[i][j][t] {
						continue
					}
					c := cellTriangles[(i+j)%2][t]
					for k := 0; k < 3; k++ {
						if !neighbourFilled(i, j, t, cellSide(c[k], c[(k+1)%3])) {
							for _, n := range [2]int{c[k], c[(k+1)%3]} {
								ci, cj := corner(i, j, n)
								walls[ci][cj]++
							}
						}
					}
				}
			}
		}
		changed := false
		for i := range walls {
			for j := range walls[i] {
				if walls[i][j] > 2 {
					empty[i][j] = true
					changed = true
				}
			}
		}
		if !changed {
			break
		}
		refill()
	}

	//pass 2: emit the geometry. add takes corners in counter-clockwise order (MakeTriangle reverses them)
	add := func(a, b, c stl.Vec3) {
		solid.AppendTriangle(m.MakeTriangle(a, c, b))
	}
	numTriangles := 0
	for i := 1; i < width; i++ {
		for j := 1; j < height; j++ {
			parity := (i + j) % 2
			for t := 0; t < 2; t++ {
				if !filled[i][j][t] {
					continue
				}
				c := cellTriangles[parity][t]
				var ci, cj [3]int
				for k := 0; k < 3; k++ {
					ci[k], cj[k] = corner(i, j, c[k])
				}
				add(roof(ci[0], cj[0]), roof(ci[1], cj[1]), roof(ci[2], cj[2]))
				add(floor(ci[0], cj[0]), floor(ci[2], cj[2]), floor(ci[1], cj[1]))
				numTriangles++
				//walls on every edge with no filled triangle behind it
				for k := 0; k < 3; k++ {
					x, y := k, (k+1)%3
					if neighbourFilled(i, j, t, cellSide(c[x], c[y])) {
						continue
					}
					add(roof(ci[y], cj[y]), roof(ci[x], cj[x]), floor(ci[x], cj[x]))
					add(roof(ci[y], cj[y]), floor(ci[x], cj[x]), floor(ci[y], cj[y]))
				}
			}
		}
	}
	fmt.Println("Created", numTriangles, " roof triangles")

	//we didnt add any normals to the triangles as we created them, so get the package to do the work
	solid.RecalculateNormals()

	fmt.Println("Added normals")

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
