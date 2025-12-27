package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

// Stop para todas as VMs
func (kvm *KVMCompose) Stop() error {
	err := kvm.loadConfig()
	if err != nil {
		return err
	}

	color.Cyan("=== Parando todas as VMs do compose ===")

	stoppedCount := 0
	alreadyStoppedCount := 0
	missingCount := 0

	for _, vm := range kvm.config.VMs {
		color.White("--- Parando VM: %s ---", vm.Name)

		if !vmExists(vm.Name) {
			color.Yellow("⚠️  VM %s não existe.", vm.Name)
			missingCount++
		} else {
			state, _ := getVMState(vm.Name)
			if state == "shut off" {
				color.Red("🔴 VM %s já está parada.", vm.Name)
				alreadyStoppedCount++
			} else {
				if err := execCommand("virsh", "shutdown", vm.Name); err != nil {
					color.Red("❌ Falha ao parar VM %s: %v", vm.Name, err)
				} else {
					color.Green("✅ VM %s parada com sucesso!", vm.Name)
					stoppedCount++
				}
			}
		}
		fmt.Println()
	}

	color.Cyan("=== Resumo ===")
	fmt.Printf("VMs paradas: %d\n", stoppedCount)
	fmt.Printf("VMs já paradas: %d\n", alreadyStoppedCount)
	fmt.Printf("VMs não existem: %d\n", missingCount)
	fmt.Printf("Total de VMs no compose: %d\n", len(kvm.config.VMs))

	return nil
}

var stopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Parar todas as VMs do compose",
	Run: func(cmd *cobra.Command, args []string) {
		kvm := NewKVMCompose(composeFile)
		if err := kvm.Stop(); err != nil {
			color.Red("Erro: %v", err)
			os.Exit(1)
		}

		// Wait a few seconds to ensure VMs are stopped
		color.Yellow("Esperando alguns segundos para garantir que as VMs foram paradas...")
		time.Sleep(3 * time.Second)

		// After successful Stop, run List to show VMs/status
		if err := kvm.List(); err != nil {
			color.Red("Erro: %v", err)
			os.Exit(1)
		}
	},
}

func init() {
	// Register stop command
	rootCmd.AddCommand(stopCmd)
}
