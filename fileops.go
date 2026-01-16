package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func runMenu(cfg Config, reader *bufio.Reader) {
	// Creer le dossier de sortie si besoin.
	_ = os.MkdirAll(cfg.OutDir, 0o755)

	currentFile := cfg.DefaultFile

	for {
		// Boucle principale du menu
		fmt.Println("\n=== Menu FileOps ===")
		fmt.Println("Fichier courant:", currentFile)
		fmt.Println("1) Choisir le fichier courant")
		fmt.Println("2) Choix A - Analyse sur fichier courant")
		fmt.Println("3) Choix B - Analyse multi-fichiers")
		fmt.Println("4) Choix C - Analyse Wikipedia")
		fmt.Println("5) Choix D - ProcessOps")
		fmt.Println("6) Quitter")

		choice := ask(reader, "Votre choix: ")
		switch choice {
		case "1":
			path := ask(reader, "Chemin du fichier (vide = default): ")
			if path == "" {
				path = cfg.DefaultFile
			}
			if !isFile(path) {
				fmt.Println("Erreur: fichier introuvable.")
				continue
			}
			currentFile = path
		case "2":
			path := ask(reader, "Chemin du fichier (vide = courant): ")
			if path == "" {
				path = currentFile
			}
			if !isFile(path) {
				fmt.Println("Erreur: fichier introuvable.")
				continue
			}
			if err := analyzeFile(path, cfg, reader); err != nil {
				fmt.Println("Erreur:", err)
			}
		case "3":
			dir := ask(reader, "Chemin du repertoire (vide = base_dir): ")
			if dir == "" {
				dir = cfg.BaseDir
			}
			if err := analyzeDir(dir, cfg); err != nil {
				fmt.Println("Erreur:", err)
			}
		case "4":
			if err := analyzeWikipedia(cfg, reader); err != nil {
				fmt.Println("Erreur:", err)
			}
		case "5":
			if err := processMenu(reader); err != nil {
				fmt.Println("Erreur:", err)
			}
		case "6":
			fmt.Println("Au revoir.")
			return
		default:
			fmt.Println("Choix invalide.")
		}
	}
}

func analyzeFile(path string, cfg Config, reader *bufio.Reader) error {
	// Lire le fichier et afficher les infos de base + stats
	lines, err := readLines(path)
	if err != nil {
		return err
	}
	info, err := os.Stat(path)
	if err != nil {
		return err
	}

	fmt.Println("\n--- Infos fichier ---")
	fmt.Println("Taille:", info.Size(), "octets")
	fmt.Println("Modification:", info.ModTime().Format(time.RFC3339))
	fmt.Println("Nb lignes:", len(lines))

	wordCount, avg := wordStats(lines)
	fmt.Println("\n--- Stats mots ---")
	fmt.Println("Nb mots (sans numeriques):", wordCount)
	fmt.Printf("Longueur moyenne: %.2f\n", avg)

	// Demander un mot-cle et creer les fichiers filtres
	keyword := ask(reader, "Mot-cle: ")
	countKey, withKey, withoutKey := filterByKeyword(lines, keyword)
	fmt.Println("Lignes contenant le mot-cle:", countKey)

	if err := writeLines(filepath.Join(cfg.OutDir, "filtered.txt"), withKey); err != nil {
		return err
	}
	if err := writeLines(filepath.Join(cfg.OutDir, "filtered_not.txt"), withoutKey); err != nil {
		return err
	}

	// Demander N pour ecrire les fichiers head/tail
	n := toInt(ask(reader, "N pour head/tail (defaut 5): "), 5)
	if n <= 0 {
		n = 5
	}
	if err := writeLines(filepath.Join(cfg.OutDir, "head.txt"), head(lines, n)); err != nil {
		return err
	}
	if err := writeLines(filepath.Join(cfg.OutDir, "tail.txt"), tail(lines, n)); err != nil {
		return err
	}

	fmt.Println("Fichiers generes dans", cfg.OutDir)
	return nil
}

func analyzeDir(dir string, cfg Config) error {
	// Analyser tous les fichiers texte du repertoire
	files, err := listTxt(dir, cfg.DefaultExt)
	if err != nil {
		return err
	}
	if len(files) == 0 {
		return fmt.Errorf("aucun fichier %s", cfg.DefaultExt)
	}

	var report strings.Builder
	report.WriteString("Rapport multi-fichiers\n\n")

	var totalLines, totalWords int
	var totalSize int64

	for _, path := range files {
		lines, err := readLines(path)
		if err != nil {
			return err
		}
		info, err := os.Stat(path)
		if err != nil {
			return err
		}
		wc, avg := wordStats(lines)
		totalLines += len(lines)
		totalWords += wc
		totalSize += info.Size()

		report.WriteString("Fichier: " + path + "\n")
		report.WriteString(fmt.Sprintf("  Taille: %d\n", info.Size()))
		report.WriteString(fmt.Sprintf("  Lignes: %d\n", len(lines)))
		report.WriteString(fmt.Sprintf("  Mots: %d (%.2f)\n\n", wc, avg))
	}

	report.WriteString("Totaux\n")
	report.WriteString(fmt.Sprintf("  Taille: %d\n", totalSize))
	report.WriteString(fmt.Sprintf("  Lignes: %d\n", totalLines))
	report.WriteString(fmt.Sprintf("  Mots: %d\n", totalWords))

	// Ecrire report index et merged dans out/
	if err := os.WriteFile(filepath.Join(cfg.OutDir, "report.txt"), []byte(report.String()), 0o644); err != nil {
		return err
	}
	if err := writeIndex(files, cfg.OutDir); err != nil {
		return err
	}
	if err := mergeFiles(cfg.BaseDir, cfg.DefaultExt, filepath.Join(cfg.OutDir, "merged.txt")); err != nil {
		return err
	}

	fmt.Println("Rapports generes dans", cfg.OutDir)
	return nil
}

func listTxt(dir, ext string) ([]string, error) {
	// Recupere tous les fichiers avec l'extension demandee
	info, err := os.Stat(dir)
	if err != nil {
		return nil, err
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("%s n'est pas un repertoire", dir)
	}

	var files []string
	err = filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if filepath.Ext(d.Name()) == ext {
			files = append(files, path)
		}
		return nil
	})
	return files, err
}

func writeIndex(files []string, outDir string) error {
	// Ecrire un index simple (chemin, taille, date)
	var b strings.Builder
	for _, path := range files {
		info, err := os.Stat(path)
		if err != nil {
			return err
		}
		b.WriteString(fmt.Sprintf("%s | %d octets | %s\n", path, info.Size(), info.ModTime().Format(time.RFC3339)))
	}
	return os.WriteFile(filepath.Join(outDir, "index.txt"), []byte(b.String()), 0o644)
}

func mergeFiles(dir, ext, outPath string) error {
	// Fusionner tous les fichiers texte dans un seul fichier
	files, err := listTxt(dir, ext)
	if err != nil {
		return err
	}
	var b strings.Builder
	for _, path := range files {
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		b.WriteString("===== " + path + " =====\n")
		b.Write(data)
		if len(data) == 0 || data[len(data)-1] != '\n' {
			b.WriteString("\n")
		}
		b.WriteString("\n")
	}
	return os.WriteFile(outPath, []byte(b.String()), 0o644)
}
