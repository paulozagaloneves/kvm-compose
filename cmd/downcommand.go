package cmd

import (
	"fmt"
	"os"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

// Down destrói todas as VMs
func (kvm *KVMCompose) Down() error {
	err := kvm.loadConfig()
	if err != nil {
		return err
	}

	color.Cyan("=== Destruindo todas as VMs do compose ===")

	destroyedCount := 0
	missingCount := 0

	for _, vm := range kvm.config.VMs {
		color.White("--- Destruindo VM: %s ---", vm.Name)

		if !vmExists(vm.Name) {
			color.Yellow("⚠️  VM %s não existe.", vm.Name)
			missingCount++
		} else {
			// Parar VM se estiver rodando
			state, _ := getVMState(vm.Name)
			if state == "running" {
				color.Cyan("Parando VM %s...", vm.Name)
				execCommand("virsh", "destroy", vm.Name)
			}

			// Remover VM
			if err := execCommand("virsh", "undefine", vm.Name); err != nil {
				color.Red("❌ Falha ao remover VM %s do libvirt: %v", vm.Name, err)
			} else {
				color.Green("✅ VM %s removida do libvirt", vm.Name)
				destroyedCount++
			}
		}

		// Remover arquivo de disco
		vmImagePath := kvm.getVMImagePath(vm.Name)
		if _, err := os.Stat(vmImagePath); err == nil {
			os.Remove(vmImagePath)
			color.Blue("💾 Arquivo de disco %s removido", vmImagePath)
		}
		fmt.Println()
	}

	color.Cyan("=== Resumo ===")
	fmt.Printf("VMs destruídas: %d\n", destroyedCount)
	fmt.Printf("VMs não existiam: %d\n", missingCount)
	fmt.Printf("Total de VMs no compose: %d\n", len(kvm.config.VMs))

	return nil
}

var downCmd = &cobra.Command{
	Use:   "down",
	Short: "Destruir todas as VMs do compose",
	Run: func(cmd *cobra.Command, args []string) {
		kvm := NewKVMCompose(composeFile)
		if err := kvm.Down(); err != nil {
			color.Red("Erro: %v", err)
			os.Exit(1)
		}
	},
}
