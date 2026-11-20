package main

import (
	"github.com/nfitbh72/gobrot/pkg/fractals"
	"reflect"
	"testing"
)

func newTestCommands() *TCommands {
	appData := &TAppData{}
	appData.initialise()
	c := &TCommands{}
	c.initialise(appData)
	return c
}

func TestSetWidthHeightRange(t *testing.T) {
	c := newTestCommands()
	c.setWidth([]int{100000}) // used to wrap to 34464
	c.setHeight([]int{0})
	if c.AppData.Fractal.Width != 1024 || c.AppData.Fractal.Height != 1024 {
		t.Errorf("out of range value was applied: %dx%d", c.AppData.Fractal.Width, c.AppData.Fractal.Height)
	}
	c.setWidth([]int{800})
	c.setHeight([]int{65535})
	if c.AppData.Fractal.Width != 800 || c.AppData.Fractal.Height != 65535 {
		t.Errorf("valid value was rejected: %dx%d", c.AppData.Fractal.Width, c.AppData.Fractal.Height)
	}
}

func TestSetMaxIterationsRange(t *testing.T) {
	c := newTestCommands()
	c.setMaxIterations([]int{-1}) // used to wrap to a huge uint
	c.setMaxIterations([]int{0})
	if c.AppData.Fractal.MaxIterations != 512 {
		t.Errorf("invalid iterations applied: %d", c.AppData.Fractal.MaxIterations)
	}
	c.setMaxIterations([]int{100})
	if c.AppData.Fractal.MaxIterations != 100 {
		t.Errorf("iterations = %d", c.AppData.Fractal.MaxIterations)
	}
}

// The commands from the issue that used to panic must now be refused,
// leaving the configuration untouched.
func TestMalformedCommandsAreRefused(t *testing.T) {
	c := newTestCommands()
	for _, line := range []string{"width", "width 1x", "width 100000", "height", "ascii 80", "coords -1 1 -1", "help coords"} {
		c.CLI.Dispatch(line)
	}
	f := c.AppData.Fractal
	if f.Width != 1024 || f.Height != 1024 {
		t.Errorf("size changed to %dx%d", f.Width, f.Height)
	}
	if minX, _ := f.MinX.Float64(); minX != -2 {
		t.Errorf("coords changed: MinX = %v", minX)
	}
}

// cstr must survive the canvas command rebuilding the canvas
func TestCanvasKeepsConfiguredStrength(t *testing.T) {
	c := newTestCommands()
	c.setCanvasStrength([]float64{0.4})
	c.prepareCanvas([]string{"dome"})
	if got := c.AppData.Canvas.MandelEffect; got != 0.6 {
		t.Errorf("MandelEffect = %v after cstr 0.4, want 0.6", got)
	}
	if c.AppData.Canvas.CanvasMethodName != "dome" {
		t.Errorf("method = %q", c.AppData.Canvas.CanvasMethodName)
	}
}

func TestCanvasDefaultStrengthWithoutCstr(t *testing.T) {
	c := newTestCommands()
	c.prepareCanvas(nil)
	if got := c.AppData.Canvas.MandelEffect; got != 0.1 {
		t.Errorf("default MandelEffect = %v, want 0.1", got)
	}
}

// a failed canvas run must not re-process the previous run's output
func TestPrepareCanvasClearsStaleOutput(t *testing.T) {
	c := newTestCommands()
	c.AppData.Pixels = [][]float64{{0.5}}
	c.prepareCanvas([]string{"dome"})
	c.AppData.Canvas.Run(1, 1, 10)
	if c.AppData.Canvas.OutPixels == nil {
		t.Fatal("setup: dome run produced no output")
	}
	c.prepareCanvas([]string{"bogus"})
	c.AppData.Canvas.Run(1, 1, 10) // invalid method: returns early
	if c.AppData.Canvas.OutPixels != nil {
		t.Errorf("stale OutPixels from previous run: %v", c.AppData.Canvas.OutPixels)
	}
}

