package main

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

var tui *tview.Application

func main() {
	tui = tview.NewApplication()
	tui.SetBeforeDrawFunc(func(screen tcell.Screen) bool {
		screen.Clear()
		screen.Fill(' ', tcell.StyleDefault.Background(theme.BackgroundColor))
		return false
	})
	if err := mainMenu(); err != nil {
		panic(err)
	}
}

// func main() {
// 	scanner := bufio.NewScanner(os.Stdin)
//
// 	// print help menu on start
// 	var args []string
// 	commandHelp(args)
//
// 	for {
// 		fmt.Printf("\n$ expense-tracking > ")
//
// 		scanner.Scan()
// 		input := scanner.Text()
// 		sanitizedInput := cleanTerminalInput(input)
//
// 		// if blank enter just prompt again
// 		if len(sanitizedInput) == 0 {
// 			continue
// 		}
//
// 		inputCommand := sanitizedInput[0]
// 		args := sanitizedInput[1:]
//
// 		command, validCommand := supportedCommands[inputCommand]
// 		cmdMatches := []string{}
//
// 		if validCommand {
// 			if _, err := command.callback(args); err != nil {
// 				fmt.Printf("\n\nError with command: %s\n", command.name)
// 				fmt.Printf("%s\n", err)
// 			}
// 			// if successful command run just prompt again
// 			continue
// 		} else {
// 			// try partial command match
// 			for cmd := range supportedCommands {
// 				if len(inputCommand) > 0 && len(cmd) >= len(inputCommand) && cmd[:len(inputCommand)] == inputCommand {
// 					cmdMatches = append(cmdMatches, cmd)
// 				}
// 			}
//
// 			if len(cmdMatches) == 1 {
// 				command = supportedCommands[cmdMatches[0]]
// 				if _, err := command.callback(args); err != nil {
// 					fmt.Printf("\n\nError with command: %s\n", command.name)
// 					fmt.Printf("%s\n", err)
// 				}
// 				// if successful command run just prompt again
// 				continue
// 			} else if len(cmdMatches) > 1 {
// 				fmt.Println("did you mean one of these?")
// 				for _, m := range cmdMatches {
// 					fmt.Printf("  - %s\n", m)
// 				}
// 				// give suggestion and re-prompt
// 				continue
// 			}
//
// 			// fallback: unknown command
// 			fmt.Printf("unkown command: \"%s\", please run the :help command to see valid options", inputCommand)
// 		}
// 	}
// }
