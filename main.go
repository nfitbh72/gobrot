package main

import (
	"fmt"
)

// keeping all the main data here
type TMain struct {
	AppData  *TAppData
	Commands *TCommands
}

func main() {
	m := TMain{
		AppData:  getAppDataInstance(),
		Commands: &TCommands{},
	}

	//init the commands
	m.Commands.initialise(m.AppData)

	//welcome message
	fmt.Println("Welcome to GOBROT")
	fmt.Println()
	fmt.Println("Add Commands / XML here:")

	//run the CLI
	m.Commands.run()
}
