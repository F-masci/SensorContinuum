package utils

import (
	"bufio"
	"os"
	"strings"
)

// ReadInput Funzione di utilità per leggere input da stdin

func ReadInput() string {
	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')
	return strings.TrimSpace(input)
}
