package cmd

import (
	"fmt"

	"github.com/fatih/color"
	"github.com/paulozagaloneves/kvm-compose/internal/common"
	"github.com/spf13/cobra"
)

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

// showBanner exibe o banner colorido
func showBanner() {
	color.Cyan("============================================================")
	color.New(color.FgGreen, color.Bold).Println("🖥️  kvm-compose - Gerenciador de VMs KVM via arquivo compose")
	color.New(color.FgYellow, color.Bold).Printf("📦 Versão: %s Codename: \"Gambiarra\" - Dezembro de 2025\n", common.GetVersion())
	color.Cyan("============================================================")
	fmt.Println()
}

func init() {
	// Flags globais
	rootCmd.PersistentFlags().StringVarP(&composeFile, "compose", "c", "kvm-compose.yaml", "Arquivo compose")

	// Adicionar subcomandos
	rootCmd.AddCommand(downCmd)
	/*rootCmd.AddCommand(upCmd)
	rootCmd.AddCommand(startCmd)
	rootCmd.AddCommand(stopCmd)
	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(statusCmd)*/
	rootCmd.AddCommand(versionCmd)
}

// Execute executa o comando raiz
func Execute() error {
	return rootCmd.Execute()
}
