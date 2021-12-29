package cmd

import (
	"github.com/spf13/cobra"
)

var rootFlags struct {
	url string
}

var rootCmd = &cobra.Command{
	Use:              "prvbin",
	Short:            "yet another privatebin cli tool",
	TraverseChildren: true,
	RunE:             create,
}

func init() {
	rootCmd.AddCommand(createCmd)
	rootCmd.Flags().StringVarP(&rootFlags.url, "url", "u", "", "url of the privatebin server")
}

func Execute() error {
	return rootCmd.Execute()
}
