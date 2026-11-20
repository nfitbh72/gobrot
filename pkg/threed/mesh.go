package threed

import "math"

// The mesher turns a grid of heights into a closed solid: a roof following the heights, a flat
// base at z = 0 and vertical walls wherever material meets empty space. docs/stl-mesh.md has the
// design; the short version is below.
//
// Grid points live on a square grid of size+1 points per side, size a power of two, with the layer
// in the corner at the origin (points beyond the layer are empty). The grid is triangulated by
// right-angled triangles that halve recursively: a triangle is split from the midpoint of its long
// side to the opposite corner, giving two triangles of the same shape. The smallest triangles are
// half a grid cell each.
//
// Every vertex position is computed from its grid point alone, so two triangles that share a point
// share it exactly, and the triangulation never has a point part-way along a neighbour's edge. The
// mesh is closed because of how it is built, not because of a repair pass.

// vertex identifies a mesh vertex: a grid point index, doubled, plus 1 for the roof or 0 for the base.
type vertex = uint32

type mesher struct {
	size   int // cells per side of the square grid, a power of two
	stride int // points per side: size + 1
	w, h   int // points per side of the layer itself

	z     []float32 // height of each point, 0 for points with no material
	solid []bool    // whether the point carries material
	// err is, for a point that is the long-side midpoint of a triangle, how far the surface could
	// move if that triangle were not split there. +Inf where it must be split regardless.
	err []float32
	// force marks the points the base has to keep: where material meets empty space, and the wall
	// corners the roof ended up with.
	force []bool

	triangles [][3]vertex
}

// newMesher builds the grid for heights[i][j], where a height of 0 means no material.
func newMesher(heights [][]float32) *mesher {
	m := &mesher{w: len(heights)}
	if m.w > 0 {
		m.h = len(heights[0])
	}
	m.size = 1
	for m.size < m.w-1 || m.size < m.h-1 {
		m.size *= 2
	}
	m.stride = m.size + 1
	n := m.stride * m.stride
	m.z = make([]float32, n)
	m.solid = make([]bool, n)
	m.err = make([]float32, n)
	m.force = make([]bool, n)
	for i := range heights {
		for j, z := range heights[i] {
			m.z[m.index(i, j)] = z
			m.solid[m.index(i, j)] = z != 0
		}
	}
	return m
}

func (m *mesher) index(x, y int) int {
	return y*m.stride + x
}

// build makes the whole solid. No point of the roof ends up further than tolerance, measured
// vertically, from where the full-resolution roof would have put it.
func (m *mesher) build(tolerance float32) {
	if m.w < 2 || m.h < 2 {
		return
	}
	m.removePinches()

	//roof: measure what each split is worth, then keep only the splits that matter
	m.eachRoot(func(ax, ay, bx, by, cx, cy int) {
		m.classify(ax, ay, bx, by, cx, cy)
	})
	for i, e := range m.err {
		m.force[i] = math.IsInf(float64(e), 1)
	}
	m.propagate()
	m.eachRoot(func(ax, ay, bx, by, cx, cy int) {
		m.emit(ax, ay, bx, by, cx, cy, tolerance, true)
	})

	//base: flat, so the only splits it needs are the ones that keep its outline equal to the
	//bottom of the walls. The roof marked those points in force as it added the walls.
	for i, f := range m.force {
		m.err[i] = 0
		if f {
			m.err[i] = float32(math.Inf(1))
		}
	}
	m.propagate()
	m.eachRoot(func(ax, ay, bx, by, cx, cy int) {
		m.emit(ax, ay, bx, by, cx, cy, 0, false)
	})
}

// eachRoot calls f for the two triangles that make up the whole grid. A triangle is always given
// as the two ends of its long side then its right-angled corner, anticlockwise seen from above.
func (m *mesher) eachRoot(f func(ax, ay, bx, by, cx, cy int)) {
	s := m.size
	f(s, s, 0, 0, s, 0)
	f(0, 0, s, s, 0, s)
}

// isSmallest reports whether a triangle with long side a-b is half a grid cell, and so cannot be split.
func isSmallest(ax, bx int) bool {
	return ax-bx == 1 || bx-ax == 1
}

// outside reports whether a triangle lies wholly beyond the layer.
func (m *mesher) outside(ax, ay, bx, by, cx, cy int) bool {
	return min(ax, bx, cx) >= m.w-1 || min(ay, by, cy) >= m.h-1
}

