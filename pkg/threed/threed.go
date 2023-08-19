package threed

import (
	"stl"
	"strconv"
	"fmt"
)

type TThreeD struct {
	FilenameBase string
}

func (m *TThreeD) Initialise () {

}

//saves a new stl file for each layer
func (m *TThreeD) Save(layers [][][]float64) {
	for n, layer := range layers {
		filename := m.FilenameBase + strconv.Itoa(n) + ".stl"
		name := "out" + strconv.Itoa(n)
		solid := m.GetSolidFromLayer(layer, name)
		fmt.Println("Saving " + filename)
		solid.WriteFile(filename)
	}
}

//given i, j, the width and height of the layer and the depth of the pixel, calculate the vector 3d point
func (m *TThreeD) GetVec(i int, j int, width int, height int, depth float64) stl.Vec3 {
	v := stl.Vec3{
		float32(i) / float32(width),
		float32(j) / float32(height),
		float32(1.0 - depth),
	}
	return v
}

//given three 3d points, create the triangle
func (m *TThreeD) GetTriangle(p1 stl.Vec3, p2 stl.Vec3, p3 stl.Vec3) stl.Triangle {
	return stl.Triangle{
		Vertices: [3]stl.Vec3{p1, p3, p2},
	}
}

//create a solid from the layer values
func (m *TThreeD) GetSolidFromLayer(layer [][]float64, name string) stl.Solid {
	solid := stl.Solid {
		Name: name,
	}
	width := len(layer)
    height := len(layer[0])

	//for all lines
	for i:= 0; i < width; i++ {
		//for all pixels
        for j := 0; j < height; j++ {
			//only if not first row/column
			if (i > 0 && j > 0) {
				//get 4 points (i-1,j-1, i-1,j, i,j-1 and i,j)
				p1 := m.GetVec(i-1, j-1, width, height, layer[i-1][j-1])
				p2 := m.GetVec(i, j-1, width, height, layer[i][j-1])
				p3 := m.GetVec(i-1, j, width, height, layer[i-1][j])
				p4 := m.GetVec(i, j, width, height, layer[i][j])
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
				//append the 2 triangles
				solid.AppendTriangle(t1)
				solid.AppendTriangle(t2)
			}
		}
	}

	//we didnt add any normals to the triangles as we created them, so get the package to do the work
	solid.RecalculateNormals()

	errors := solid.Validate()


	for key, error := range errors {
		fmt.Println(key, "has a problem", error.HasEqualVertices, error.NormalDoesNotMatch, len(error.EdgeErrors))
		if (len(error.EdgeErrors) > 0) {
			for c, ee := range error.EdgeErrors {
				if ee != nil {
					vs := solid.Triangles[key].Vertices
					switch c {
						case 0: fmt.Println("0->1: ", vs[0], "to", vs[1])
						case 1: fmt.Println("1->2: ", vs[1], "to", vs[2])
						case 2: fmt.Println("2->0: ", vs[2], "to", vs[0])
						default:
					}
				}
			}
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
			{0.9, 0.9, 0.9},
			{0.1, 0.7, 0.1},
			{0.5, 0.5, 0.5},
		},
	}
	m.Save(layers)
}