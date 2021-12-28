package cmd

import (
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "prvbin",
	Short: "yet another privatebin cli tool",
	RunE: func(cmd *cobra.Command, args []string) error {
		return cmd.Help()
	},
}

func init() {
	rootCmd.AddCommand(createCmd)
}

func Execute() error {
	return rootCmd.Execute()
}
