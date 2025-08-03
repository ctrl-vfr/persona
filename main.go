package main

import (
	"context"
	"os"

	"github.com/charmbracelet/fang"
	"github.com/ctrl-vfr/persona/cmd"
)

func main() {
	rootCmd := cmd.GetRootCmd()
	if err := fang.Execute(context.Background(), rootCmd); err != nil {
		os.Exit(1)
	}
}
