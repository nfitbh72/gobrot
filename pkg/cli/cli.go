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

// Unlimited as MaxArguments means the command accepts any number of arguments.
const Unlimited = ^uint16(0)

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
			if command.MaxArguments == Unlimited {
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
	clinterface.AddCommand("help", 0, 1, clinterface.Help, "help: a little recursive, don't you think?")
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
			clinterface.Dispatch(line)
		} //endif multiline
	} //while not finish

}

// Dispatch runs every command whose keyword starts the line. The arguments are
// checked against the command's min/max count and parsed for the callback's
// type; if either fails a message is printed and the callback is not called.
func (clinterface *TCLI) Dispatch(line string) {
	for _, command := range clinterface.Commands {
		if !clinterface.doesCommandMatch(line, command.Keyword) {
			continue
		}
		tokens := strings.Fields(line[len(command.Keyword):])
		if len(tokens) < int(command.MinArguments) ||
			(command.MaxArguments != Unlimited && len(tokens) > int(command.MaxArguments)) {
			fmt.Println("wrong number of arguments for", command.Keyword)
			fmt.Println("usage:", usage(command))
			continue
		}
		//depending on the type of the argument of the callback function,
		//we format string, int or float64 differently
		switch reflect.TypeOf(command.Callback).String() {
		case "func([]string)":
			command.Callback.(func([]string))(tokens)
		case "func([]int)":
			args := make([]int, 0, len(tokens))
			for _, token := range tokens {
				i, err := strconv.ParseInt(token, 10, 64)
				if err != nil {
					fmt.Println(command.Keyword+": could not parse as int:", token)
					args = nil
					break
				}
				args = append(args, int(i))
			}
			if args != nil {
				command.Callback.(func([]int))(args)
			}
		case "func([]float64)":
			args := make([]float64, 0, len(tokens))
			for _, token := range tokens {
				f, err := strconv.ParseFloat(token, 64)
				if err != nil {
					fmt.Println(command.Keyword+": could not parse as number:", token)
					args = nil
					break
				}
				args = append(args, f)
			}
			if args != nil {
				command.Callback.(func([]float64))(args)
			}
		}
	}
}

// usage describes how many arguments a command takes
func usage(command TCommand) string {
	switch {
	case command.MaxArguments == Unlimited:
		return fmt.Sprintf("%s (at least %d arguments)", command.Keyword, command.MinArguments)
	case command.MinArguments == command.MaxArguments:
		return fmt.Sprintf("%s (%d arguments)", command.Keyword, command.MinArguments)
	default:
		return fmt.Sprintf("%s (%d to %d arguments)", command.Keyword, command.MinArguments, command.MaxArguments)
	}
}
