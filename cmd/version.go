package main

import (
	"github.com/apex/log"
	"github.com/spf13/cobra"
)

var version string = "dev"
var commit string = "none"
var date string = "unknown"

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print kappy version information",
	Run: func(cmd *cobra.Command, args []string) {
		log.Infof("Version: %s (%s) - %s", version, commit, date)
	},
}
