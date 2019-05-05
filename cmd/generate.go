package cmd

import (
	"github.com/spf13/cobra"
)

var generateCmd = &cobra.Command{
	Use:   "generate [stackname]",
	Short: "Generate all kubernetes resources that can be applied using kubectl apply -f",
	Args:  ArgsLoadConfigAndStackName,
	Run: func(cmd *cobra.Command, args []string) {

	},
}

func initGenerateCmd() {

}
