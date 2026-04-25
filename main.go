package main

import (
	"context"
	"fmt"
	"os"

	"github.com/charmbracelet/fang"
	"github.com/ctrl-vfr/persona/cmd"
	"github.com/ctrl-vfr/persona/internal/storage"
)

func main() {
	mgr, err := storage.NewManager()
	if err != nil {
		fmt.Fprintf(os.Stderr, "persona: init storage manager: %v\n", err)
		os.Exit(1)
	}
	if err := mgr.InitializeStructure(); err != nil {
		fmt.Fprintf(os.Stderr, "persona: init storage layout: %v\n", err)
		os.Exit(1)
	}
	cmd.Setup(mgr)

	rootCmd := cmd.GetRootCmd()
	if err := fang.Execute(context.Background(), rootCmd); err != nil {
		os.Exit(1)
	}
}
