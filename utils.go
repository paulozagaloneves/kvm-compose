package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/fatih/color"
)

// readSSHKey lê o conteúdo da chave SSH
func readSSHKey(keyFile string) (string, error) {
	expandedPath := expandPath(keyFile)
	content, err := ioutil.ReadFile(expandedPath)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(content)), nil
}

// getVMState retorna o estado atual de uma VM
func getVMState(name string) (string, error) {
	if !vmExists(name) {
		return "not created", nil
	}
	return execCommandOutput("virsh", "domstate", name)
}

// expandPath expande o ~ no caminho para o home directory
func expandPath(path string) string {
	if strings.HasPrefix(path, "~/") {
		home, _ := os.UserHomeDir()
		return filepath.Join(home, path[2:])
	}
	return path
}

// showBanner exibe o banner colorido
func showBanner() {
	color.Cyan("============================================================")
	color.New(color.FgGreen, color.Bold).Println("🖥️  kvm-compose - Gerenciador de VMs KVM via arquivo compose")
	color.New(color.FgYellow, color.Bold).Println("📦 Versão: 0.1.0 Codename: \"Gambiarra\" - Dezembro de 2025")
	color.Cyan("============================================================")
	fmt.Println()
}
