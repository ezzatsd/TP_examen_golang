package main

import (
	"bufio"
	"flag"
	"os"
)

func main() {
	// Point d'entree simple.
	configPath := flag.String("config", "config.json", "chemin du fichier de config")
	flag.Parse()

	cfg := readConfig(*configPath)
	reader := bufio.NewReader(os.Stdin)
	runMenu(cfg, reader)
}
