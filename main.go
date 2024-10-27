package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"time"

	"github.com/creack/pty"
)

const MaxBufferSize = 16 // Buffer size limit

func main() {
	// Set up a new Bash command in a pseudo-terminal (pty)
	cmd := exec.Command("bash")
	ptmx, err := pty.Start(cmd)
	if err != nil {
		log.Fatalf("Failed to start pty: %v", err)
	}
	defer func() { _ = ptmx.Close(); _ = cmd.Process.Kill() }()

	// Goroutine to capture and send user input to pty
	go func() {
		reader := bufio.NewReader(os.Stdin)
		for {
			// Read input character by character
			r, _, err := reader.ReadRune()
			if err != nil {
				log.Printf("Error reading input: %v", err)
				return
			}

			// Write character to pty and display it on the console
			_, _ = ptmx.Write([]byte(string(r)))
			fmt.Print(string(r)) // Echo the character to the console
		}
	}()

	// Buffer to store lines and manage output display
	buffer := [][]rune{}
	reader := bufio.NewReader(ptmx)
	go func() {
		line := []rune{}
		buffer = append(buffer, line)

		for {
			r, _, err := reader.ReadRune()
			if err != nil {
				if err == io.EOF {
					return
				}
				log.Fatalf("Error reading from pty: %v", err)
			}

			line = append(line, r)
			buffer[len(buffer)-1] = line

			if r == '\n' {
				if len(buffer) > MaxBufferSize {
					buffer = buffer[1:]
				}
				line = []rune{}
				buffer = append(buffer, line)
			}
		}
	}()

	// Render output to console every 100ms
	for range time.Tick(100 * time.Millisecond) {
		fmt.Print("\033[H\033[2J") // Clear screen
		for _, line := range buffer {
			fmt.Print(string(line))
		}
	}
}
