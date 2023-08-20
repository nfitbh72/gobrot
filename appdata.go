package main

import (
    "wilf/gobrot/pkg/fractals"
    "wilf/gobrot/pkg/images"
    "wilf/gobrot/pkg/canvas"
    "wilf/gobrot/pkg/layers"
    "wilf/gobrot/pkg/threed"
    "sync"
)

type TAppData struct {
    Mandel mandelbrot.TMandel
    Layers layers.TLayers
    Image images.TImage
    ThreeD threed.TThreeD
    Pixels [][]float64
    ShadowPixels [][]float64
    Canvas canvas.TCanvas
    IsShadowGenerated bool
    StretchValues bool
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
    appData.Mandel = mandelbrot.TMandel {
        Width: uint16(512),
        Height: 512,
        Min: -2.0 - 1.25i,
        Max: 0.5 + 1.25i,
        MaxIterations: 2047,
    }

    //default config for layers
    appData.Layers = layers.TLayers {
        LayersConfig: [][]float64{{0.0, 1.0}},
    }

    //default config for image output
    appData.Image = images.TImage {
        FilenameBase: "./out",
        ShowPreview: false,
    }
    appData.Image.Initialise()
    defaultColourProfileString := "gray2"
    if appData.Image.SetColourProfile(defaultColourProfileString) {
        appData.ColourProfileString = defaultColourProfileString
    }

    appData.ThreeD = threed.TThreeD {
        FilenameBase: "./out2",
        InverseDepth: true,
    }
    //whether to create the shadow mandel plot
    appData.IsShadowGenerated = false

    //whether to stretch the values to 0 -> 65536
    appData.StretchValues = true
}