// filledAt reports whether the smallest triangle containing a position holds material: it does if
// all three of its corners do. Positions are in quarters of a grid cell and must not lie on a
// triangle edge.
//
// Cells are cut in half along alternating diagonals, which is where the recursive halving ends up:
//
//	x-x-x
//	|\|/|
//	x-x-x
//	|/|\|
//	x-x-x
func (m *mesher) filledAt(qx, qy int) bool {
	if qx < 0 || qy < 0 {
		return false
	}
	x, y := qx>>2, qy>>2
	if x >= m.w-1 || y >= m.h-1 {
		return false
	}
	u, v := qx&3, qy&3
	i := m.index(x, y)
	if (x+y)%2 == 0 {
		//diagonal from (x, y) to (x+1, y+1), both shared by the cell's two triangles
		if !m.solid[i] || !m.solid[i+m.stride+1] {
			return false
		}
		if u > v {
			return m.solid[i+1]
		}
		return m.solid[i+m.stride]
	}
	//diagonal from (x+1, y) to (x, y+1)
	if !m.solid[i+1] || !m.solid[i+m.stride] {
		return false
	}
	if u+v < 4 {
		return m.solid[i]
	}
	return m.solid[i+m.stride+1]
}

func sign(n int) int {
	switch {
	case n < 0:
		return -1
	case n > 0:
		return 1
	}
	return 0
}

// filledBeside reports whether there is material just to one side of the edge from p to q, next
// to p. Everything along one side of a triangle's edge is the same, so next to p speaks for the
// whole edge.
func (m *mesher) filledBeside(px, py, qx, qy int, left bool) bool {
	dx, dy := sign(qx-px), sign(qy-py)
	//half a step along the edge, then a quarter step sideways: (dy, -dx) points right
	nx, ny := dy, -dx
	if left {
		nx, ny = -nx, -ny
	}
	return m.filledAt(4*px+2*dx+nx, 4*py+2*dy+ny)
}

// smallestTriangles calls f with the corners, anticlockwise, of both halves of every cell of the layer.
func (m *mesher) smallestTriangles(f func(x [3]int, y [3]int)) {
	for cy := 0; cy < m.h-1; cy++ {
		for cx := 0; cx < m.w-1; cx++ {
			if (cx+cy)%2 == 0 {
				f([3]int{cx, cx + 1, cx + 1}, [3]int{cy, cy, cy + 1})
				f([3]int{cx, cx + 1, cx}, [3]int{cy, cy + 1, cy + 1})
			} else {
				f([3]int{cx, cx + 1, cx}, [3]int{cy, cy, cy + 1})
				f([3]int{cx + 1, cx + 1, cx}, [3]int{cy, cy + 1, cy + 1})
			}
		}
	}
}

// removePinches empties the points where two regions of material touch at a single point. Left in,
// such a point would put four walls on one vertical edge, which is not a valid solid. A point on
// the outline of a region normally has 2 outline edges meeting at it; a pinch has more.
// Emptying a point can expose another, so repeat until there are none.
func (m *mesher) removePinches() {
	outline := make([]uint8, len(m.solid))
	for {
		clear(outline)
		m.smallestTriangles(func(x [3]int, y [3]int) {
			if !m.solid[m.index(x[0], y[0])] || !m.solid[m.index(x[1], y[1])] || !m.solid[m.index(x[2], y[2])] {
				return
			}
			for k := 0; k < 3; k++ {
				n := (k + 1) % 3
				if !m.filledBeside(x[k], y[k], x[n], y[n], false) {
					outline[m.index(x[k], y[k])]++
					outline[m.index(x[n], y[n])]++
				}
			}
		})
		changed := false
		for i, n := range outline {
			if n > 2 {
				m.solid[i] = false
				m.z[i] = 0
				changed = true
			}
		}
		if !changed {
			return
		}
	}
}

const (
	stateEmpty uint8 = iota
	stateFilled
	stateMixed
)

// classify works out whether a triangle is all material, all empty or a mix, and records in err
// what leaving each triangle unsplit would cost on its own: nothing where it is empty, the height
// its midpoint would move where it is material, and +Inf where it is a mix, because the outline
// of the material has to be kept exactly.
func (m *mesher) classify(ax, ay, bx, by, cx, cy int) uint8 {
	if m.outside(ax, ay, bx, by, cx, cy) {
		return stateEmpty
	}
	a, b, c := m.index(ax, ay), m.index(bx, by), m.index(cx, cy)
	if isSmallest(ax, bx) {
		if m.solid[a] && m.solid[b] && m.solid[c] {
			return stateFilled
		}
		return stateEmpty
	}
	mx, my := (ax+bx)/2, (ay+by)/2
	mid := m.index(mx, my)
	first := m.classify(cx, cy, ax, ay, mx, my)
	second := m.classify(bx, by, cx, cy, mx, my)
	if first != second || first == stateMixed {
		m.err[mid] = float32(math.Inf(1))
		return stateMixed
	}
	if first == stateFilled {
		move := m.z[mid] - (m.z[a]+m.z[b])/2
		if move < 0 {
			move = -move
		}
		if move > m.err[mid] {
			m.err[mid] = move
		}
	}
	return first
}