func TestBigAppliesToAlreadySetBounds(t *testing.T) {
	c := newTestCommands()
	c.setCoords([]string{"-0.7436447860", "-0.7436447859", "0.1318252536", "0.1318252537"})
	c.big([]int{100})
	f := c.AppData.Fractal
	for name, b := range map[string]interface{ Prec() uint }{"MinX": f.MinX, "MaxX": f.MaxX, "MinY": f.MinY, "MaxY": f.MaxY} {
		if b.Prec() != 100 {
			t.Errorf("%s precision = %d after big 100, want 100", name, b.Prec())
		}
	}
	// the original decimal text is re-parsed, not widened from the 53-bit value
	if got := f.MinX.Text('g', 25); got != "-0.743644786" {
		t.Errorf("MinX = %s", got)
	}
}

func TestSetFractalTypeUnknownKeepsCurrent(t *testing.T) {
	c := newTestCommands()
	c.setFractalType([]string{"bsjulia"})
	c.setFractalType([]string{"foo"})
	if c.AppData.Fractal.FractalType != fractals.BSJulia {
		t.Errorf("unknown type changed FractalType to %v", c.AppData.Fractal.FractalType)
	}
	c.setFractalType([]string{"mandel"})
	if c.AppData.Fractal.FractalType != fractals.Mandel {
		t.Errorf("mandel not applied")
	}
}

func TestCoordsHelpTextMatchesArgumentOrder(t *testing.T) {
	c := newTestCommands()
	for _, cmd := range c.CLI.Commands {
		if cmd.Keyword == "coords" {
			want := "Sets the real-min, real-max, imaginary-min, imaginary-max values"
			if len(cmd.HelpText) != 1 || cmd.HelpText[0] != want {
				t.Errorf("coords help = %v, want %q", cmd.HelpText, want)
			}
			return
		}
	}
	t.Fatal("coords command not registered")
}

func TestLayersConfig(t *testing.T) {
	got, err := layersConfig(nil)
	if err != nil || !reflect.DeepEqual(got, [][]float64{{0, 1}}) {
		t.Errorf("no breaks: %v, %v", got, err)
	}
	got, err = layersConfig([]float64{0.25, 0.5})
	if err != nil || !reflect.DeepEqual(got, [][]float64{{0, 0.25}, {0.25, 0.5}, {0.5, 1}}) {
		t.Errorf("two breaks: %v, %v", got, err)
	}
	for _, bad := range [][]float64{{1.0}, {0.9, 0.1}, {1.5}, {0.5, 0.5}, {0}, {-0.1}} {
		if _, err := layersConfig(bad); err == nil {
			t.Errorf("layersConfig(%v) accepted invalid breaks", bad)
		}
	}
}

func TestSetLayersRejectsInvalidAndKeepsConfig(t *testing.T) {
	c := newTestCommands()
	c.setLayers([]float64{1.0}) // used to panic on the following go
	c.setLayers([]float64{0.9, 0.1})
	if !reflect.DeepEqual(c.AppData.Layers.LayersConfig, [][]float64{{0, 1}}) {
		t.Errorf("config changed: %v", c.AppData.Layers.LayersConfig)
	}
	c.setLayers([]float64{0.5})
	if len(c.AppData.Layers.LayersConfig) != 2 {
		t.Errorf("valid layers rejected: %v", c.AppData.Layers.LayersConfig)
	}
}

func TestSetToleranceRange(t *testing.T) {
	c := newTestCommands()
	if c.AppData.ThreeD.Tolerance != 0.02 {
		t.Errorf("default tolerance = %v, want 0.02", c.AppData.ThreeD.Tolerance)
	}
	c.setTolerance([]float64{-0.1})
	if c.AppData.ThreeD.Tolerance != 0.02 {
		t.Errorf("negative tolerance was applied: %v", c.AppData.ThreeD.Tolerance)
	}
	c.setTolerance([]float64{0})
	if c.AppData.ThreeD.Tolerance != 0 {
		t.Errorf("tolerance 0 was rejected: %v", c.AppData.ThreeD.Tolerance)
	}
}
