package main

import (
	"fmt"
	"github.com/nfitbh72/gobrot/pkg/cli"
	"github.com/nfitbh72/gobrot/pkg/fractals"
	"github.com/nfitbh72/gobrot/pkg/shapes"
	"github.com/nfitbh72/gobrot/pkg/usage"
	fractalxml "github.com/nfitbh72/gobrot/pkg/xml"
	"math"
)

type TCommands struct {
	AppData *TAppData
	CLI     cli.TCLI
	Usage   usage.TUsage
}

func (m *TCommands) initialise(appData *TAppData) {
	m.AppData = appData
	//create the CLI
	m.CLI = cli.TCLI{
		Commands: []cli.TCommand{},
		HelpText: "This is GOBROT - it creates fractals in the form of PNG files.\nTo get more detailed help per command: GoBrot.exe help <command>",
	}

	//add the commands to the CLI
	m.CLI.AddCommand("fractal", 1, 1, m.setFractalType, "Sets the type of fractal, currently supports \"mandel\" or \"bsjulia\"")
	m.CLI.AddCommand("fname", 1, 1, m.setFilename, "Sets the basename of the file output")
	m.CLI.AddCommand("height", 1, 1, m.setHeight, "Sets the height in pixels of the output")
	m.CLI.AddCommand("width", 1, 1, m.setWidth, "Sets the width in pixels of the output")
	m.CLI.AddCommand("iterations", 1, 1, m.setMaxIterations, "Sets the maximum iterations of the mandelbrot algorithm")
	m.CLI.AddCommand("cstr", 1, 1, m.setCanvasStrength, "Sets the strength of the canvas (compared to the fractal)")
	m.CLI.AddCommand("coords", 4, 4, m.setCoords, "Sets the real-min, real-max, imaginary-min, imaginary-max values")
	m.CLI.AddCommand("config", 0, 0, m.showConfig, "Displays the configuration currently being used")
	m.CLI.AddCommand("go", 0, 0, m.buildMandel, "Runs the algorithm, outputs the png(s)")
	m.CLI.AddCommand("layers", 0, cli.Unlimited, m.setLayers, "If you want many pngs for different layers then specify percentages for which the layer breaks exist (do not include 0 or 100, these are assumed)")
	m.CLI.AddCommand("ascii", 2, 2, m.buildMandelAscii, "Displays an ascii version of the Mandelbrot - specify width and height")
	m.CLI.AddCommand("canvas", 0, 1, m.canvas, "Applies the fractal to a canvas")
	m.CLI.AddCommand("3d", 0, 0, m.threeD, "Generates an STL file for each layer")
	m.CLI.AddCommand("test", 0, 0, m.threeDTest, "Tests an STL file output")
	m.CLI.AddCommand("depthscale", 1, 1, m.setDepthScale, "Sets the depth scale of 3d output")
	m.CLI.AddCommand("cutbelow", 1, 1, m.setCutBelow, "Sets the value to cut when the value (0-1) is below")
	m.CLI.AddCommand("cutabove", 1, 1, m.setCutAbove, "Sets the value to cut when the value (0-1) is above")
	m.CLI.AddCommand("usage", 0, 0, m.usage, "Reports on memory usage")
	m.CLI.AddCommand("cp", 1, 1, m.setColourProfile, "Sets the colour profile to use")
	m.CLI.AddCommand("shadow", 1, 1, m.setShadow, "Sets whether shadows are used with on / off argument")
	m.CLI.AddCommand("stretch", 1, 1, m.setStretch, "Sets whether to force the range of colours to be 0 -> 65536")
	m.CLI.AddCommand("img", 0, 0, m.buildImageOnly, "Builds in image again from existing data")
	m.CLI.AddCommand("step", 1, 1, m.setStep, "Mods the height to give steps to it")
	m.CLI.AddCommand("circle", 1, 1, m.circle, "Trims the p[x][y] square into a circle")
	m.CLI.AddCommand("big", 1, 1, m.big, "Uses big floats for higher precision - 'big 60' for 60 points of precision")
	m.CLI.AddMultiLineCommand("<?xml version='1.0'?>", "</mandelbrot_settings_2>", m.loadXml, "Imports XML from https://math.hws.edu/eck/js/mandelbrot/MB.html")
}

func (m *TCommands) run() {
	m.CLI.Run()
}

func (m *TCommands) big(arguments []int) {
	m.AppData.Fractal.SetPrecision(uint(arguments[0]))
	fmt.Println("big float precision set to:", m.AppData.Fractal.BF.Prec)
}

func (m *TCommands) circle(arguments []float64) {
	c := shapes.TCircle{
		Pow: arguments[0],
	}
	fmt.Println("Starting Circle")
	m.AppData.Layers.Layers = c.TrimCircle(
		m.AppData.Layers.Layers, m.AppData.Fractal.Width, m.AppData.Fractal.Height,
	)
	fmt.Println("Circle done")
}

