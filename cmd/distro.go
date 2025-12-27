package cmd

import (
	"fmt"
	"os"
	"os/user"
	"path/filepath"

	"gopkg.in/ini.v1"
)

// DistroInfo representa informações de uma distribuição
type DistroInfo struct {
	URL       string
	Source    string
	OSVariant string
}

// Função auxiliar para procurar template
func findTemplate(filename string) (string, error) {
	// 1. Procurar na pasta local templates/
	local := filepath.Join("templates", filename)
	if _, err := os.Stat(local); err == nil {
		return local, nil
	}
	// 2. Procurar em ~/.config/kvm-compose/templates/
	usr, _ := user.Current()
	if usr != nil {
		home := filepath.Join(usr.HomeDir, ".config", "kvm-compose", "templates", filename)
		if _, err := os.Stat(home); err == nil {
			return home, nil
		}
	}
	return "", fmt.Errorf("template %s não encontrado", filename)
}

// loadDistroInfo lê o ficheiro templates/<distro>.ini e retorna as informações da distro
func loadDistroInfo(distro string) (*DistroInfo, error) {
	iniPath, err := findTemplate(fmt.Sprintf("%s.ini", distro)) //filepath.Join("templates", fmt.Sprintf("%s.ini", distro))
	cfg, err := ini.Load(iniPath)
	if err != nil {
		return nil, fmt.Errorf("erro ao carregar INI da distro: %v", err)
	}
	return &DistroInfo{
		URL:       cfg.Section("").Key("URL").String(),
		Source:    cfg.Section("").Key("SOURCE").String(),
		OSVariant: cfg.Section("").Key("OSVARIANT").String(),
	}, nil
}
