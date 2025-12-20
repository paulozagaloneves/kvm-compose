package main

import (
	"fmt"
	"os"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

// Start inicia todas as VMs
func (kvm *KVMCompose) Start() error {
	err := kvm.loadConfig()
	if err != nil {
		return err
	}

	color.Cyan("=== Iniciando todas as VMs do compose ===")

	startedCount := 0
	runningCount := 0
	missingCount := 0

	for _, vm := range kvm.config.VMs {
		color.White("--- Iniciando VM: %s ---", vm.Name)

		if !vmExists(vm.Name) {
			color.Yellow("⚠️  VM %s não existe. Use 'up' para criar.", vm.Name)
			missingCount++
		} else {
			state, _ := getVMState(vm.Name)
			if state == "running" {
				color.Green("🟢 VM %s já está em execução.", vm.Name)
				runningCount++
			} else {
				if err := execCommand("virsh", "start", vm.Name); err != nil {
					color.Red("❌ Falha ao iniciar VM %s: %v", vm.Name, err)
				} else {
					color.Green("✅ VM %s iniciada com sucesso!", vm.Name)
					startedCount++
				}
			}
		}
		fmt.Println()
	}

	color.Cyan("=== Resumo ===")
	fmt.Printf("VMs iniciadas: %d\n", startedCount)
	fmt.Printf("VMs já rodando: %d\n", runningCount)
	fmt.Printf("VMs não existem: %d\n", missingCount)
	fmt.Printf("Total de VMs no compose: %d\n", len(kvm.config.VMs))

	return nil
}

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Iniciar todas as VMs do compose",
	Run: func(cmd *cobra.Command, args []string) {
		kvm := NewKVMCompose(composeFile)
		if err := kvm.Start(); err != nil {
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
	// Register start command
	rootCmd.AddCommand(startCmd)
}
