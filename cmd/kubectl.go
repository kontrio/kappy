package cmd

import (
	"io"
	"os"
	"os/exec"

	"github.com/apex/log"
	"github.com/kontr/kappy/pkg"
	"github.com/kontr/kappy/pkg/kubernetes"
	"github.com/kr/pty"
	"github.com/spf13/cobra"
)

var Stack string

var kubectlCmd = &cobra.Command{
	Use:  "kubectl",
	Args: cobra.ArbitraryArgs,
	Run: func(cmd *cobra.Command, args []string) {
		log.Errorf("Not fully implemented..!!!")
		config, errConfig := pkg.LoadConfig()
		if errConfig != nil {
			log.Errorf("Failed to load config file %s", errConfig)
			os.Exit(1)
			return
		}

		stackDef := config.GetStackByName(Stack)
		if stackDef == nil {
			log.Errorf("Stack not configured in .kappy.yaml: %s", Stack)
			os.Exit(1)
			return
		}

		_, errClusterConf := kubernetes.GetConfig(stackDef.ClusterName)

		if errClusterConf != nil {
			log.Errorf("Could not get cluster configuration: %s", errClusterConf)
			os.Exit(1)
			return
		}

		//  TODO: forward credentials straight to kubectl

		kubectlCmd := exec.Command("kubectl", args...)
		kubectlCmd.Stdin = os.Stdin

		outFile, err := pty.Start(kubectlCmd)

		if err != nil {
			os.Exit(1)
			return
		}

		io.Copy(os.Stdout, outFile)
		err = kubectlCmd.Wait()

		if err != nil {
			if exitErr, ok := err.(*exec.ExitError); ok {
				os.Exit(exitErr.ExitCode())
				return
			}

			os.Exit(1)
		}
	},
}

func initKubectlCmd() {
	kubectlCmd.Flags().StringVarP(&Stack, "stack", "", "", "Stack to connect to with kubectl")
	kubectlCmd.MarkFlagRequired("stack")
}
