package main

import (
	"sync"
	"wilf/gobrot/pkg/bigfloat"
	"wilf/gobrot/pkg/canvas"
	"wilf/gobrot/pkg/fractals"
	"wilf/gobrot/pkg/images"
	"wilf/gobrot/pkg/layers"
	"wilf/gobrot/pkg/threed"
)

type TAppData struct {
	Fractal             fractals.TFractal
	Layers              layers.TLayers
	Image               images.TImage
	ThreeD              threed.TThreeD
	Pixels              [][]float64
	ShadowPixels        [][]float64
	Canvas              canvas.TCanvas
	IsShadowGenerated   bool
	StretchValues       bool
	ColourProfileString string
}

var lock = &sync.Mutex{}

var appDataInstance *TAppData

func getAppDataInstance() *TAppData {
	if appDataInstance == nil {
		lock.Lock()
		defer lock.Unlock()
		if appDataInstance == nil {
			appDataInstance = &TAppData{}
			appDataInstance.initialise()
		}
	}
	return appDataInstance
}

func (appData *TAppData) initialise() {
	//default config for mandelbrot
	appData.Fractal = fractals.TFractal{
		Width:         uint16(1024),
		Height:        1024,
		MaxIterations: 512,
		FractalType:   fractals.Mandel,
		Step:          0.0,
		BF:            &bigfloat.TBigFloats{Prec: 53},
	}
	appData.Fractal.SetBounds("-2.0", "2", "-2", "2")

	//default config for layers
	appData.Layers = layers.TLayers{
		LayersConfig: [][]float64{{0.0, 1.0}},
	}

	//default config for image output
	appData.Image = images.TImage{
		FilenameBase: "./out",
		ShowPreview:  false,
	}
	appData.Image.Initialise()
	defaultColourProfileString := "gray2"
	if appData.Image.SetColourProfile(defaultColourProfileString) {
		appData.ColourProfileString = defaultColourProfileString
	}

	appData.ThreeD = threed.TThreeD{
		FilenameBase: "./out2",
		InverseDepth: true,
		Scale:        100.0,
		DepthScale:   0.1,
		CutBelow:     0.0,
		CutAbove:     1.0,
	}
	//whether to create the shadow mandel plot
	appData.IsShadowGenerated = false

	//whether to stretch the values to 0 -> 65536
	appData.StretchValues = true
}