// toDimension checks a width/height in pixels fits a uint16 and is non-zero
func toDimension(name string, v int) (uint16, bool) {
	if v < 1 || v > math.MaxUint16 {
		fmt.Printf("%s must be between 1 and %d\n", name, math.MaxUint16)
		return 0, false
	}
	return uint16(v), true
}

func (m *TCommands) setHeight(arguments []int) {
	height, ok := toDimension("height", arguments[0])
	if !ok {
		return
	}
	m.AppData.Fractal.Height = height
	fmt.Println("Height set to", m.AppData.Fractal.Height)
}

func (m *TCommands) setWidth(arguments []int) {
	width, ok := toDimension("width", arguments[0])
	if !ok {
		return
	}
	m.AppData.Fractal.Width = width
	fmt.Println("width set to", m.AppData.Fractal.Width)
}

func (m *TCommands) setStep(arguments []float64) {
	m.AppData.Fractal.Step = arguments[0]
	fmt.Println("mod set to", m.AppData.Fractal.Step)
}

func (m *TCommands) setMaxIterations(arguments []int) {
	if arguments[0] < 1 {
		fmt.Println("iterations must be at least 1")
		return
	}
	m.AppData.Fractal.MaxIterations = uint(arguments[0])
	fmt.Println("max iterations set to", m.AppData.Fractal.MaxIterations)
}

func (m *TCommands) setCanvasStrength(arguments []float64) {
	m.AppData.Canvas.MandelEffect = 1.0 - arguments[0]
	fmt.Println("mandel effect set to", m.AppData.Canvas.MandelEffect)
}

func (m *TCommands) setCoords(arguments []string) {
	m.AppData.Fractal.SetBounds(arguments[0], arguments[1], arguments[2], arguments[3])
	fmt.Println("coords", m.AppData.Fractal.MinX, m.AppData.Fractal.MaxX, m.AppData.Fractal.MinY, m.AppData.Fractal.MaxY)
}

func (m *TCommands) setFilename(arguments []string) {
	m.AppData.Image.FilenameBase = arguments[0]
}

// actually generates the mandel and saves the image
func (m *TCommands) buildMandel(arguments []string) {
	m.AppData.Pixels = m.AppData.Fractal.GetFractal()
	m.AppData.Layers.ProcessLayers(m.AppData.Pixels, false)
	if m.AppData.IsShadowGenerated {
		m.AppData.Fractal.AdjustForShadowMandel()
		m.AppData.ShadowPixels = m.AppData.Fractal.GetFractal()
		m.AppData.Layers.ProcessLayers(m.AppData.ShadowPixels, true)
	}
	m.AppData.Image.SaveImage(
		m.AppData.Layers.Layers, m.AppData.Layers.ShadowLayers, m.AppData.IsShadowGenerated, m.AppData.StretchValues,
	)
}

// saves the layers as stl files
func (m *TCommands) threeD(arguments []string) {
	m.AppData.ThreeD.Save(m.AppData.Layers.Layers)
}

// runs the test function of threed
func (m *TCommands) threeDTest(arguments []string) {
	m.AppData.ThreeD.Test()
}

func (m *TCommands) setDepthScale(arguments []float64) {
	m.AppData.ThreeD.DepthScale = arguments[0]
	fmt.Println("DepthScale set to", arguments[0])
}

func (m *TCommands) setCutBelow(arguments []float64) {
	m.AppData.ThreeD.CutBelow = arguments[0]
	fmt.Println("CutBelow set to", arguments[0])
}

func (m *TCommands) setCutAbove(arguments []float64) {
	m.AppData.ThreeD.CutAbove = arguments[0]
	fmt.Println("CutAbove set to", arguments[0])
}

func (m *TCommands) buildImageOnly(arguments []string) {
	m.AppData.Image.SaveImage(
		m.AppData.Layers.Layers, m.AppData.Layers.ShadowLayers, m.AppData.IsShadowGenerated, m.AppData.StretchValues,
	)
}

func (m *TCommands) buildMandelAscii(arguments []int) {
	//clone the mandel
	asciiMandelConfig := m.AppData.Fractal
	//set width and height
	width, ok := toDimension("width", arguments[0])
	if !ok {
		return
	}
	height, ok := toDimension("height", arguments[1])
	if !ok {
		return
	}
	asciiMandelConfig.Width = width
	asciiMandelConfig.Height = height
	//generate the mandel and then call PrintAscii function
	var pixels [][]float64 = asciiMandelConfig.GetFractal()
	m.AppData.Image.PrintAscii(
		pixels, asciiMandelConfig.Width, asciiMandelConfig.Height, asciiMandelConfig.MaxIterations,
	)
}

