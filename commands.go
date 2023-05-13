package main

import (
    "fmt"
    "wilf/gobrot/pkg/images"
    "wilf/gobrot/pkg/xml"
    "wilf/gobrot/pkg/cli"
    "wilf/gobrot/pkg/canvas"
    "wilf/gobrot/pkg/usage"
)

type TCommands struct {
    AppData *TAppData
    CLI cli.TCLI
    Usage usage.TUsage
}

func (m *TCommands) initialise(appData *TAppData) {
    m.AppData = appData
    //create the CLI
    m.CLI = cli.TCLI{
        Commands: []cli.TCommand{},
        HelpText: "This is GOBROT - it creates fractals in the form of PNG files.\nTo get more detailed help per command: GoBrot.exe help <command>",
    }

    //add the commands to the CLI
    m.CLI.AddCommand("fname", 1, 1, m.setFilename, "Sets the basename of the file output")
    m.CLI.AddCommand("height", 1, 1, m.setHeight, "Sets the height in pixels of the output")
    m.CLI.AddCommand("width", 1, 1, m.setWidth, "Sets the width in pixels of the output")
    m.CLI.AddCommand("iterations", 1, 1, m.setMaxIterations, "Sets the maximum iterations of the mandelbrot algorithm")
    m.CLI.AddCommand("cstr", 1, 1, m.setCanvasStrength, "Sets the strength of the canvas (compared to the fractal)")
    m.CLI.AddCommand("coords", 4, 4, m.setCoords, "Sets the real-min, imaginary-min, real-max, imaginary-max values")
    m.CLI.AddCommand("config", 0, 0, m.showConfig, "Displays the configuration currently being used")
    m.CLI.AddCommand("go", 0, 0, m.buildMandel, "Runs the algorithm, outputs the png(s)")
    m.CLI.AddCommand("layers", 1, 0, m.setLayers, "If you want many pngs for different layers then specify percentages for which the layer breaks exist (do not include 0 or 100, these are assumed)")
    m.CLI.AddCommand("ascii", 1, 2, m.buildMandelAscii, "Displays an ascii version of the Mandelbrot - specify width and height")
    m.CLI.AddCommand("canvas", 0, 1, m.canvas, "Applies the fractal to a canvas")
    m.CLI.AddCommand("usage", 0, 0, m.usage, "Reports on memory usage")
    m.CLI.AddMultiLineCommand("<?xml version='1.0'?>", "</mandelbrot_settings_2>", m.loadXml, "Imports XML from https://math.hws.edu/eck/js/mandelbrot/MB.html")
}

func (m *TCommands) run() {
    m.CLI.Run()
}

func (m *TCommands) setHeight(arguments []int) {
    m.AppData.Mandel.Height = uint16(arguments[0])
    fmt.Println("Height set to", m.AppData.Mandel.Height)
}

func (m *TCommands) setWidth(arguments []int) {
    m.AppData.Mandel.Width = uint16(arguments[0])
    fmt.Println("width set to", m.AppData.Mandel.Width)
}

func (m *TCommands) setMaxIterations(arguments []int) {
    m.AppData.Mandel.MaxIterations = uint16(arguments[0])
    fmt.Println("max iterations set to", m.AppData.Mandel.MaxIterations)
}

func (m *TCommands) setCanvasStrength(arguments[] float64) {
    m.AppData.Canvas.MandelEffect = 1.0 - arguments[0]
    fmt.Println("mandel effect set to", m.AppData.Canvas.MandelEffect)
}

func (m *TCommands) setCoords(arguments []float64) {
    m.AppData.Mandel.Min = complex(arguments[0], arguments[2])
    m.AppData.Mandel.Max = complex(arguments[1], arguments[3])
    fmt.Println("coords", m.AppData.Mandel.Min, m.AppData.Mandel.Max)
}

func (m *TCommands) setFilename(arguments []string) {
    m.AppData.Image.FilenameBase = arguments[0]
}

//actually generates the mandel and saves the image
func (m *TCommands) buildMandel(arguments []string) {
    m.AppData.Pixels = m.AppData.Mandel.GetMandel()
    m.AppData.Layers.ProcessLayers(m.AppData.Pixels)
    m.AppData.Image.SaveImage(m.AppData.Layers.Layers)
}

func (m *TCommands) buildMandelAscii(arguments []int) {
    //clone the mandel
    asciiMandelConfig := m.AppData.Mandel
    //set width and height
    asciiMandelConfig.Width = uint16(arguments[0])
    asciiMandelConfig.Height = uint16(arguments[1])
    //generate the mandel and then call PrintAscii function
    var pixels[][]float64 = asciiMandelConfig.GetMandel()
    slicetoimage.PrintAscii(pixels, asciiMandelConfig.Width, asciiMandelConfig.Height, asciiMandelConfig.MaxIterations)
}

func (m *TCommands) loadXml(xml string) {
    mandelbrotSettings := fractalxml.GetMandelBrotSettingsFromString(xml);
    m.AppData.Mandel.Min = mandelbrotSettings.GetMin()
    m.AppData.Mandel.Max = mandelbrotSettings.GetMax()
    m.AppData.Mandel.MaxIterations = mandelbrotSettings.GetMaxIterations()
    fmt.Println("Mandel configured with", m.AppData.Mandel.Min, m.AppData.Mandel.Max, m.AppData.Mandel.MaxIterations)
    //delta := m.Mandel.Max - m.Mandel.Min
}

func (m *TCommands) setLayers(arguments []float64) {
    if len(arguments) < 1 {
        m.AppData.Layers.LayersConfig[0][0] = 0.0
        m.AppData.Layers.LayersConfig[0][1] = 1.0
    } else {
        m.AppData.Layers.LayersConfig[0][0] = 0.0
        count := 0
        for _, layerValue := range arguments {
            fmt.Println(layerValue)

            m.AppData.Layers.LayersConfig[count][1] = layerValue
            m.AppData.Layers.LayersConfig[count+1][0] = layerValue
            count++
            
        }
        m.AppData.Layers.LayersConfig[count][1] = 1.0
    }
    fmt.Println("layers set", m.AppData.Layers.LayersConfig)
}

//command to show the current configuration
func (m *TCommands) showConfig(arguments []string) {
    fmt.Println(m.AppData.Mandel)
    fmt.Println(m.AppData.Layers.LayersConfig)
    fmt.Println(m.AppData.Image)
}

//command to execute the canvas
func (m *TCommands) canvas(arguments []string) {
    
    m.AppData.Canvas = canvas.TCanvas{
        InPixels: m.AppData.Pixels,
        MandelEffect: 0.1,
    }
    if (len(arguments) > 0) {
        m.AppData.Canvas.CanvasMethodName = arguments[0]
    } else {
        m.AppData.Canvas.CanvasMethodName = "test"
    }
    m.AppData.Canvas.Run(m.AppData.Mandel.Width, m.AppData.Mandel.Height, m.AppData.Mandel.MaxIterations)
    m.AppData.Layers.ProcessLayers(m.AppData.Canvas.OutPixels)
    m.AppData.Image.SaveImage(m.AppData.Layers.Layers)
}

//command to report on memory usage
func (m *TCommands) usage(arguments[] string) {
    m.Usage.PrintMemUsage()
}
