package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	DefaultFile string
	BaseDir     string
	OutDir      string
	DefaultExt  string
}

func main() {
	// Charger la config et creer le dossier de sortie
	cfg := readConfig("config.txt")
	_ = os.MkdirAll(cfg.OutDir, 0o755)

	reader := bufio.NewReader(os.Stdin)
	currentFile := cfg.DefaultFile

	for {
		// Boucle principale du menu
		fmt.Println("\n=== Menu FileOps ===")
		fmt.Println("Fichier courant:", currentFile)
		fmt.Println("1) Choisir le fichier courant")
		fmt.Println("2) Choix A - Analyse sur fichier courant")
		fmt.Println("3) Choix B - Analyse multi-fichiers")
		fmt.Println("4) Quitter")

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
			fmt.Println("Au revoir.")
			return
		default:
			fmt.Println("Choix invalide.")
		}
	}
}

func readConfig(path string) Config {
	// Valeurs par defaut si config.txt est absent ou incomplet
	cfg := Config{
		DefaultFile: "data/input.txt",
		BaseDir:     "data",
		OutDir:      "out",
		DefaultExt:  ".txt",
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return cfg
	}
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		val := strings.TrimSpace(parts[1])
		switch key {
		case "default_file":
			cfg.DefaultFile = val
		case "base_dir":
			cfg.BaseDir = val
		case "out_dir":
			cfg.OutDir = val
		case "default_ext":
			cfg.DefaultExt = val
		}
	}
	return cfg
}

func ask(r *bufio.Reader, prompt string) string {
	// Lire une ligne saisie par l'utilisateur
	fmt.Print(prompt)
	text, _ := r.ReadString('\n')
	return strings.TrimSpace(text)
}

func isFile(path string) bool {
	// Verifier que le chemin existe et que c'est un fichier
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return !info.IsDir()
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

func readLines(path string) ([]string, error) {
	// Lire toutes les lignes d'un fichier texte
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var lines []string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	return lines, scanner.Err()
}

func wordStats(lines []string) (int, float64) {
	// Compter les mots (ignorer les nombres) et calculer la moyenne
	count := 0
	totalLen := 0
	for _, line := range lines {
		for _, tok := range strings.Fields(line) {
			w := strings.Trim(tok, ".,;:!?\"'()[]{}")
			if w == "" || isNumber(w) {
				continue
			}
			count++
			totalLen += len(w)
		}
	}
	if count == 0 {
		return 0, 0
	}
	return count, float64(totalLen) / float64(count)
}

func isNumber(s string) bool {
	// Retourner vrai si la chaine ne contient que des chiffres
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}
	return s != ""
}

func filterByKeyword(lines []string, key string) (int, []string, []string) {
	// Separer les lignes selon le mot-cle
	count := 0
	var withKey []string
	var withoutKey []string
	for _, line := range lines {
		if strings.Contains(line, key) {
			count++
			withKey = append(withKey, line)
		} else {
			withoutKey = append(withoutKey, line)
		}
	}
	return count, withKey, withoutKey
}

func writeLines(path string, lines []string) error {
	// Ecrire les lignes dans un fichier (avec saut de ligne final)
	content := strings.Join(lines, "\n")
	if content != "" {
		content += "\n"
	}
	return os.WriteFile(path, []byte(content), 0o644)
}

func toInt(s string, def int) int {
	// Convertir en int ou retourner la valeur par defaut
	if s == "" {
		return def
	}
	n, err := strconv.Atoi(s)
	if err != nil {
		return def
	}
	return n
}

func head(lines []string, n int) []string {
	// Prendre les N premieres lignes
	if n >= len(lines) {
		return lines
	}
	return lines[:n]
}

func tail(lines []string, n int) []string {
	// Prendre les N dernieres lignes
	if n >= len(lines) {
		return lines
	}
	return lines[len(lines)-n:]
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
