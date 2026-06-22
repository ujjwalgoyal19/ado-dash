package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// version is set at build time via goreleaser ldflags.
var version = "dev"

func main() {
	var cfgFile string

	root := &cobra.Command{
		Use:     "ado-dash",
		Short:   "Terminal dashboard for Azure DevOps",
		Version: version,
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("ado-dash: TUI not yet implemented. Run --help for options.")
			return nil
		},
	}

	root.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default: ~/.config/ado-dash/config.yml)")
	_ = cfgFile // consumed in later slices

	root.AddCommand(versionCmd(), doctorCmd())

	if err := root.Execute(); err != nil {
		os.Exit(1)
	}
}

func versionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print version information",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("ado-dash %s\n", version)
		},
	}
}

func doctorCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "doctor",
		Short: "Check runtime dependencies",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("doctor: not yet implemented")
		},
	}
}
