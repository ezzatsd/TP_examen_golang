package main

import (
	"bufio"
	"encoding/csv"
	"errors"
	"fmt"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
)

type ProcInfo struct {
	PID  string
	Name string
}

func processMenu(cfg Config, reader *bufio.Reader) error {
	// Sous-menu ProcessOps
	defTop := cfg.ProcessTopN
	if defTop <= 0 {
		defTop = 10
	}
	for {
		fmt.Println("\n=== ProcessOps ===")
		fmt.Println("1) Lister les processus (top N)")
		fmt.Println("2) Rechercher / filtrer")
		fmt.Println("3) Kill securise")
		fmt.Println("4) Retour")

		choice := ask(reader, "Votre choix: ")
		switch choice {
		case "1":
			prompt := fmt.Sprintf("N (defaut %d): ", defTop)
			n := toInt(ask(reader, prompt), defTop)
			procs, err := listProcesses(n)
			if err != nil {
				fmt.Println("Erreur:", err)
				continue
			}
			printProcs(procs)
		case "2":
			term := ask(reader, "Mot a chercher: ")
			procs, err := listProcesses(0)
			if err != nil {
				fmt.Println("Erreur:", err)
				continue
			}
			filtered := filterProcs(procs, term)
			printProcs(filtered)
		case "3":
			pid := ask(reader, "PID a tuer: ")
			if pid == "" {
				fmt.Println("PID vide.")
				continue
			}
			name := findProcName(pid)
			fmt.Printf("PID: %s | Nom: %s\n", pid, name)
			confirm := strings.ToLower(ask(reader, "Confirmer (yes/no): "))
			if confirm != "yes" {
				fmt.Println("Annule.")
				continue
			}
			if err := killProcess(pid); err != nil {
				fmt.Println("Erreur:", err)
				continue
			}
			logAudit(cfg.OutDir, "kill", fmt.Sprintf("pid=%s name=%s", pid, name))
			fmt.Println("Processus termine.")
		case "4":
			return nil
		default:
			fmt.Println("Choix invalide.")
		}
	}
}

func listProcesses(top int) ([]ProcInfo, error) {
	if runtime.GOOS == "darwin" {
		return listProcessesDarwin(top)
	}
	if runtime.GOOS == "windows" {
		return listProcessesWindows(top)
	}
	return nil, errors.New("OS non supporte")
}

func listProcessesDarwin(top int) ([]ProcInfo, error) {
	cmd := exec.Command("ps", "-Ao", "pid,comm")
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	lines := strings.Split(string(out), "\n")
	var procs []ProcInfo
	for i, line := range lines {
		if i == 0 || strings.TrimSpace(line) == "" {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}
		pid := fields[0]
		name := strings.Join(fields[1:], " ")
		procs = append(procs, ProcInfo{PID: pid, Name: name})
		if top > 0 && len(procs) >= top {
			break
		}
	}
	return procs, nil
}

func listProcessesWindows(top int) ([]ProcInfo, error) {
	cmd := exec.Command("tasklist", "/FO", "CSV")
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	r := csv.NewReader(strings.NewReader(string(out)))
	records, err := r.ReadAll()
	if err != nil {
		return nil, err
	}
	var procs []ProcInfo
	for i, rec := range records {
		if i == 0 || len(rec) < 2 {
			continue
		}
		procs = append(procs, ProcInfo{PID: rec[1], Name: rec[0]})
		if top > 0 && len(procs) >= top {
			break
		}
	}
	return procs, nil
}

func filterProcs(procs []ProcInfo, term string) []ProcInfo {
	if term == "" {
		return procs
	}
	term = strings.ToLower(term)
	var out []ProcInfo
	for _, p := range procs {
		if strings.Contains(strings.ToLower(p.Name), term) || strings.Contains(p.PID, term) {
			out = append(out, p)
		}
	}
	return out
}

func printProcs(procs []ProcInfo) {
	if len(procs) == 0 {
		fmt.Println("Aucun processus.")
		return
	}
	for _, p := range procs {
		fmt.Printf("PID: %s | %s\n", p.PID, p.Name)
	}
}

func findProcName(pid string) string {
	procs, err := listProcesses(0)
	if err != nil {
		return "unknown"
	}
	for _, p := range procs {
		if p.PID == pid {
			return p.Name
		}
	}
	return "unknown"
}

func killProcess(pid string) error {
	if _, err := strconv.Atoi(pid); err != nil {
		return fmt.Errorf("PID invalide")
	}
	if runtime.GOOS == "darwin" {
		cmd := exec.Command("kill", pid)
		return cmd.Run()
	}
	if runtime.GOOS == "windows" {
		cmd := exec.Command("taskkill", "/PID", pid, "/T")
		return cmd.Run()
	}
	return errors.New("OS non supporte")
}
