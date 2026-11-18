package images

import (
	"os"
	"testing"
)

type pixel struct {
	depth, shadowDepth float64
	isShadow           bool
}

// recordingProfile records every SetPixel call, keyed by image coordinate.
type recordingProfile struct {
	width, height int
	calls         map[[2]int]pixel
	count         int
}

func (r *recordingProfile) Initialise(filename string, width int, height int) {
	r.width, r.height = width, height
	r.calls = map[[2]int]pixel{}
}

func (r *recordingProfile) SetPixel(i int, j int, depth float64, isShadow bool, shadowDepth float64) {
	r.calls[[2]int{i, j}] = pixel{depth, shadowDepth, isShadow}
	r.count++
}

func (r *recordingProfile) SaveImage(file *os.File) {}

// 3 wide, 2 high; layer[x][y]
func testLayers() (layer, shadow [][]float64) {
	layer = [][]float64{{0.1, 0.2}, {0.3, 0.4}, {0.5, 0.6}}
	shadow = [][]float64{{0.9, 0.8}, {0.7, 0.6}, {0.5, 0.4}}
	return
}

func render(isShadow bool) *recordingProfile {
	layer, shadow := testLayers()
	rec := &recordingProfile{}
	m := &TImage{ColourProfile: rec}
	m.GetImageFromLayer(layer, shadow, isShadow, 0, 1, false)
	return rec
}

func TestEveryPixelWrittenInBounds(t *testing.T) {
	rec := render(false)
	if rec.count != 6 || len(rec.calls) != 6 {
		t.Fatalf("want 6 distinct writes, got %d calls, %d distinct", rec.count, len(rec.calls))
	}
	for x := 0; x < rec.width; x++ {
		for y := 0; y < rec.height; y++ {
			if _, ok := rec.calls[[2]int{x, y}]; !ok {
				t.Errorf("pixel (%d,%d) never written", x, y)
			}
		}
	}
}

func TestImageIsFlippedVertically(t *testing.T) {
	rec := render(false)
	layer, _ := testLayers()
	for x := range layer {
		for y := range layer[x] {
			if got := rec.calls[[2]int{x, rec.height - 1 - y}].depth; got != layer[x][y] {
				t.Errorf("layer[%d][%d]=%v ended up as %v at row %d", x, y, layer[x][y], got, rec.height-1-y)
			}
		}
	}
}

func TestShadowDepthPassedToProfile(t *testing.T) {
	rec := render(true)
	_, shadow := testLayers()
	for x := range shadow {
		for y := range shadow[x] {
			p := rec.calls[[2]int{x, rec.height - 1 - y}]
			if !p.isShadow || p.shadowDepth != shadow[x][y] {
				t.Errorf("(%d,%d): shadowDepth = %v, want %v", x, y, p.shadowDepth, shadow[x][y])
			}
		}
	}
}
