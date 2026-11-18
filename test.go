package main

import (
	"fmt"
)

type ITestInterface interface {
	ModMe(newStr string)
	GetMe() string
}

type TTestStruct struct {
	MyString string
}

func (m TTestStruct) ModMe(newStr string) {
	m.MyString = newStr
}

func (m TTestStruct) GetMe() string {
	return m.MyString
}

type TTestStruct2 struct {
	MyString string
}

func (m *TTestStruct2) ModMe(newStr string) {
	m.MyString = newStr
}

func (m TTestStruct2) GetMe() string {
	return m.MyString
}


func testThis() {
	wibble := "wibble"
	wobble := "wobble"
	fmt.Println("TTestStruct has method signatures:")
	fmt.Println("\tfunc (m TTestStruct) GetMe() string")
	fmt.Println("\tfunc (m TTestStruct) ModMe(newStr string)")
	fmt.Println()
	fmt.Println("create new instance of TTestStruct as ITestInterface, set MyString to", wibble)
	fmt.Println()
	fmt.Println("var t ITestInterface = TTestStruct{MyString: wibble}")
	fmt.Println()
	var t ITestInterface = TTestStruct{MyString: wibble}
	fmt.Println("confirm that GetMe() returns wibble:", t.GetMe())
	fmt.Println()
	fmt.Println("now use the interface's function to mod the string to", wobble)
	fmt.Println()
	fmt.Println("t.ModMe(wobble)")
	fmt.Println()
	t.ModMe(wobble)
	fmt.Println("now compare the result of t.GetMe() with wobble")
	fmt.Println()
	fmt.Println(t.GetMe(), "compared with", wobble)
	fmt.Println()

	fmt.Println()
	fmt.Println("TTestStruct2 has method signatures - ModMe has to be a pointer to function ref to modify data:")
	fmt.Println("\tfunc (m TTestStruct2) GetMe() string")
	fmt.Println("\tfunc (m *TTestStruct2) ModMe(newStr string)")
	fmt.Println()
	fmt.Println("create new instance of TTestStruct2 as ITestInterface, set MyString to", wibble)
	fmt.Println()
	fmt.Println("o2 := &TTestStruct2{MyString: wibble}\nvar t2 ITestInterface = o2")
	fmt.Println()
	o2 := &TTestStruct2{MyString: wibble}
	var t2 ITestInterface = o2
	fmt.Println("confirm that GetMe returns wibble:", t2.GetMe())
	fmt.Println()
	fmt.Println("now use the interface's function to mod the string to", wobble)
	fmt.Println()
	fmt.Println("t2.ModMe(wobble)")
	fmt.Println()
	t2.ModMe(wobble)
	fmt.Println("now compare the result of t2.GetMe() with wobble")
	fmt.Println()
	fmt.Println(t2.GetMe(), "compared with", wobble)
	fmt.Println()

	
}


// Loader defines a content loader
type Loader interface {
    load(src string) string
}

// WebLoader is a web content loader
type WebLoader struct{}

// load loads the content of a page
func (w *WebLoader) load(src string) string {
    return fmt.Sprintf("I loaded this page %s", src)
}

