package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func secureMenu(cfg Config, reader *bufio.Reader) error {
	// Sous-menu SecureOps (lock / unlock)
	for {
		fmt.Println("\n=== SecureOps ===")
		fmt.Println("1) Verrouiller un fichier")
		fmt.Println("2) Deverrouiller un fichier")
		fmt.Println("3) Retour")

		choice := ask(reader, "Votre choix: ")
		switch choice {
		case "1":
			path := ask(reader, "Chemin du fichier a verrouiller: ")
			if path == "" {
				fmt.Println("Chemin vide.")
				continue
			}
			if !isFile(path) {
				fmt.Println("Fichier introuvable.")
				continue
			}
			confirm := strings.ToLower(ask(reader, "Confirmer (yes/no): "))
			if confirm != "yes" {
				fmt.Println("Annule.")
				continue
			}
			lockPath := lockFilePath(cfg.OutDir, path)
			if _, err := os.Stat(lockPath); err == nil {
				fmt.Println("Deja verrouille:", lockPath)
				continue
			}
			if err := os.WriteFile(lockPath, []byte(time.Now().Format(time.RFC3339)), 0o644); err != nil {
				fmt.Println("Erreur:", err)
				continue
			}
			logAudit(cfg.OutDir, "lock", fmt.Sprintf("file=%s lock=%s", path, lockPath))
			fmt.Println("Verrou cree:", lockPath)
		case "2":
			path := ask(reader, "Chemin du fichier a deverrouiller: ")
			if path == "" {
				fmt.Println("Chemin vide.")
				continue
			}
			confirm := strings.ToLower(ask(reader, "Confirmer (yes/no): "))
			if confirm != "yes" {
				fmt.Println("Annule.")
				continue
			}
			lockPath := lockFilePath(cfg.OutDir, path)
			if _, err := os.Stat(lockPath); err != nil {
				fmt.Println("Pas de verrou:", lockPath)
				continue
			}
			if err := os.Remove(lockPath); err != nil {
				fmt.Println("Erreur:", err)
				continue
			}
			logAudit(cfg.OutDir, "unlock", fmt.Sprintf("file=%s lock=%s", path, lockPath))
			fmt.Println("Verrou supprime:", lockPath)
		case "3":
			return nil
		default:
			fmt.Println("Choix invalide.")
		}
	}
}

func lockFilePath(outDir, targetPath string) string {
	base := filepath.Base(targetPath)
	return filepath.Join(outDir, base+".lock")
}
