package threed

import (
	"stl"
	"strconv"
	"fmt"
	"math/rand"
)

type TThreeD struct {
	FilenameBase string
	InverseDepth bool
}

func (m *TThreeD) Initialise () {
	m.InverseDepth = false
}

//saves a new stl file for each layer
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

//given i, j, the width and height of the layer and the depth of the pixel, calculate the vector 3d point
func (m *TThreeD) GetVec(i int, j int, width int, height int, depth float64, inverseDepth bool) stl.Vec3 {
	if inverseDepth {
		depth = 1.0 - depth
	}
	v := stl.Vec3{
		float32(i) / float32(width-1),
		float32(j) / float32(height-1),
		float32(depth),
	}
	return v
}

//given three 3d points, create the triangle
func (m *TThreeD) GetTriangle(p1 stl.Vec3, p2 stl.Vec3, p3 stl.Vec3) stl.Triangle {
	return stl.Triangle{
		Vertices: [3]stl.Vec3{p1, p3, p2},
		Attributes: uint16(rand.Intn(65535)),
	}
}

//create a solid from the layer values
func (m *TThreeD) GetSolidFromLayer(layer [][]float64, name string) stl.Solid {
	solid := stl.Solid {
		Name: name,
	}
	width := len(layer)
    height := len(layer[0])

	//make a base of 2 triangles that covers the entire of the bottom
	p1 := stl.Vec3{0, 0, 0}
	p2 := stl.Vec3{0, 1, 0}
	p3 := stl.Vec3{1, 0, 0}
	p4 := stl.Vec3{1, 1, 0}
	//solid.AppendTriangle(m.GetTriangle(p2, p1, p3))
	//solid.AppendTriangle(m.GetTriangle(p2, p4, p3))

	//for all lines
	for i:= 0; i < width; i++ {
		firstLine := (i == 0)
		lastLine := (i == (width-1))
		//for all pixels
        for j := 0; j < height; j++ {
			firstPixel := (j == 0)
			lastPixel := (j == (height-1))

			//first/last lines need special treatment to tie them into the bottom of the solid
			if !firstPixel && (firstLine || lastLine) {
				p1 = m.GetVec(i, j-1, width, height, layer[i][j-1], m.InverseDepth)
				p2 = m.GetVec(i, j, width, height, layer[i][j], m.InverseDepth)
				p3 = m.GetVec(i, j-1, width, height, 0.0, false)
				p4 = m.GetVec(i, j, width, height, 0.0, false)
				if (firstLine) {
					solid.AppendTriangle(m.GetTriangle(p1, p3, p2))
					solid.AppendTriangle(m.GetTriangle(p3, p4, p2))
				} else {
					solid.AppendTriangle(m.GetTriangle(p3, p1, p2))
					solid.AppendTriangle(m.GetTriangle(p4, p3, p2))
				}
			}
			//first/last pixels need special treatment to tie them into the bottom of the solid
			if !firstLine && (firstPixel || lastPixel) {
				p1 = m.GetVec(i-1, j, width, height, layer[i-1][j], m.InverseDepth)
				p2 = m.GetVec(i, j, width, height, layer[i][j], m.InverseDepth)
				p3 = m.GetVec(i-1, j, width, height, 0.0, false)
				p4 = m.GetVec(i, j, width, height, 0.0, false)
				if (firstPixel) {
					solid.AppendTriangle(m.GetTriangle(p3, p1, p2))
					solid.AppendTriangle(m.GetTriangle(p4, p3, p2))
				} else {
					solid.AppendTriangle(m.GetTriangle(p1, p3, p2))
					solid.AppendTriangle(m.GetTriangle(p3, p4, p2))
					
				}
			}
			//only if not first row/column
			if !firstLine && !firstPixel {
				//these are the "roof" of our solid
				//get 4 points (i-1,j-1, i-1,j, i,j-1 and i,j)
				p1 = m.GetVec(i-1, j-1, width, height, layer[i-1][j-1], m.InverseDepth)
				p2 = m.GetVec(i, j-1, width, height, layer[i][j-1], m.InverseDepth)
				p3 = m.GetVec(i-1, j, width, height, layer[i-1][j], m.InverseDepth)
				p4 = m.GetVec(i, j, width, height, layer[i][j], m.InverseDepth)

				var t1 stl.Triangle
				var t2 stl.Triangle
				//triangle composition alternates between diagonals... as seen from above:
				// x-x-x
				// |/|\|
				// x-x-x
				// |\|/|
				// x-x-x
				if ((i + j) % 2 == 0) {
					t1 = m.GetTriangle(p2, p1, p3)
					t2 = m.GetTriangle(p3, p4, p2)
				} else {
					t1 = m.GetTriangle(p4, p2, p1)
					t2 = m.GetTriangle(p1, p3, p4)
				}
				drawT1 := false
				drawT2 := false
				//append the 2 triangles
				if p1[2] != 0.0 || p2[2] != 0.0 || p3[2] != 0.0 {
					drawT1 = true
					solid.AppendTriangle(t1)
				}
				if p4[2] != 0.0 || p2[2] != 0.0 || p3[2] != 0.0 {
					drawT2 = true
					solid.AppendTriangle(t2)
				}

				//these are the base of the solid
				p1 = m.GetVec(i-1, j-1, width, height, 0.0, false)
				p2 = m.GetVec(i, j-1, width, height, 0.0, false)
				p3 = m.GetVec(i-1, j, width, height, 0.0, false)
				p4 = m.GetVec(i, j, width, height, 0.0, false)
				
				if ((i + j) % 2 == 0) {
					t1 = m.GetTriangle(p1, p2, p3)
					t2 = m.GetTriangle(p4, p3, p2)
				} else {
					t1 = m.GetTriangle(p2, p4, p1)
					t2 = m.GetTriangle(p3, p1, p4)
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

	//we didnt add any normals to the triangles as we created them, so get the package to do the work
	solid.RecalculateNormals()

	errors := solid.Validate()


	for key, error := range errors {
		fmt.Println(key, "has a problem", error.HasEqualVertices, error.NormalDoesNotMatch, len(error.EdgeErrors))
		if (len(error.EdgeErrors) > 0) {
			vs := solid.Triangles[key].Vertices
			fmt.Println(error.EdgeErrors[0], "0->1: ", vs[0], "to", vs[1])
			fmt.Println(error.EdgeErrors[1], "1->2: ", vs[1], "to", vs[2])
			fmt.Println(error.EdgeErrors[2], "2->0: ", vs[2], "to", vs[0])
			/*
			for c, ee := range error.EdgeErrors {
				if ee != nil {
					fmt.Println(ee)
					
					vs := solid.Triangles[key].Vertices
					switch c {
						case 0: fmt.Println("0->1: ", vs[0], "to", vs[1])
						case 1: fmt.Println("1->2: ", vs[1], "to", vs[2])
						case 2: fmt.Println("2->0: ", vs[2], "to", vs[0])
						default:
					}
					
				}
			}
			*/
			//fmt.Println(error.EdgeErrors)
			//fmt.Println(solid.Triangles[key])
		}
	}

	fmt.Println("Solid has", len(errors), "validation errors")

	//return the solid
	return solid
}

func (m *TThreeD) Test() {
	layers := [][][]float64{
		{
			{0.1, 0.9, 0.1},
			{0.9, 0.8, 0.9},
			{0.1, 0.7, 0.1},
			{0.5, 0.5, 0.5},
		},
	}
	m.Save(layers)
}