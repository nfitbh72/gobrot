package threed

import (
	"stl"
	"strconv"
	"fmt"
	//"os/exec"
)

type TThreeD struct {
	FilenameBase string
	//Solid stl.Solid
}

func (m *TThreeD) Initialise () {

}

func (m *TThreeD) Save(layers [][][]float64) {
	for n, layer := range layers {
		filename := m.FilenameBase + strconv.Itoa(n) + ".stl"
		name := "out" + strconv.Itoa(n)
		solid := m.GetSolidFromLayer(layer, name)
		fmt.Println("Saving " + filename)
		solid.WriteFile(filename)
		/*
		if (n == 0) {
			//for now, just lauch the app, would like this to spawn rather than block
			cmd := exec.Command(`mspaint`, filename[2:len(filename)] + " /ForceBootstrapPaint3D")
			if err := cmd.Run(); err != nil {
				fmt.Println("Error:", err)
			}
		}
		*/
	}
}

func (m *TThreeD) GetVec(i int, j int, width int, height int, depth float64) stl.Vec3 {
	v := stl.Vec3{
		float32(i) / float32(width),
		float32(j) / float32(height),
		float32(1.0 - depth),
	}
	return v
}

func (m *TThreeD) GetTriangle(p1 stl.Vec3, p2 stl.Vec3, p3 stl.Vec3) stl.Triangle {
	return stl.Triangle{
		Vertices: [3]stl.Vec3{p1, p3, p2},
	}
}

func (m *TThreeD) GetSolidFromLayer(layer [][]float64, name string) stl.Solid {
	solid := stl.Solid {
		Name: name,
	}
	width := len(layer)
    height := len(layer[0])

	for i:= 0; i < width; i++ {
        for j := 0; j < height; j++ {
			if (i > 0 && j > 0) {
				p1 := m.GetVec(i-1, j-1, width, height, layer[i-1][j-1])
				p2 := m.GetVec(i, j-1, width, height, layer[i][j-1])
				p3 := m.GetVec(i-1, j, width, height, layer[i-1][j])
				p4 := m.GetVec(i, j, width, height, layer[i][j])
				var t1 stl.Triangle
				var t2 stl.Triangle
				if ((i + j) % 2 == 0) {
					t1 = m.GetTriangle(p1, p3, p2)
					t2 = m.GetTriangle(p3, p4, p2)
				} else {
					t1 = m.GetTriangle(p1, p3, p4)
					t2 = m.GetTriangle(p1, p4, p2)
				}
				solid.AppendTriangle(t1)
				solid.AppendTriangle(t2)
			}
		}
	}

	solid.RecalculateNormals()

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