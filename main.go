package main

import (
    "fmt"
    "strconv"
    "wilf/gobrot/pkg/fractals"
    "wilf/gobrot/pkg/images"
    "wilf/gobrot/pkg/xml"
    "wilf/gobrot/pkg/cli"
)

type TMain struct {
    Mandel mandelbrot.MandelConfig
    Image slicetoimage.ImageConfig
    CLI cli.CLI
    Pixels [][]uint16
}

func (m *TMain) setHeight(arguments []int) {
    m.Mandel.Height = uint16(arguments[0])
    fmt.Println("Height set to", m.Mandel.Height)
}

func (m TMain) setWidth(arguments []int) {
    m.Mandel.Width = uint16(arguments[0])
    fmt.Println("width set to", m.Mandel.Width)
}

func (m *TMain) setMaxIterations(arguments []int) {
    m.Mandel.MaxIterations = uint16(arguments[0])
    fmt.Println("max iterations set to", m.Mandel.MaxIterations)
}

func (m *TMain) setCoords(arguments []float64) {
    m.Mandel.Xmin = arguments[0]
    m.Mandel.Xmax = arguments[1]
    m.Mandel.Ymin = arguments[2]
    m.Mandel.Ymax = arguments[3]
    fmt.Println("coords", m.Mandel.Xmin, m.Mandel.Xmax, m.Mandel.Ymin, m.Mandel.Ymax)
}

func (m *TMain) setFilename(arguments []string) {
    m.Image.FilenameBase = arguments[0]
}

func (m *TMain) buildMandel(arguments []string) {
    m.Pixels = m.Mandel.GetMandel()
    m.Image.SaveImage(m.Pixels)
}

func (m *TMain) buildMandelAscii(arguments []int) {
    asciiMandelConfig := m.Mandel
    asciiMandelConfig.Width = uint16(arguments[0])
    asciiMandelConfig.Height = uint16(arguments[1])
    var pixels[][]uint16 = asciiMandelConfig.GetMandel()
    slicetoimage.PrintAscii(pixels, asciiMandelConfig.Width, asciiMandelConfig.Height, asciiMandelConfig.MaxIterations)
}

func (m *TMain) loadXml(xml string) {
    mandelbrotSettings := fractalxml.GetMandelBrotSettingsFromString(xml);
    m.Mandel.Xmin, _ = strconv.ParseFloat(mandelbrotSettings.Coords.XMin, 64)
    m.Mandel.Ymin, _ = strconv.ParseFloat(mandelbrotSettings.Coords.YMin, 64)
    m.Mandel.Xmax, _ = strconv.ParseFloat(mandelbrotSettings.Coords.XMax, 64)
    m.Mandel.Ymax, _ = strconv.ParseFloat(mandelbrotSettings.Coords.YMax, 64)
    its, _ := strconv.ParseUint(mandelbrotSettings.Iterations.Max, 10, 16)
    m.Mandel.MaxIterations = uint16(its)
    fmt.Println("Firing up the threads to generate Mandelbrot")
}

func (m *TMain) setLayers(arguments []string) {
    if len(arguments) < 1 {
        m.Image.Layers[0][0] = 0
        m.Image.Layers[0][1] = 100
    } else {
        m.Image.Layers[0][0] = 0
        count := 0
        for _, layer := range arguments {
            fmt.Println(layer)
            if intVal, err := strconv.ParseInt(layer,10,64); err == nil {
                m.Image.Layers[count][1] = int(intVal)
                m.Image.Layers[count+1][0] = int(intVal)
                count++
            } else {
                fmt.Printf("Error:%v\n", err)
            }
        }
        m.Image.Layers[count][1] = 100
    }
    fmt.Println("layers set", m.Image.Layers)
}

func (m *TMain) showConfig(arguments []string) {
    fmt.Println(m.Mandel)
    fmt.Println(m.Image)
}

//var mandelConfig mandelbrot.MandelConfig
//var imageConfig slicetoimage.ImageConfig
//var clint cli.CLI

var Main TMain

func main() {
    //default config
    layers := [][]int{{0, 100}, {100, 100}, {100, 100}, {100, 100}}
    Main.Mandel = mandelbrot.MandelConfig{
        Width: uint16(8161),
        Height: 8161,
        Xmin: -2.0,
        Xmax: 0.5,
        Ymin: -1.25,
        Ymax: 1.25,
        MaxIterations: 2047,
    }
    Main.Image = slicetoimage.ImageConfig{
        Layers: layers,
        FilenameBase: "./out",
    }
    Main.CLI = cli.CLI{
        Commands: []cli.Command{},
        HelpText: "This is GOBROT - it creates fractals in the form of PNG files.\nTo get more detailed help per command: GoBrot.exe help <command>",
    }

    fmt.Println("Welcome to GOBROT")
    fmt.Println()
    fmt.Println("Add Commands / XML here:")

    Main.CLI.AddCommand("height", 1, 1, Main.setHeight, "Sets the height in pixels of the output")
    Main.CLI.AddCommand("width", 1, 1, Main.setWidth, "Sets the width in pixels of the output")
    Main.CLI.AddCommand("iterations", 1, 1, Main.setMaxIterations, "Sets the maximum iterations of the mandelbrot algorithm")
    Main.CLI.AddCommand("coords", 4, 4, Main.setCoords, "Sets the real-min, imaginary-min, real-max, imaginary-max values")
    Main.CLI.AddCommand("config", 0, 0, Main.showConfig, "Displays the configuration currently being used")
    Main.CLI.AddCommand("go", 0, 0, Main.buildMandel, "Runs the algorithm, outputs the png(s)")
    Main.CLI.AddCommand("layers", 1, 0, Main.setLayers, "If you want many pngs for different layers then specify percentages for which the layer breaks exist (do not include 0 or 100, these are assumed)")
    Main.CLI.AddCommand("ascii", 1, 2, Main.buildMandelAscii, "Displays an ascii version of the Mandelbrot - specify width and height")
    //clint.AddCommand("nil", 0, 0, nil)
    Main.CLI.AddMultiLineCommand("<?xml version='1.0'?>", "</mandelbrot_settings_2>", Main.loadXml, "Imports XML from https://math.hws.edu/eck/js/mandelbrot/MB.html")

    Main.CLI.Run()

}