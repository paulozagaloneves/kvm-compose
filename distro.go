package main

import (
	"fmt"
	"path/filepath"

	"gopkg.in/ini.v1"
)

// DistroInfo representa informações de uma distribuição
type DistroInfo struct {
	URL       string
	Source    string
	OSVariant string
}

// loadDistroInfo lê o ficheiro templates/<distro>.ini e retorna as informações da distro
func loadDistroInfo(distro string) (*DistroInfo, error) {
	iniPath := filepath.Join("templates", fmt.Sprintf("%s.ini", distro))
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
