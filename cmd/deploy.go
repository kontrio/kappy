package cmd

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/apex/log"
	"github.com/joho/godotenv"
	"github.com/kontrio/kappy/pkg/kstrings"
	"github.com/kontrio/kappy/pkg/kubernetes"
	"github.com/kontrio/kappy/pkg/model"
	"github.com/kontrio/kappy/pkg/phase"
	"github.com/spf13/cobra"
	k8s "k8s.io/client-go/kubernetes"
)

var DeployVersion string

var deployCmd = &cobra.Command{
	Use:   "deploy [stackname]",
	Short: "Deploy an application or a set of applications to a Kubernetes cluster",
	Args:  ArgsLoadConfigAndStackName,
	Run: func(cmd *cobra.Command, args []string) {
		deploymentStack := args[0]
		log.Infof("Deploying stack: %s", deploymentStack)

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

			if present {
				for _, containerConfig := range serviceConfig.ContainerConfig {
					err := configureSecrets(client, serviceName, namespace, &containerConfig)

					if err != nil {
						log.Errorf("Failed to configure secrets for '%s' in the '%s' container", serviceName, containerConfig.Name)
						log.Errorf("Error: %s", err)
						os.Exit(1)
						return
					}
				}
			} else {
				serviceConfig = model.ServiceConfig{}
			}

			// Create ingress for everything external
			if !service.InternalOnly {
				err := kubernetes.CreateUpdateIngress(client, &serviceConfig, serviceName, namespace)
				if err != nil {
					log.Errorf("Failed to create ingress: '%s'", serviceName)
					log.Errorf("Error: %s", err)
					os.Exit(1)
					return
				}
			}

			err := kubernetes.CreateUpdateService(client, &service, namespace)
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

			log.Infof("Watching deployment status - ctrl+c to cancel watch and continue..")
			err = phase.NewPhase(context.Background()).
				WithTimeout(1 * time.Minute).
				CancelWithSignal(os.Interrupt).
				Run(func(ctx context.Context) error {
					return kubernetes.WatchDeployment(ctx, client, namespace, serviceName)
				})

			if err != nil {
				log.Errorf("Failed to watch deployment.. '%s'", serviceName)
				log.Errorf("Error: %s", err)
			}

			err = kubernetes.GetFailingPods(context.Background(), client, namespace, serviceName)
			if err != nil {
				log.Errorf("Failed to get pod statuses '%s'", serviceName)
				log.Errorf("Error: %s", err)
			}
		}
	},
}

func deployService(client *k8s.Clientset, service *model.ServiceDefinition, serviceConfig *model.ServiceConfig, namespace, deployVersion, dockerRegistry string) error {
	return kubernetes.UpsertDeployment(client, service, serviceConfig, namespace, deployVersion, dockerRegistry)
}

func createSecretReference(serviceName, containerName string) string {
	return fmt.Sprintf("%s-%s-secrets", serviceName, containerName)
}

func configureSecrets(client *k8s.Clientset, serviceName, namespace string, containerConfig *model.ContainerConfig) error {
	secretReference := createSecretReference(serviceName, containerConfig.Name)
	envVars := containerConfig.Env
	if len(containerConfig.EnvFile) > 0 {
		log.Infof("Loading environment variables from %s", containerConfig.EnvFile)
		loadedEnvVars, err := godotenv.Read(containerConfig.EnvFile)
		if err != nil {
			return err
		}

		// Second argument take precedence
		envVars = kstrings.MergeMaps(envVars, loadedEnvVars)
	}

	return kubernetes.CreateSecret(client, secretReference, namespace, envVars, map[string]string{
		"kappy.managed": serviceName,
	})
}

func initDeployCmd() {
	deployCmd.Flags().StringVarP(&DeployVersion, "version", "", "", "Version (corresponds to repository image tag")
	deployCmd.MarkFlagRequired("version")
}
