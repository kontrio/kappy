package model

type Config struct {
	DockerRegistry string                       `mapstructure:"docker_registry"`
	Services       map[string]ServiceDefinition `mapstructure:"services"`
	Stacks         []StackDefinition            `mapstructure:"stacks"`
}

func (c *Config) GetStackByName(name string) *StackDefinition {
	for _, stack := range c.Stacks {
		if name == stack.Namespace {
			return &stack
		}
	}

	return nil
}

type StackDefinition struct {
	Namespace   string                   `mapstructure:"name"`
	ClusterName string                   `mapstructure:"k8s_cluster_name"`
	Config      map[string]ServiceConfig `mapstructure:"services"`
}

type ServiceConfig struct {
	Ingress         []string          `mapstructure:"ingress"`
	Replicas        int32             `mapstructure:"replicas"`
	ContainerConfig []ContainerConfig `mapstructure:"containers"`
}

func (sc *ServiceConfig) GetContainerConfigByName(name string) *ContainerConfig {
	for _, conf := range sc.ContainerConfig {
		if conf.Name == name {
			return &conf
		}
	}

	return nil
}

type ContainerConfig struct {
	Name    string            `mapstructure:"name"`
	AppRoot string            `mapstructure:"app_dir"`
	EnvFile string            `mapstructure:"env_file"`
	Env     map[string]string `mapstructure:"env"`
}

type ServiceDefinition struct {
	Name                    string                `mapstructure:"name"`
	Replicas                int32                 `mapstructure:"replicas"`
	MinReadySeconds         int32                 `mapstructure:"minReadySeconds"`
	ProgressDeadlineSeconds int32                 `mapstructure:"ProgressDeadlineSeconds"`
	RevisionHistoryLimit    int32                 `mapstructure:"revisionHistoryLimit"`
	MaxSurge                int32                 `mapstructure:"maxSurge"`
	MaxUnavailable          int32                 `mapstructure:"maxUnavailable"`
	Containers              []ContainerDefinition `mapstructure:"containers"`
	// TODO add volumes
}

type ProbeDefinition struct {
	// todo
}

type ContainerDefinition struct {
	Name           string           `mapstructure:"name"`
	Command        []string         `mapstructure:"cmd"`
	Args           []string         `mapstructure:"args"`
	Build          *BuildDefinition `mapstructure:"build"`
	Image          string           `mapstructure:"image"`
	ExposePort     int32            `mapstructure:"port"`
	ReadinessProbe *ProbeDefinition `mapstructure:"readiness_probe"`
	LivenessProbe  *ProbeDefinition `mapstructure:"liveness_probe"`
	// TODO: Add volume mounts
}

type BuildDefinition struct {
	Context    string   `mapstructure:"context"`
	Dockerfile string   `mapstructure:"dockerfile"`
	BuildArgs  []string `mapstructure:"args"`
	CacheFrom  []string `mapstructure:"cache_from"`
	Tags       []string `mapstructure:"tags"`
	Labels     []string `mapstructure:"labels"`
	ShmSize    string   `mapstructure:"shm_size"`
	Target     string   `mapstructure:"target"`
}
