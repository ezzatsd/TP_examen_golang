package main

import (
	"bufio"
	"os"
)

func main() {
	// Point d'entree simple.
	cfg := readConfig("config.txt")
	reader := bufio.NewReader(os.Stdin)
	runMenu(cfg, reader)
}