func (m *TCommands) loadXml(xml string) {
	mandelbrotSettings := fractalxml.GetMandelBrotSettingsFromString(xml)
	m.AppData.Fractal.SetBounds(
		mandelbrotSettings.Coords.XMin, mandelbrotSettings.Coords.XMax,
		mandelbrotSettings.Coords.YMin, mandelbrotSettings.Coords.YMax,
	)
	m.AppData.Fractal.MaxIterations = mandelbrotSettings.GetMaxIterations()
	fmt.Println(
		"Mandel configured with", m.AppData.Fractal.MinX, m.AppData.Fractal.MaxX,
		m.AppData.Fractal.MinY, m.AppData.Fractal.MaxY, m.AppData.Fractal.MaxIterations,
	)
	//delta := m.Mandel.Max - m.Mandel.Min
	m.buildMandel([]string{})
	m.threeD([]string{})
}

// layersConfig turns layer break points (fractions strictly between 0 and 1, ascending)
// into [min, max] bands covering 0 to 1
func layersConfig(breaks []float64) ([][]float64, error) {
	prev := 0.0
	for _, b := range breaks {
		if !(b > prev && b < 1.0) {
			return nil, fmt.Errorf("layer break points must be strictly ascending and between 0 and 1 (exclusive), got %v", breaks)
		}
		prev = b
	}
	edges := append(append([]float64{0.0}, breaks...), 1.0)
	config := make([][]float64, len(edges)-1)
	for i := range config {
		config[i] = []float64{edges[i], edges[i+1]}
	}
	return config, nil
}

func (m *TCommands) setLayers(arguments []float64) {
	config, err := layersConfig(arguments)
	if err != nil {
		fmt.Println(err)
		return
	}
	m.AppData.Layers.LayersConfig = config
	fmt.Println("layers set", m.AppData.Layers.LayersConfig)
}

// command to show the current configuration
func (m *TCommands) showConfig(arguments []string) {
	fmt.Printf("%+v\n", m.AppData.Fractal)
	fmt.Println(m.AppData.Layers.LayersConfig)
	fmt.Println(m.AppData.Image)
	fmt.Println("ColourProfileString", m.AppData.ColourProfileString)
	fmt.Println("IsShadowGenerated", m.AppData.IsShadowGenerated, "StretchValues", m.AppData.StretchValues)
	fmt.Println("3d", m.AppData.ThreeD)
	fmt.Println("Step", m.AppData.Fractal.Step)
	fmt.Println("big float precision", m.AppData.Fractal.BF.Prec)
}

// prepareCanvas sets up the canvas for a run from the current pixels and the command arguments
func (m *TCommands) prepareCanvas(arguments []string) {
	//only InPixels and the method change: MandelEffect keeps whatever cstr set
	m.AppData.Canvas.InPixels = m.AppData.Pixels
	//a failed run returns early, so drop any output left by the previous run
	m.AppData.Canvas.OutPixels = nil
	if len(arguments) > 0 {
		m.AppData.Canvas.CanvasMethodName = arguments[0]
	} else {
		m.AppData.Canvas.CanvasMethodName = "test"
	}
}

// command to execute the canvas
func (m *TCommands) canvas(arguments []string) {
	m.prepareCanvas(arguments)
	m.AppData.Canvas.Run(m.AppData.Fractal.Width, m.AppData.Fractal.Height, m.AppData.Fractal.MaxIterations)
	//var shadowPixels = [][][]float64
	m.AppData.Layers.ProcessLayers(m.AppData.Canvas.OutPixels, false)
	m.AppData.Image.SaveImage(
		m.AppData.Layers.Layers, m.AppData.Layers.ShadowLayers, m.AppData.IsShadowGenerated, m.AppData.StretchValues,
	)
}

// command to report on memory usage
func (m *TCommands) usage(arguments []string) {
	m.Usage.PrintMemUsage()
}

func (m *TCommands) setColourProfile(arguments []string) {

	if m.AppData.Image.SetColourProfile(arguments[0]) {
		m.AppData.ColourProfileString = arguments[0] //keeping this just to report on it - probably a better way
		fmt.Println("ColourProfileString set to", m.AppData.ColourProfileString)
	}
}

func (m *TCommands) setShadow(arguments []string) {
	m.AppData.IsShadowGenerated = (arguments[0] == "on")
	fmt.Println("IsShadowGenerated", m.AppData.IsShadowGenerated)
}

func (m *TCommands) setStretch(arguments []string) {
	m.AppData.StretchValues = (arguments[0] == "on")
	fmt.Println("StretchValues", m.AppData.StretchValues)
}

func (m *TCommands) setFractalType(arguments []string) {
	switch arguments[0] {
	case "mandel":
		m.AppData.Fractal.FractalType = fractals.Mandel
		fmt.Println("Fractal Type set to mandel")
	case "bsjulia":
		m.AppData.Fractal.FractalType = fractals.BSJulia
		fmt.Println("Fractal Type set to bsjulia")
	default:
		fmt.Printf("unknown fractal type %q, expected \"mandel\" or \"bsjulia\"\n", arguments[0])
	}

}
