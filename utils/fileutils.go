package utils

import (
	"bufio"
	"io"
	"os"
)

// Exists reports whether the named file or directory exists.
func Exists(name string) bool {
	if _, err := os.Stat(name); err != nil {
		if os.IsNotExist(err) {
			return false
		}
	}
	return true
}

// LineCounter counts the number of lines of the file named fileName
func LineCounter(fileName string) (int, error) {
	file, err := os.Open(fileName)
	fileScanner := bufio.NewScanner(file)
	lineCount := 0
	if err != nil && err != io.EOF {
		return lineCount, err
	}
	for fileScanner.Scan() {
		lineCount++
	}
	return lineCount, err
}
