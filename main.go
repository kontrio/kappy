package main

import "github.com/kontrio/kappy/cmd"

var version string = "dev"
var commit string = "none"
var date string = "unknown"

func main() {
	cmd.Init(&cmd.VersionInfo{
		Version: version,
		Commit:  commit,
		Date:    date,
	})
	cmd.Execute()
}
