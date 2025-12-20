package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

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

func funcName(imagePath string, url string) (error, bool) {
	if _, err := os.Stat(imagePath); os.IsNotExist(err) {
		color.Cyan("📥 Baixando imagem base do Debian 13...")
		color.Cyan("📂 Salvando em: %s", imagePath)

		return execCommand("wget", "-O", imagePath, url), true
	}
	return nil, false
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

var (
	composeFile string
	rootCmd     = &cobra.Command{
		Use:   "kvm-compose",
		Short: "Gerenciador de VMs KVM via arquivo compose",
		Long:  `kvm-compose é uma ferramenta para gerenciar múltiplas VMs KVM usando um arquivo de configuração YAML estilo Docker Compose.`,
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			showBanner()
		},
	}
)

func init() {
	// Flags globais
	rootCmd.PersistentFlags().StringVarP(&composeFile, "compose", "c", "kvm-compose.yaml", "Arquivo compose")

	// Adicionar subcomandos
	rootCmd.AddCommand(downCmd)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		color.Red("Erro: %v", err)
		os.Exit(1)
	}
}
