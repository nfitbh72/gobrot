package fractalxml

import (
	"encoding/xml"
	"fmt"
)
type MandelbrotXMLSettings struct {
	XMLName xml.Name `xml:"mandelbrot_settings_2"`
	Coords Coords `xml:"limits"`
	Iterations Iterations `xml:"max_iterations"`
}

type Coords struct {
	XMLName xml.Name `xml:"limits"`
	XMin string `xml:"xmin"`
	YMin string `xml:"ymin"`
	XMax string `xml:"xmax"`
	YMax string `xml:"ymax"`
}

type Iterations struct {
	XMLName xml.Name `xml:"max_iterations"`
	Max string `xml:"value,attr"`
}

func GetMandelBrotSettings(fractalXmlBytes []byte) MandelbrotXMLSettings {
	var mandelbrotXMLSettings MandelbrotXMLSettings
	xml.Unmarshal(fractalXmlBytes, &mandelbrotXMLSettings);
	return mandelbrotXMLSettings
}

func GetMandelBrotSettingsFromString(fractalXml string) MandelbrotXMLSettings {
	fractalXmlBytes := []byte(fractalXml)
	return GetMandelBrotSettings(fractalXmlBytes)
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