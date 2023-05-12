package main

import (
    "fmt"
    "strconv"
    "wilf/gobrot/pkg/fractals"
    "wilf/gobrot/pkg/images"
    "wilf/gobrot/pkg/xml"
    "wilf/gobrot/pkg/cli"
    "wilf/gobrot/pkg/canvas"
    //"math"
)

//keeping all the main data here
type TMain struct {
    Mandel mandelbrot.TMandel
    Image slicetoimage.TImage
    CLI cli.TCLI
    Pixels [][]float64
    Canvas canvas.TCanvas
}

func (m *TMain) setHeight(arguments []int) {
    m.Mandel.Height = uint16(arguments[0])
    fmt.Println("Height set to", m.Mandel.Height)
}

func (m *TMain) setWidth(arguments []int) {
    m.Mandel.Width = uint16(arguments[0])
    fmt.Println("width set to", m.Mandel.Width)
}

func (m *TMain) setMaxIterations(arguments []int) {
    m.Mandel.MaxIterations = uint16(arguments[0])
    fmt.Println("max iterations set to", m.Mandel.MaxIterations)
}

func (m *TMain) setCanvasStrength(arguments[] float64) {
    m.Canvas.MandelEffect = 1.0 - arguments[0]
    fmt.Println("mandel effect set to", m.Canvas.MandelEffect)
}

func (m *TMain) setCoords(arguments []float64) {
    m.Mandel.Min = complex(arguments[0], arguments[2])
    m.Mandel.Max = complex(arguments[1], arguments[3])
    fmt.Println("coords", m.Mandel.Min, m.Mandel.Max)
}

func (m *TMain) setFilename(arguments []string) {
    m.Image.FilenameBase = arguments[0]
}

//actually generates the mandel and saves the image
func (m *TMain) buildMandel(arguments []string) {
    m.Pixels = m.Mandel.GetMandel()
    m.Image.SaveImage(m.Pixels)
}

func (m *TMain) buildMandelAscii(arguments []int) {
    //clone the mandel
    asciiMandelConfig := m.Mandel
    //set width and height
    asciiMandelConfig.Width = uint16(arguments[0])
    asciiMandelConfig.Height = uint16(arguments[1])
    //generate the mandel and then call PrintAscii function
    var pixels[][]float64 = asciiMandelConfig.GetMandel()
    slicetoimage.PrintAscii(pixels, asciiMandelConfig.Width, asciiMandelConfig.Height, asciiMandelConfig.MaxIterations)
}

func (m *TMain) loadXml(xml string) {
    mandelbrotSettings := fractalxml.GetMandelBrotSettingsFromString(xml);
    m.Mandel.Min = mandelbrotSettings.GetMin()
    m.Mandel.Max = mandelbrotSettings.GetMax()
    m.Mandel.MaxIterations = mandelbrotSettings.GetMaxIterations()
    fmt.Println("Mandel configured with", m.Mandel.Min, m.Mandel.Max, m.Mandel.MaxIterations)
    //delta := m.Mandel.Max - m.Mandel.Min
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

//command to show the current configuration
func (main *TMain) showConfig(arguments []string) {
    fmt.Println(main.Mandel)
    fmt.Println(main.Image)
}

//command to execute the canvas
func (main *TMain) canvas(arguments []string) {
    
    main.Canvas = canvas.TCanvas{
        InPixels: main.Pixels,
        MandelEffect: 0.1,
    }
    if (len(arguments) > 0) {
        main.Canvas.CanvasMethodName = arguments[0]
    } else {
        main.Canvas.CanvasMethodName = "test"
    }
    main.Canvas.Run(main.Mandel.Width, main.Mandel.Height, main.Mandel.MaxIterations)
    main.Image.SaveImage(main.Canvas.OutPixels)
}

var Main TMain

func main() {
    //default config for mandelbrot
    Main.Mandel = mandelbrot.TMandel {
        Width: uint16(8161),
        Height: 8161,
        Min: -2.0 - 1.25i,
        Max: 0.5 + 1.25i,
        MaxIterations: 2047,
    }

    //default config for image output
    Main.Image = slicetoimage.TImage {
        Layers: [][]int{{0, 100}, {100, 100}, {100, 100}, {100, 100}},
        FilenameBase: "./out",
    }

    //create the CLI
    Main.CLI = cli.TCLI{
        Commands: []cli.TCommand{},
        HelpText: "This is GOBROT - it creates fractals in the form of PNG files.\nTo get more detailed help per command: GoBrot.exe help <command>",
    }

    //add the commands to the CLI
    Main.CLI.AddCommand("fname", 1, 1, Main.setFilename, "Sets the basename of the file output")
    Main.CLI.AddCommand("height", 1, 1, Main.setHeight, "Sets the height in pixels of the output")
    Main.CLI.AddCommand("width", 1, 1, Main.setWidth, "Sets the width in pixels of the output")
    Main.CLI.AddCommand("iterations", 1, 1, Main.setMaxIterations, "Sets the maximum iterations of the mandelbrot algorithm")
    Main.CLI.AddCommand("cstr", 1, 1, Main.setCanvasStrength, "Sets the strength of the canvas (compared to the fractal)")
    Main.CLI.AddCommand("coords", 4, 4, Main.setCoords, "Sets the real-min, imaginary-min, real-max, imaginary-max values")
    Main.CLI.AddCommand("config", 0, 0, Main.showConfig, "Displays the configuration currently being used")
    Main.CLI.AddCommand("go", 0, 0, Main.buildMandel, "Runs the algorithm, outputs the png(s)")
    Main.CLI.AddCommand("layers", 1, 0, Main.setLayers, "If you want many pngs for different layers then specify percentages for which the layer breaks exist (do not include 0 or 100, these are assumed)")
    Main.CLI.AddCommand("ascii", 1, 2, Main.buildMandelAscii, "Displays an ascii version of the Mandelbrot - specify width and height")
    Main.CLI.AddCommand("canvas", 0, 1, Main.canvas, "Applies the fractal to a canvas")
    Main.CLI.AddMultiLineCommand("<?xml version='1.0'?>", "</mandelbrot_settings_2>", Main.loadXml, "Imports XML from https://math.hws.edu/eck/js/mandelbrot/MB.html")

    //welcome message
    fmt.Println("Welcome to GOBROT")
    fmt.Println()
    fmt.Println("Add Commands / XML here:")

    //run the CLI
    Main.CLI.Run()

}