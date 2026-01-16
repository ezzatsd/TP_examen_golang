package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	DefaultFile string `json:"default_file"`
	BaseDir     string `json:"base_dir"`
	OutDir      string `json:"out_dir"`
	DefaultExt  string `json:"default_ext"`
	WikiLang    string `json:"wiki_lang"`
	ProcessTopN int    `json:"process_top_n"`
}

func defaultConfig() Config {
	return Config{
		DefaultFile: "data/input.txt",
		BaseDir:     "data",
		OutDir:      "out",
		DefaultExt:  ".txt",
		WikiLang:    "fr",
		ProcessTopN: 10,
	}
}

func readConfig(path string) Config {
	// Valeurs par defaut si config.json est absent ou incomplet
	cfg := defaultConfig()
	data, err := os.ReadFile(path)
	if err != nil {
		return cfg
	}

	var userCfg Config
	if err := json.Unmarshal(data, &userCfg); err != nil {
		return cfg
	}

	if userCfg.DefaultFile != "" {
		cfg.DefaultFile = userCfg.DefaultFile
	}
	if userCfg.BaseDir != "" {
		cfg.BaseDir = userCfg.BaseDir
	}
	if userCfg.OutDir != "" {
		cfg.OutDir = userCfg.OutDir
	}
	if userCfg.DefaultExt != "" {
		cfg.DefaultExt = userCfg.DefaultExt
	}
	if userCfg.WikiLang != "" {
		cfg.WikiLang = userCfg.WikiLang
	}
	if userCfg.ProcessTopN != 0 {
		cfg.ProcessTopN = userCfg.ProcessTopN
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

func logAudit(outDir, action, details string) {
	// Journaliser les actions sensibles dans out/audit.log
	_ = os.MkdirAll(outDir, 0o755)
	path := outDir + "/audit.log"
	ts := time.Now().Format(time.RFC3339)
	line := fmt.Sprintf("%s | %s | %s\n", ts, action, details)
	f, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return
	}
	defer f.Close()
	_, _ = f.WriteString(line)
}
