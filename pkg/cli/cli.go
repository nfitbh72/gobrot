package cli

import (
	"bufio"
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"
	"syscall"

	"golang.org/x/term"
)

type TCommand struct {
	Callback     any //func([]int)|func([]float64)|func([]string)
	Keyword      string
	MinArguments uint16
	MaxArguments uint16
	HelpText     []string
}

type TMultiLineCommand struct {
	firstLine string
	lastLine  string
	callback  func(string)
	helpText  []string
}

type TCLI struct {
	Commands          []TCommand
	MultiLineCommands []TMultiLineCommand
	HelpText          string
	Finish            bool
	Test              []string
	Silent            bool
}

func (clinterface *TCLI) AddCommand(keyword string, minArguments uint16, maxArguments uint16, callback any, helpText ...string) {
	//fmt.Println("start commands", clinterface.Commands)
	//fmt.Println(reflect.TypeOf(callback))
	if callback != nil {
		typeString := reflect.TypeOf(callback).String()
		allowedTypes := map[string]bool{
			"func([]int)":     true,
			"func([]string)":  true,
			"func([]float64)": true,
		}
		if !allowedTypes[typeString] {
			fmt.Println("callback type not allowed:", typeString)
		} else {
			nc := TCommand{
				Callback:     callback,
				Keyword:      keyword,
				MinArguments: minArguments,
				MaxArguments: maxArguments,
				HelpText:     helpText,
			}

			clinterface.Commands = append(clinterface.Commands, nc)
			//fmt.Println("commands", clinterface.Commands)
		}
	} else {
		fmt.Println("callback not allowed: nil")
	}
	//fmt.Println("commands", clinterface.Commands)
}

func (clinterface *TCLI) TestMe(str string) {
	clinterface.Test = append(clinterface.Test, str)
}

func (clinterface *TCLI) Help(arguments []string) {
	if len(arguments) == 0 {
		fmt.Println(clinterface.HelpText)
		fmt.Println("")
		for _, command := range clinterface.Commands {
			typeString := reflect.TypeOf(command.Callback).String()
			argumentType := "<nil>"
			switch typeString {
			case "func([]int)":
				argumentType = "<int>"
			case "func([]string)":
				argumentType = "<string>"
			case "func([]float64)":
				argumentType = "<float>"
			}
			argumentListStr := ""
			if command.MaxArguments < command.MinArguments {
				argumentListStr = "[" + argumentType + "]"
			} else {
				for i := command.MinArguments; i <= command.MaxArguments; i++ {
					argumentListStr = argumentListStr + argumentType + " "
				}
			}

			fmt.Println("\t" + command.Keyword + " " + argumentListStr)
		}
		fmt.Println()
		if len(clinterface.MultiLineCommands) > 0 {
			fmt.Println("There are also Multi-line commands available:")
			for _, mlc := range clinterface.MultiLineCommands {
				fmt.Println("\tFirst line must be: \"" + mlc.firstLine + "\" with last line as \"" + mlc.lastLine + "\"")
				for _, line := range mlc.helpText {
					fmt.Println("\t\t" + line)
				}
			}
			fmt.Println()
		}
	} else {
		for _, command := range clinterface.Commands {
			if command.Keyword == arguments[0] {
				for _, line := range command.HelpText {
					fmt.Println(line)
				}
			}
		}
	}
}

func (clinterface *TCLI) AddMultiLineCommand(firstLine string, lastLine string, callback func(string), helpText ...string) {
	mlc := TMultiLineCommand{
		firstLine: firstLine,
		lastLine:  lastLine,
		callback:  callback,
		helpText:  helpText,
	}
	clinterface.MultiLineCommands = append(clinterface.MultiLineCommands, mlc)
}

func (clinterface *TCLI) exit(arguments []string) {
	fmt.Println("Bye!")
	fmt.Println()
	clinterface.Finish = true
}

func (m *TCLI) doesCommandMatch(inputString string, commandString string) bool {
	commandLen := len(commandString)
	if inputString == commandString {
		return true
	}
	if len(inputString) > commandLen && inputString[0:commandLen+1] == commandString+" " {
		return true
	}
	return false
}

func (clinterface *TCLI) Run() {
	clinterface.Finish = false
	//add exit function
	clinterface.AddCommand("exit", 0, 0, clinterface.exit, "exit: quits this program")
	clinterface.AddCommand("help", 0, 0, clinterface.Help, "help: a little recursive, don't you think?")
	//fmt.Println("commands:", clinterface.Commands)

	reader := bufio.NewReader(os.Stdin)
	multiLine := false
	var currentMultiLineCommand TMultiLineCommand
	buffer := ""
	for !clinterface.Finish {
		if !multiLine && !clinterface.Silent {
			fmt.Print("> ")
		}
		var line string
		if clinterface.Silent {
			byteLine, _ := term.ReadPassword(int(syscall.Stdin))
			line = string(byteLine)
		} else {
			line, _ = reader.ReadString('\n')
		}
		//line, _ := reader.ReadString('\n')
		line = strings.Replace(line, "\r", "", -1)
		line = strings.Replace(line, "\n", "", -1)
		if !multiLine {
			//first check all of the multiLineCommands
			for _, mlc := range clinterface.MultiLineCommands {
				if line == mlc.firstLine {
					multiLine = true
					currentMultiLineCommand = mlc
					break
				}
			}
		}
		//check multiline again, it may have changed
		if multiLine {
			buffer = buffer + line
			if line == currentMultiLineCommand.lastLine {
				//we found the last line, lets go execute
				currentMultiLineCommand.callback(buffer)
				multiLine = false
				buffer = ""
			}
		} else {
			//fmt.Println("got command", line)
			//fmt.Println("commands", clinterface.Commands)
			for _, command := range clinterface.Commands {
				commandLen := len(command.Keyword)
				if clinterface.doesCommandMatch(line, command.Keyword) {
					//fmt.Println("got command", command.Keyword)
					//depending on the type of the argument of the callback function,
					//we format string, int or float64 differently
					switch reflect.TypeOf(command.Callback).String() {
					case "func([]string)":
						var args []string
						if len(line) > commandLen {
							args = strings.Split(line[commandLen+1:], " ")
						} else {
							args = []string{}
						}
						//fmt.Println("string function about to be called with args", args)
						command.Callback.(func([]string))(args)
					case "func([]int)":
						var args []int
						if len(line) > commandLen {
							stringArgs := strings.Split(line[commandLen+1:], " ")
							for _, stringArg := range stringArgs {
								i, err := strconv.ParseInt(stringArg, 10, 64)
								if err == nil {
									args = append(args, int(i))
								} else {
									fmt.Println("could not parse as int", stringArg)
									fmt.Printf("Error:%v\n", err)
								}
							}
						} else {
							args = []int{}
						}
						command.Callback.(func([]int))(args)
					case "func([]float64)":
						var args []float64
						if len(line) > commandLen {
							stringArgs := strings.Split(line[commandLen+1:], " ")
							for _, stringArg := range stringArgs {
								f, err := strconv.ParseFloat(stringArg, 64)
								if err == nil {
									args = append(args, f)
								} else {
									fmt.Println("could not parse as float64", stringArg)
									fmt.Printf("Error:%v\n", err)
								}
							}
						} else {
							args = []float64{}
						}
						command.Callback.(func([]float64))(args)
					}

				} //endif this keyword is the first word
			} //end foreach command
		} //endif multiline
	} //while not finish

}
