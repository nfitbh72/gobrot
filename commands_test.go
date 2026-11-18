package main

import "testing"

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

// The commands from the issue that used to panic must now be refused.
func TestMalformedCommandsDoNotPanic(t *testing.T) {
	c := newTestCommands()
	for _, line := range []string{"width", "ascii 80", "coords -2 2 -2", "width 1x", "help coords", "layers"} {
		c.CLI.Dispatch(line)
	}
}