// propagate turns the per-triangle costs in err into totals. Leaving a triangle unsplit also
// leaves out every split inside it, so its cost is its own plus the worst of the splits below it.
// That total is a true upper bound on how far any point inside can move. It also makes a split
// cost at least as much as any split that relies on it, so the kept splits always fit together
// with no point part-way along a neighbour's edge.
//
// The splits below a point are at fixed offsets from it, so this runs over the points directly,
// smallest triangles first.
func (m *mesher) propagate() {
	worst := func(x, y int, offsets [4][2]int) float32 {
		var e float32
		for _, o := range offsets {
			px, py := x+o[0], y+o[1]
			if px >= 0 && py >= 0 && px <= m.size && py <= m.size {
				e = max(e, m.err[m.index(px, py)])
			}
		}
		return e
	}
	for s := 2; s <= m.size; s *= 2 {
		half := s / 2
		if quarter := half / 2; quarter > 0 {
			//midpoints of the sides of each block of s cells: the splits below are the centres of
			//the four half-size blocks around the point
			diagonal := [4][2]int{{-quarter, -quarter}, {quarter, -quarter}, {-quarter, quarter}, {quarter, quarter}}
			for y := 0; y <= m.size; y += half {
				//rows on a block boundary hold the midpoints of horizontal sides, the rows between
				//hold those of vertical sides
				first := half
				if y%s != 0 {
					first = 0
				}
				for x := first; x <= m.size; x += s {
					m.err[m.index(x, y)] += worst(x, y, diagonal)
				}
			}
		}
		//centres of each block of s cells: the splits below are the midpoints of its four sides
		straight := [4][2]int{{-half, 0}, {half, 0}, {0, -half}, {0, half}}
		for y := half; y < m.size; y += s {
			for x := half; x < m.size; x += s {
				m.err[m.index(x, y)] += worst(x, y, straight)
			}
		}
	}
}

// emit adds the triangles for one triangle of the grid, splitting it for as long as leaving it
// whole would cost more than the tolerance. For the roof it also adds the walls.
func (m *mesher) emit(ax, ay, bx, by, cx, cy int, tolerance float32, roof bool) {
	if m.outside(ax, ay, bx, by, cx, cy) {
		return
	}
	if !isSmallest(ax, bx) {
		mx, my := (ax+bx)/2, (ay+by)/2
		if m.err[m.index(mx, my)] > tolerance {
			m.emit(cx, cy, ax, ay, mx, my, tolerance, roof)
			m.emit(bx, by, cx, cy, mx, my, tolerance, roof)
			return
		}
	}
	//not split, so all material or all empty: look just inside the long side to see which
	if !m.filledBeside(ax, ay, bx, by, true) {
		return
	}
	x := [3]int{ax, bx, cx}
	y := [3]int{ay, by, cy}
	var top, bottom [3]vertex
	for k := 0; k < 3; k++ {
		bottom[k] = vertex(m.index(x[k], y[k])) << 1
		top[k] = bottom[k] | 1
	}
	if !roof {
		//seen from below, so the other way round
		m.triangles = append(m.triangles, [3]vertex{bottom[0], bottom[2], bottom[1]})
		return
	}
	m.triangles = append(m.triangles, top)
	//a wall on every edge with no material on the far side
	for k := 0; k < 3; k++ {
		n := (k + 1) % 3
		if m.filledBeside(x[k], y[k], x[n], y[n], false) {
			continue
		}
		m.triangles = append(m.triangles,
			[3]vertex{top[n], top[k], bottom[k]},
			[3]vertex{top[n], bottom[k], bottom[n]},
		)
		m.force[m.index(x[k], y[k])] = true
		m.force[m.index(x[n], y[n])] = true
	}
}

// point returns the grid point and height of a vertex.
func (m *mesher) point(v vertex) (x, y int, z float32) {
	i := int(v >> 1)
	if v&1 == 1 {
		z = m.z[i]
	}
	return i % m.stride, i / m.stride, z
}
