package main

import (
	"kvm-compose/cmd"
	"os"

	"github.com/fatih/color"
)

func main() {
	if err := cmd.Execute(); err != nil {
		color.Red("Erro: %v", err)
		os.Exit(1)
	}
}
