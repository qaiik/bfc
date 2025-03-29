package main

import (
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"
)

// Brainfuck interpreter template (Go program that will be generated in run.go)
const runTemplate = `package main

import (
	"fmt"
	"os"
)

func interpretBrainfuck(code string) {
	const tapeSize = 30000
	tape := make([]byte, tapeSize)
	ptr := 0
	loopStack := []int{}

	for i := 0; i < len(code); i++ {
		switch code[i] {
		case '>':
			ptr = (ptr + 1) %% tapeSize
		case '<':
			ptr = (ptr - 1 + tapeSize) %% tapeSize
		case '+':
			tape[ptr]++
		case '-':
			tape[ptr]--
		case '.':
			// Print the value at tape[ptr] as a character
			fmt.Printf("%%c", tape[ptr])
		case ',':
			b := make([]byte, 1)
			if _, err := os.Stdin.Read(b); err == nil {
				tape[ptr] = b[0]
			}
		case '[':
			if tape[ptr] == 0 {
				openBrackets := 1
				for openBrackets > 0 {
					i++
					if i >= len(code) {
						fmt.Println("Error: Unmatched '['")
						return
					}
					if code[i] == '[' {
						openBrackets++
					} else if code[i] == ']' {
						openBrackets--
					}
				}
			} else {
				loopStack = append(loopStack, i)
			}
		case ']':
			if len(loopStack) == 0 {
				fmt.Println("Error: Unmatched ']'")
				return
			}
			if tape[ptr] != 0 {
				i = loopStack[len(loopStack)-1] - 1
			} else {
				loopStack = loopStack[:len(loopStack)-1]
			}
		}
	}
}

func main() {
	brainfuckCode := "%s"
	interpretBrainfuck(brainfuckCode)
}
`

func main() {
	// Validate input arguments
	if len(os.Args) < 3 {
		fmt.Println("Usage: bfc <brainfuck_file> <output_exe>")
		return
	}

	bfFile := os.Args[1]
	outputExe := os.Args[2]

	// Read Brainfuck source code
	code, err := os.ReadFile(bfFile)
	if err != nil {
		fmt.Println("Error reading bf file:", err)
		return
	}

	// Clean the Brainfuck code: Remove non-valid characters and whitespace
	validBFPattern := regexp.MustCompile(`[^\+\-\>\<\[\]\.,]`)
	cleanCode := validBFPattern.ReplaceAllString(string(code), "")

	// Escape special characters to ensure valid Go string
	escapedCode := strings.ReplaceAll(cleanCode, `"`, `\"`)             // Escape double quotes
	escapedCode = strings.ReplaceAll(escapedCode, "`", "` + \"`\" + `") // Escape backticks

	// Generate the Go file (run.go)
	runCode := fmt.Sprintf(runTemplate, escapedCode)
	runFile := "run.go"
	err = os.WriteFile(runFile, []byte(runCode), 0644)
	if err != nil {
		fmt.Println("Error writing run.go:", err)
		return
	}

	// Build run.go into the specified executable
	cmd := exec.Command("go", "build", "-o", outputExe, runFile)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err != nil {
		fmt.Println("Error building executable:", err)
		return
	}

	// Cleanup: Delete run.go after successful compilation
	err = os.Remove(runFile)
	if err != nil {
		fmt.Println("Warning: could not delete run.go:", err)
	} else {
		// fmt.Println("Deleted run.go.")
	}
}
