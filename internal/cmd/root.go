package cmd

import (
	"github.com/deahtstroke/tast/internal/cmd/edit"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(edit.NewEditCommand())
}

var rootCmd = &cobra.Command{
	Use:   "tast",
	Short: "TOML parser tool",
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() error {
	rootCmd.CompletionOptions.SetDefaultShellCompDirective(cobra.ShellCompDirectiveNoFileComp)
	if err := rootCmd.Execute(); err != nil {
		return err
	}
	return nil
}
