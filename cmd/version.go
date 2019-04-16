package cmd

import (
	"github.com/apex/log"
	"github.com/spf13/cobra"
)

var versionInfo *VersionInfo

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print kappy version information",
	Run: func(cmd *cobra.Command, args []string) {
		if versionInfo == nil {
			log.Infof("Version: unavailable")
			return
		}

		log.Infof("Version: %s (%s) - %s", versionInfo.Version, versionInfo.Commit, versionInfo.Date)
	},
}

func initVersionCmd(version *VersionInfo) {
	versionInfo = version
}
