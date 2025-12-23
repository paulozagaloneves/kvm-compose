package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

// List lista todas as VMs com seus status
func (kvm *KVMCompose) List() error {
	err := kvm.loadConfig()
	if err != nil {
		return err
	}

	fmt.Println()
	color.Cyan("=== VMs disponíveis no %s ===", kvm.composeFile)

	color.New(color.FgGreen, color.Bold).Printf("%-15s %-15s %-10s %-6s %-8s %-16s %-16s %-18s\n",
		strings.Repeat("-", 15),
		strings.Repeat("-", 15),
		strings.Repeat("-", 10),
		strings.Repeat("-", 6),
		strings.Repeat("-", 8),
		strings.Repeat("-", 16),
		strings.Repeat("-", 16),
		strings.Repeat("-", 18))
	color.New(color.FgGreen, color.Bold).Printf("%-15s %-15s %-10s %-6s %-8s %-16s %-16s %-18s\n", "Nome", "Distro", "Memória", "vCPUs", "Disco", "Username", "IP", "Status")
	color.New(color.FgGreen, color.Bold).Printf("%-15s %-15s %-10s %-6s %-8s %-16s %-16s %-18s\n",
		strings.Repeat("-", 15),
		strings.Repeat("-", 15),
		strings.Repeat("-", 10),
		strings.Repeat("-", 6),
		strings.Repeat("-", 8),
		strings.Repeat("-", 16),
		strings.Repeat("-", 16),
		strings.Repeat("-", 18))

	for _, vm := range kvm.config.VMs {
		// Aplicar valores padrão
		memory := vm.Memory
		if memory == 0 {
			memory = 4096
		}
		vcpus := vm.VCPUs
		if vcpus == 0 {
			vcpus = 4
		}
		diskSize := vm.DiskSize
		if diskSize == 0 {
			diskSize = 20
		}

		user := vm.Username
		if user == "" {
			user = kvm.appConfig.Main.Username
		}

		ip := "N/A"
		if len(vm.Networks) > 0 {
			ip = vm.Networks[0].GuestIPv4
		}

		// Verificar status
		statusText := "⚪ not created"
		if vmExists(vm.Name) {
			state, _ := getVMState(vm.Name)
			switch state {
			case "running":
				statusText = "🟢 running"
			case "shut off":
				statusText = "🔴 stopped"
			case "paused":
				statusText = "🟡 paused"
			case "suspended":
				statusText = "🟠 suspended"
			default:
				statusText = "❓ " + state
			}
		}

		// Formatar dados com larguras fixas
		memoryStr := fmt.Sprintf("%dMB", memory)
		vcpusStr := fmt.Sprintf("%d", vcpus)
		diskStr := fmt.Sprintf("%dGB", diskSize)

		fmt.Printf("%-15s %-15s %-10s %-6s %-8s %-16s %-16s %-18s\n",
			vm.Name, vm.Distro, memoryStr, vcpusStr, diskStr, user, ip, statusText)
	}

	return nil
}

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "Listar VMs disponíveis no compose",
	Run: func(cmd *cobra.Command, args []string) {
		kvm := NewKVMCompose(composeFile)
		if err := kvm.List(); err != nil {
			color.Red("Erro: %v", err)
			os.Exit(1)
		}
	},
}

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Mostrar o status das VMs do compose",
	Run: func(cmd *cobra.Command, args []string) {
		kvm := NewKVMCompose(composeFile)
		if err := kvm.List(); err != nil {
			color.Red("Erro: %v", err)
			os.Exit(1)
		}
	},
}

func init() {
	// Registrar comandos list e status
	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(statusCmd)
}
