package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/fatih/color"
)

// readSSHKey lê o conteúdo da chave SSH
func readSSHKey(keyFile string) (string, error) {
	expandedPath := expandPath(keyFile)
	content, err := os.ReadFile(expandedPath)
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

// getVMByName busca uma VM pelo nome
func (kvm *KVMCompose) getVMByName(name string) (*VM, error) {
	for _, vm := range kvm.config.VMs {
		if vm.Name == name {
			return &vm, nil
		}
	}
	return nil, fmt.Errorf("VM '%s' não encontrada", name)
}

// vmExists verifica se uma VM existe no libvirt
func vmExists(name string) bool {
	_, err := execCommandOutput("virsh", "dominfo", name)
	return err == nil
}

// downloadBaseImage baixa a imagem base da distro da VM se não existir
func (kvm *KVMCompose) downloadBaseImage(vm *VM) error {
	// Carregar informações da distro
	distroInfo, err := loadDistroInfo(vm.Distro)
	if err != nil {
		return fmt.Errorf("erro ao obter informações da distro '%s': %v", vm.Distro, err)
	}

	// Criar diretórios se não existirem
	upstreamDir := expandPath(kvm.appConfig.Images.PathUpstreamImages)
	err = os.MkdirAll(upstreamDir, 0755)
	if err != nil {
		color.Yellow("⚠️  Erro ao criar diretório %s: %v", upstreamDir, err)
		upstreamDir = "." // Fallback para diretório atual
	}

	imageName := distroInfo.Source
	imagePath := filepath.Join(upstreamDir, imageName)
	url := distroInfo.URL
	if _, err := os.Stat(imagePath); os.IsNotExist(err) {
		color.Cyan("📥 Baixando imagem base da distro %s...", vm.Distro)
		color.Cyan("📂 Salvando em: %s", imagePath)
		if err := execCommand("wget", "-O", imagePath, url); err != nil {
			return fmt.Errorf("erro ao baixar imagem: %v", err)
		}
	} else {
		color.Green("✅ Imagem base já existe: %s", imagePath)
	}
	return nil
}

// getBaseImagePath retorna o caminho da imagem base da distro da VM
func (kvm *KVMCompose) getBaseImagePath(vm *VM) string {
	upstreamDir := expandPath(kvm.appConfig.Images.PathUpstreamImages)
	distroInfo, err := loadDistroInfo(vm.Distro)
	imageName := ""
	if err == nil {
		imageName = distroInfo.Source
	} else {
		imageName = "base.qcow2" // fallback
	}
	return filepath.Join(upstreamDir, imageName)
}

// getVMImagePath retorna o caminho para a imagem da VM
func (kvm *KVMCompose) getVMImagePath(vmName string) string {
	vmImagesDir := expandPath(kvm.appConfig.Images.PathVMImages)
	err := os.MkdirAll(vmImagesDir, 0755)
	if err != nil {
		color.Yellow("⚠️  Erro ao criar diretório %s: %v", vmImagesDir, err)
		return vmName + ".qcow2" // Fallback para diretório atual
	}
	return filepath.Join(vmImagesDir, vmName+".qcow2")
}
