package cmd

import (
	"fmt"

	"github.com/kontrio/kappy/pkg"
	"github.com/kontrio/kappy/pkg/model"
	"github.com/spf13/cobra"
)

var stackDef *model.StackDefinition

func ArgsLoadConfigAndStackName(cmd *cobra.Command, args []string) (err error) {
	config, err = pkg.LoadConfig(&KappyFile)

	if err != nil {
		return
	}

	argsError := fmt.Errorf("Requests [stackname] argument")
	missingError := fmt.Errorf("Stack '%s' is not defined in the .kappy configuration", args[0])

	if len(args) < 1 {
		return argsError
	}

	stackDef = config.GetStackByName(args[0])

	if stackDef == nil {
		return missingError
	}

	return
}
