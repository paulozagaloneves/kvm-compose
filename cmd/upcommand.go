package cmd

import (
	"fmt"
	"os"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

// Up cria e inicia todas as VMs
func (kvm *KVMCompose) Up() error {
	err := kvm.loadConfig()
	if err != nil {
		return err
	}

	color.Cyan("=== Criando todas as VMs do compose ===")

	// Baixar cada imagem base de distro apenas uma vez
	downloadedDistros := make(map[string]bool)

	createdCount := 0
	skippedCount := 0

	for _, vm := range kvm.config.VMs {
		// Baixar imagem base da distro da VM apenas se ainda não foi baixada nesta execução
		if !downloadedDistros[vm.Distro] {
			if err := kvm.downloadBaseImage(&vm); err != nil {
				color.Red("❌ Erro ao baixar imagem base para %s: %v", vm.Name, err)
				continue
			}
			downloadedDistros[vm.Distro] = true
		}
		color.White("--- Processando VM: %s ---", vm.Name)

		// Verificar se VM já existe
		if vmExists(vm.Name) {
			color.Yellow("⚠️  VM %s já existe, pulando...", vm.Name)
			skippedCount++
			fmt.Println()
			continue
		}

		// Mostrar configurações
		color.Blue("🛠️ Configurações:")
		fmt.Printf("  Distro: %s\n", vm.Distro)
		fmt.Printf("  Usuário: %s\n", vm.Username)
		fmt.Printf("  IP: %s\n", vm.Networks[0].GuestIPv4)
		fmt.Printf("  Memória: %dMB\n", vm.Memory)
		fmt.Printf("  vCPUs: %d\n", vm.VCPUs)
		fmt.Printf("  Disco: %dGB\n", vm.DiskSize)
		fmt.Printf("  Bridge: %s\n", vm.Networks[0].HostBridge)

		// Copiar imagem base para imagem da VM
		baseImagePath := kvm.getBaseImagePath(&vm)
		vmImagePath := kvm.getVMImagePath(vm.Name)

		color.Cyan("📋 Copiando: %s → %s", baseImagePath, vmImagePath)
		err := execCommand("cp", baseImagePath, vmImagePath)
		if err != nil {
			color.Red("❌ Erro ao copiar imagem para %s: %v", vm.Name, err)
			continue
		}

		// Ajustar permissões
		execCommand("sudo", "chmod", "644", vmImagePath)

		// Redimensionar imagem para o tamanho configurado
		color.Cyan("🔧 Redimensionando imagem %s para %dG...", vmImagePath, vm.DiskSize)
		if err := execCommand("qemu-img", "resize", vmImagePath, fmt.Sprintf("%dG", vm.DiskSize)); err != nil {
			color.Red("❌ Erro ao redimensionar imagem %s: %v", vm.Name, err)
			continue
		}

		// Criar arquivos cloud-init
		err = kvm.createCloudInitFiles(&vm)
		if err != nil {
			color.Red("❌ Erro ao criar arquivos cloud-init para %s: %v", vm.Name, err)
			continue
		}

		// Executar virt-install
		color.Cyan("🚀 Criando VM %s...", vm.Name)
		// Carregar OSVARIANT do INI da distro
		distroInfo, err := loadDistroInfo(vm.Distro)
		osVariant := ""
		if err == nil {
			osVariant = distroInfo.OSVariant
		} else {
			osVariant = "generic"
		}
		args := []string{
			"--name", vm.Name,
			"--memory", fmt.Sprintf("%d", vm.Memory),
			"--vcpus", fmt.Sprintf("%d", vm.VCPUs),
			"--os-variant", osVariant,
			"--virt-type", "kvm",
			"--disk", fmt.Sprintf("%s,size=%d,format=qcow2", vmImagePath, vm.DiskSize),
			"--network", fmt.Sprintf("bridge=%s,model=virtio", vm.Networks[0].HostBridge),
			"--graphics", "spice,listen=0.0.0.0",
			"--noautoconsole",
			"--import",
			"--cloud-init",
			fmt.Sprintf("user-data=%s-user-data.yaml,network-config=%s-network-config.yaml,meta-data=%s-meta-data.yaml",
				vm.Name, vm.Name, vm.Name),
		}

		if err := execCommand("virt-install", args...); err != nil {
			color.Red("❌ Falha ao criar VM %s: %v", vm.Name, err)
		} else {
			color.Green("✅ VM %s criada com sucesso!", vm.Name)
			color.Cyan("   SSH: ssh %s@%s", vm.Username, vm.Networks[0].GuestIPv4)
			createdCount++
		}

		// Limpar arquivos temporários
		cleanupCloudInitFiles(vm.Name)
		fmt.Println()
	}

	color.Cyan("=== Resumo ===")
	fmt.Printf("VMs criadas: %d\n", createdCount)
	fmt.Printf("VMs puladas (já existem): %d\n", skippedCount)
	fmt.Printf("Total de VMs no compose: %d\n", len(kvm.config.VMs))

	return nil
}

var upCmd = &cobra.Command{
	Use:   "up",
	Short: "Criar e iniciar todas as VMs do compose",
	Run: func(cmd *cobra.Command, args []string) {
		kvm := NewKVMCompose(composeFile)
		if err := kvm.Up(); err != nil {
			color.Red("Erro: %v", err)
			os.Exit(1)
		}
		// After successful Up, run List to show VMs/status
		if err := kvm.List(); err != nil {
			color.Red("Erro: %v", err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(upCmd)
}
