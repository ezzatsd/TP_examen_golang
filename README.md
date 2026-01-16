# TP_examen_golang

Bonjour, voici mon projet en Go pour le controle.

## Ce que fait l'application
- **FileOps (niveau 10)** : lire un fichier texte, faire des stats, filtrer, head/tail, rapport.
- **WebOps (niveau 13)** : lire un article Wikipedia et faire des stats.
- **ProcOps (niveau 16)** : lister les processus, filtrer, et tuer un processus (avec confirmation).

## Comment lancer
1) Ouvrir le terminal et aller dans le dossier
```bash
cd TP_examen_golang
```

2) Initialiser le module Go (si besoin)
```bash
go mod init TP_examen_golang
```

3) Installer la dependance Wikipedia
```bash
go get github.com/PuerkitoBio/goquery
```

4) Lancer le programme
```bash
go run .
```

## Fichiers importants
- `config.txt` : configuration de base
- `data/` : fichiers d'entree
- `out/` : fichiers generes par le programme

## Utilisation rapide
Quand le menu s'affiche :
- **Choix 2** = analyser un fichier
- **Choix 3** = analyser tous les .txt d'un dossier
- **Choix 4** = analyser Wikipedia
- **Choix 5** = ProcessOps (processus)

Exemple simple :
1) Choix 2
2) Mot-cle : `Go`
3) N : `3`
4) Regarder le dossier `out/`

## Tests manuels (exemples)
### Test FileOps (Choix A)
1) Choix `2`
2) Entrer un mot-cle, par ex `Go`
3) N = `3`
4) Verifier avec :
```bash
ls out
```

### Test WebOps (Wikipedia)
1) Choix `4`
2) Article : `Go_(langage)`
3) Mot-cle : `langage`
4) Verifier avec :
```bash
ls out
```

### Test ProcOps (ProcessOps)
1) Ouvrir Calculator (Calculatrice)
2) Choix `5` -> `2` (filtrer)
3) Mot a chercher : `Calculator`
4) Noter le PID (le numero)
5) Choix `3` (kill securise)
6) Entrer le PID puis `yes`
7) Calculator se ferme

## Notes
- `go.sum` est cree automatiquement par Go (normal).
- Si `gofmt` ne montre rien, c'est normal.
