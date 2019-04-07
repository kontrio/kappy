package cmd

import (
	"fmt"
	"os"

	"github.com/apex/log"
	"github.com/ericchiang/k8s"
	"github.com/joho/godotenv"
	"github.com/kontr/kappy/pkg"
	"github.com/kontr/kappy/pkg/kubernetes"
	"github.com/kontr/kappy/pkg/model"
	"github.com/spf13/cobra"
)

var DeployVersion string

var deployCmd = &cobra.Command{
	Use:   "deploy",
	Short: "Deploy an application or a set of applications to a Kubernetes cluster",
	Args:  cobra.ArbitraryArgs,
	Run: func(cmd *cobra.Command, args []string) {
		log.Debug("Loading config file")
		config, err := pkg.LoadConfig()

		if err != nil {
			log.Errorf("Failed to load config file %s", err)
			os.Exit(1)
		}

		deploymentStack := args[0]

		log.Infof("Deploying stack: %s", deploymentStack)
		stackDef := config.GetStackByName(deploymentStack)

		if stackDef == nil {
			log.Errorf("Stack not configured in .kappy.yaml: %s", deploymentStack)
			os.Exit(1)
			return
		}

		client, errK8s := kubernetes.CreateClient(stackDef.ClusterName)

		if errK8s != nil {
			log.Errorf("Failed to create Kubernetes client: %s", errK8s)
			os.Exit(1)
			return
		}

		errNamespace := kubernetes.CreateNamespace(client, stackDef.Namespace)

		if errNamespace != nil {
			log.Errorf("Could not create k8s namespace for stack definition: %s", errNamespace)
			os.Exit(1)
			return
		}

		for serviceName, service := range config.Services {
			// TODO: refactor moving the key to the property.
			service.Name = serviceName
			namespace := stackDef.Namespace

			serviceConfig, present := stackDef.Config[serviceName]

			if !present {
				continue
			}

			for _, containerConfig := range serviceConfig.ContainerConfig {
				err := configureSecrets(client, serviceName, namespace, &containerConfig)

				if err != nil {
					log.Errorf("Failed to configure secrets for '%s' in the '%s' container", serviceName, containerConfig.Name)
					log.Errorf("Error: %s", err)
					os.Exit(1)
					return
				}
			}

			err := kubernetes.CreateUpdateIngress(client, &serviceConfig, serviceName, namespace)
			if err != nil {
				log.Errorf("Failed to create ingress: '%s'", serviceName)
				log.Errorf("Error: %s", err)
				os.Exit(1)
				return
			}

			err = kubernetes.CreateUpdateService(client, &service, namespace)
			if err != nil {
				log.Errorf("Failed to create service: '%s'", serviceName)
				log.Errorf("Error: %s", err)
				os.Exit(1)
				return
			}

			err = deployService(client, &service, &serviceConfig, namespace, DeployVersion, config.DockerRegistry)
			if err != nil {
				log.Errorf("Failed to deploy.. '%s'", serviceName)
				log.Errorf("Error: %s", err)
				os.Exit(1)
				return
			}
		}
	},
}

func deployService(client *k8s.Client, service *model.ServiceDefinition, serviceConfig *model.ServiceConfig, namespace, deployVersion, dockerRegistry string) error {
	return kubernetes.CreateUpdateDeployment(client, service, serviceConfig, namespace, deployVersion, dockerRegistry)
}

func createSecretReference(serviceName, containerName string) string {
	return fmt.Sprintf("%s-%s-secrets", serviceName, containerName)
}

func configureSecrets(client *k8s.Client, serviceName, namespace string, containerConfig *model.ContainerConfig) error {
	secretReference := createSecretReference(serviceName, containerConfig.Name)
	envVars := containerConfig.Env
	if len(containerConfig.EnvFile) > 0 {
		log.Infof("Loading environment variables from %s", containerConfig.EnvFile)
		loadedEnvVars, err := godotenv.Read(containerConfig.EnvFile)
		if err != nil {
			return err
		}

		// Second argument take precedence
		envVars = mergeMap(envVars, loadedEnvVars)
	}

	return kubernetes.CreateSecret(client, secretReference, namespace, envVars)
}

func initDeployCmd() {
	deployCmd.Flags().StringVarP(&DeployVersion, "version", "", "", "Version (corresponds to repository image tag")
	deployCmd.MarkFlagRequired("version")
}

func mergeMap(a, b map[string]string) map[string]string {
	newMap := make(map[string]string)
	for ka, va := range a {
		newMap[ka] = va
	}

	for kb, vb := range b {
		newMap[kb] = vb
	}

	return newMap
}
