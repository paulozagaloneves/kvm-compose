package cmd

import (
	"fmt"

	"kvm-compose/internal/common"

	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Exibe a versão do kvm-compose",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Print(common.GetVersion())
	},
}
