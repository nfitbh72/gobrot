package main

import (
    "fmt"
    "wilf/gobrot/pkg/fractals"
    "wilf/gobrot/pkg/images"
    "wilf/gobrot/pkg/canvas"
    "wilf/gobrot/pkg/layers"
    "sync"
)

type TAppData struct {
    Mandel mandelbrot.TMandel
    Layers layers.TLayers
    Image slicetoimage.TImage
    Pixels [][]float64
    Canvas canvas.TCanvas
}

var lock = &sync.Mutex{}

var appDataInstance *TAppData

func getAppDataInstance() *TAppData {
    if appDataInstance == nil {
        lock.Lock()
        defer lock.Unlock()
        if appDataInstance == nil {
            fmt.Println("Creating single instance now.")
            appDataInstance = &TAppData{}
            appDataInstance.initialise()
        } else {
            fmt.Println("Single instance already created.")
        }
    } else {
        fmt.Println("Single instance already created.")
    }

    return appDataInstance
}

func (appData *TAppData) initialise() {
    //default config for mandelbrot
    appData.Mandel = mandelbrot.TMandel {
        Width: uint16(8161),
        Height: 8161,
        Min: -2.0 - 1.25i,
        Max: 0.5 + 1.25i,
        MaxIterations: 2047,
    }

    //default config for layers
    appData.Layers = layers.TLayers {
        LayersConfig: [][]float64{{0.0, 1.0}},
    }

    //default config for image output
    appData.Image = slicetoimage.TImage {
        FilenameBase: "./out",
    }
}