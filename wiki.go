package main

import (
	"bufio"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

func analyzeWikipedia(cfg Config, reader *bufio.Reader) error {
	// Telecharger une page Wikipedia et appliquer des traitements simples
	article := ask(reader, "Nom de l'article (ex: Go_(langage)): ")
	if article == "" {
		return fmt.Errorf("article vide")
	}

	lines, err := fetchWikiParagraphs(article)
	if err != nil {
		return err
	}
	if len(lines) == 0 {
		return fmt.Errorf("aucun paragraphe trouve")
	}

	wordCount, avg := wordStats(lines)
	fmt.Println("Nb mots (sans numeriques):", wordCount)
	fmt.Printf("Longueur moyenne: %.2f\n", avg)

	keyword := ask(reader, "Mot-cle pour filtrer: ")
	countKey, withKey, _ := filterByKeyword(lines, keyword)
	fmt.Println("Lignes contenant le mot-cle:", countKey)

	safeName := strings.NewReplacer(" ", "_", "/", "_", ":", "_", "?", "_").Replace(article)
	outPath := filepath.Join(cfg.OutDir, "wiki_"+safeName+".txt")

	var b strings.Builder
	b.WriteString("Article: " + article + "\n")
	b.WriteString(fmt.Sprintf("Mots: %d | Moyenne: %.2f\n\n", wordCount, avg))
	if keyword != "" {
		b.WriteString("Mot-cle: " + keyword + "\n")
		b.WriteString(fmt.Sprintf("Lignes contenant le mot-cle: %d\n\n", countKey))
	}
	b.WriteString(strings.Join(withKey, "\n"))
	if err := os.WriteFile(outPath, []byte(b.String()), 0o644); err != nil {
		return err
	}

	fmt.Println("Fichier cree:", outPath)
	return nil
}

func fetchWikiParagraphs(article string) ([]string, error) {
	// Recuperer le texte des paragraphes Wikipedia
	wikiURL := "https://fr.wikipedia.org/wiki/" + url.PathEscape(article)
	req, err := http.NewRequest(http.MethodGet, wikiURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; TP-Examen-Golang/1.0)")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("http status %s", resp.Status)
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, err
	}

	var lines []string
	doc.Find("p").Each(func(i int, s *goquery.Selection) {
		text := strings.TrimSpace(s.Text())
		if text != "" {
			lines = append(lines, text)
		}
	})
	return lines, nil
}
