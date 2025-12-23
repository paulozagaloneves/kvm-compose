package cmd

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

// sshCmd connects to a VM via SSH using the VM username and guest IP.
var sshCmd = &cobra.Command{
	Use:   "ssh <vm-name> [ssh-args...]",
	Short: "Open SSH to a VM using its configured user and IP",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]
		extra := []string{}
		if len(args) > 1 {
			extra = args[1:]
		}

		kvm := NewKVMCompose(composeFile)
		if err := kvm.loadConfig(); err != nil {
			color.Red("Erro: %v", err)
			os.Exit(1)
		}

		vm, err := kvm.getVMByName(name)
		if err != nil {
			color.Red("Erro: %v", err)
			os.Exit(1)
		}

		user := vm.Username
		if user == "" {
			user = kvm.appConfig.Main.Username
		}

		if len(vm.Networks) == 0 || vm.Networks[0].GuestIPv4 == "" {
			color.Red("Erro: IP não disponível para VM %s", name)
			os.Exit(1)
		}

		target := fmt.Sprintf("%s@%s", user, vm.Networks[0].GuestIPv4)
		color.Cyan("🔗 SSH to %s", target)

		argsToPass := append([]string{target}, extra...)
		if err := runInteractiveCommand("ssh", argsToPass...); err != nil {
			color.Red("Erro ao executar ssh: %v", err)
			os.Exit(1)
		}
	},
}

func init() {
	// Register ssh command
	rootCmd.AddCommand(sshCmd)
}

// runInteractiveCommand runs a command attaching the current terminal to it.
// If stdin is a terminal, request a pseudo-tty from ssh with `-t`.
func runInteractiveCommand(name string, args ...string) error {
	// If the calling stdin is a terminal, ask ssh to allocate a tty.
	if term.IsTerminal(int(os.Stdin.Fd())) {
		args = append([]string{"-t"}, args...)
	}

	cmd := exec.Command(name, args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}
