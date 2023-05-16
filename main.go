package main

import (
    "fmt"
)

//keeping all the main data here
type TMain struct {
    AppData *TAppData
    Commands TCommands
}
var Main TMain

func main() {

    //testThis()
    //get the app data
    Main.AppData = getAppDataInstance()

    //init the commands
    Main.Commands.initialise(Main.AppData)

    //welcome message
    fmt.Println("Welcome to GOBROT")
    fmt.Println()
    fmt.Println("Add Commands / XML here:")

    //run the CLI
    Main.Commands.run()

}