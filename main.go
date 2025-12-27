package main

import (
	"os"

	"github.com/paulozagaloneves/kvm-compose/cmd"

	"github.com/fatih/color"
)

func main() {
	if err := cmd.Execute(); err != nil {
		color.Red("Erro: %v", err)
		os.Exit(1)
	}
}
