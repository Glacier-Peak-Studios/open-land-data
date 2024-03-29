package utils

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"golang.org/x/term"
)

// CheckRequiredFlags checks if all flags requred to be passed are passed
func CheckRequiredFlags(required ...string) {
	seen := make(map[string]bool)
	flag.Visit(func(f *flag.Flag) { seen[f.Name] = true })
	for _, req := range required {
		if !seen[req] {
			// or possibly use `log.Fatalf` instead of:
			fmt.Fprintf(os.Stderr, "Missing required argument/flag: -%s \nSee -h for usage\n", req)
			os.Exit(2) // the same exit code flag.Parse uses
		}
	}
}

// CliUsage prints out a formatted usage message for a
// program with a given title and description
func CliUsage(title, description string) func() {
	return func() {
		printString := wrapString(description, 80)
		fmt.Fprintf(os.Stderr, "\n\033[1m%s\033[0m: %s\n\n", title, printString)
		fmt.Fprintf(os.Stderr, "\033[4mUsage of %s:\033[0m\n", title)
		flag.PrintDefaults()
	}
}

func wrapString(s string, width int) string {
	cols, _, _ := term.GetSize(int(os.Stdout.Fd()))
	if cols < width {
		width = cols
	}

	words := strings.Fields(s)
	var result string
	line := ""
	for _, word := range words {
		if len(line)+len(word)+1 > width {
			result += line + "\n"
			line = ""
		}
		if line == "" {
			line = word
		} else {
			line += " " + word
		}
	}
	if line != "" {
		result += line
	}
	return result
}
