package cmd

import (
	"fmt"
	"os"

	"github.com/apex/log"
	"github.com/apex/log/handlers/cli"
	"github.com/spf13/cobra"
)

var Verbose bool
var KappyFile string = ""

var rootCmd = &cobra.Command{
	Use:   "kappy",
	Short: "kappy is an opinionated kubectl wrapper",
	Long:  "kappy helps make building, configuration and deployment of your microservices to Kubernetes easier",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		if Verbose {
			log.SetLevel(log.DebugLevel)
		}
	},
}

func Init() {
	rootCmd.PersistentFlags().BoolVarP(&Verbose, "verbose", "v", false, "verbose output")
	rootCmd.PersistentFlags().StringVarP(&KappyFile, "file", "f", "", "Alternative .kappy.yml")
	rootCmd.AddCommand(deployCmd)
	rootCmd.AddCommand(buildCmd)

	initBuildCmd()
	initDeployCmd()
}

func Execute() {
	log.SetHandler(cli.Default)
	log.SetLevel(log.InfoLevel)
	if err := rootCmd.Execute(); err != nil {
		fmt.Errorf("Something went wrong: %s", err)
		os.Exit(1)
	}
}
