package fractalxml

import (
	"encoding/xml"
	"fmt"
	"strconv"
)

//struct for the top level xml
type TMandelbrotXMLSettings struct {
	XMLName xml.Name `xml:"mandelbrot_settings_2"`
	Coords TCoords `xml:"limits"`
	Iterations TIterations `xml:"max_iterations"`
	ImageSize TImageSize `xml:"image_size"`
}

//struct for the coords part of the xml
type TCoords struct {
	XMLName xml.Name `xml:"limits"`
	XMin string `xml:"xmin"`
	YMin string `xml:"ymin"`
	XMax string `xml:"xmax"`
	YMax string `xml:"ymax"`
}

//struct for the iterations part of the xml
type TIterations struct {
	XMLName xml.Name `xml:"max_iterations"`
	Max string `xml:"value,attr"`
}

//struct for the image size part of the xml
type TImageSize struct {
	XMLName xml.Name `xml:"image_size"`
	Width string `xml:"width,attr"`
	Height string `xml:"height,attr"`
}

//parses the XML and returns a new TMandelbrotXMLSettings object
func GetMandelBrotSettingsFromString(fractalXml string) TMandelbrotXMLSettings {
	var mandelbrotXMLSettings TMandelbrotXMLSettings
	xml.Unmarshal([]byte(fractalXml), &mandelbrotXMLSettings);
	mandelbrotXMLSettings.FixAspectBug()
	fmt.Println(mandelbrotXMLSettings)
	return mandelbrotXMLSettings
}

//assumes no error
func strToFloat64(s string) float64 {
    f, _ := strconv.ParseFloat(s, 64)
    return f
}

func strToUint16(s string) uint16 {
	i, _ := strconv.ParseUint(s, 10, 16)
	return uint16(i)
}

//there is a bug in the XML where if you change the aspect ratio the y min/max assumes the 800x600 ratio
func (settings *TMandelbrotXMLSettings) FixAspectBug() {
	min := settings.GetMin()
	max := settings.GetMax()
	//width := float64(settings.GetWidth())
	//height := float64(settings.GetHeight())
	deltaX := real(max) - real(min)
	deltaY := imag(max) - imag(min)
	diff := deltaX - deltaY
	minY := imag(min) - diff / 2.0
	maxY := imag(max) + diff / 2.0
	settings.Coords.YMin = fmt.Sprintf("%e", minY)
	settings.Coords.YMax = fmt.Sprintf("%e", maxY)
}


//gets the minX, minY
func (settings *TMandelbrotXMLSettings) GetMin() complex128 {
	return complex(
		strToFloat64(settings.Coords.XMin),
		strToFloat64(settings.Coords.YMin),
	)
}

//gets the maxX, maxY
func (settings *TMandelbrotXMLSettings) GetMax() complex128 {
	return complex(
		strToFloat64(settings.Coords.XMax),
		strToFloat64(settings.Coords.YMax),
	)
}

//gets the max iterations
func (settings *TMandelbrotXMLSettings) GetMaxIterations() uint16 {
	return strToUint16(settings.Iterations.Max)
}

//gets the width
func (settings *TMandelbrotXMLSettings) GetWidth() uint16 {
	return strToUint16(settings.ImageSize.Width)
}

//gets the height
func (settings *TMandelbrotXMLSettings) GetHeight() uint16 {
	return strToUint16(settings.ImageSize.Height)
}


func TestFractalXml() {
	testXml := `<?xml version='1.0'?>
<mandelbrot_settings_2>
<image_size width='800' height='600'/>
<limits>
   <xmin>-0.739226440637529520517</xmin>
   <xmax>-0.739125603734140714117</xmax>
   <ymin>0.183842015829915221496</ymin>
   <ymax>0.183917572959250164696</ymax>
</limits>
<palette colorType='HSB'>
   <divisionPoint position='0' color='0;1;1'/>
   <divisionPoint position='1' color='1;1;1'/>
</palette>
<palette_mapping length='250' offset='0'/>
<max_iterations value='2500'/>
</mandelbrot_settings_2>
`
	mandelbrotSettings := GetMandelBrotSettingsFromString(testXml)
	fmt.Println(mandelbrotSettings)
	fmt.Println(mandelbrotSettings.Coords.XMin)
	fmt.Println(mandelbrotSettings.Coords.YMin)
	fmt.Println(mandelbrotSettings.Coords.XMax)
	fmt.Println(mandelbrotSettings.Coords.YMax)
	fmt.Println(mandelbrotSettings.Iterations.Max)
	//return mandelbrotSettings
}