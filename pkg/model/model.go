package model

import "github.com/kontrio/kappy/pkg/kstrings"

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

func (sd *StackDefinition) GetServiceConfig(name string) *ServiceConfig {
	conf, ok := sd.Config[name]
	if !ok {
		return nil
	}

	return &conf
}

type ServiceConfig struct {
	Ingress         []string          `mapstructure:"ingress"`
	ContainerConfig []ContainerConfig `mapstructure:"containers"`
}

func (sc *ServiceConfig) GetContainerConfigByName(name string) *ContainerConfig {
	if sc.ContainerConfig == nil || sc == nil {
		return nil
	}
	for _, conf := range sc.ContainerConfig {
		if conf.Name == name {
			return &conf
		}
	}

	return nil
}

type ContainerConfig struct {
	Name     string            `mapstructure:"name"`
	AppRoot  string            `mapstructure:"app_dir"`
	EnvFile  string            `mapstructure:"env_file"`
	Env      map[string]string `mapstructure:"env"`
	BuildTag string            `mapstructure:"build_tag"`

	// Overrides build section values in service
	Build *BuildDefinition `mapstructure:"build"`
}

type ServiceDefinition struct {
	Name                    string                `mapstructure:"name"`
	Replicas                int32                 `mapstructure:"replicas"`
	MinReadySeconds         int32                 `mapstructure:"minReadySeconds"`
	ProgressDeadlineSeconds int32                 `mapstructure:"ProgressDeadlineSeconds"`
	RevisionHistoryLimit    int32                 `mapstructure:"revisionHistoryLimit"`
	MaxSurge                int32                 `mapstructure:"maxSurge"`
	MaxUnavailable          int32                 `mapstructure:"maxUnavailable"`
	ServicePorts            []string              `mapstructure:"service_ports"`
	InternalOnly            bool                  `mapstructure:"internal"`
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
	Context    string            `mapstructure:"context"`
	Dockerfile string            `mapstructure:"dockerfile"`
	BuildArgs  map[string]string `mapstructure:"args"`
	CacheFrom  []string          `mapstructure:"cache_from"`
	Tags       []string          `mapstructure:"tags"`
	Labels     []string          `mapstructure:"labels"`
	ShmSize    string            `mapstructure:"shm_size"`
	Target     string            `mapstructure:"target"`
}

// a takes priority
func MergeBuildDefinitions(base *BuildDefinition, a *BuildDefinition) *BuildDefinition {
	if base != nil && a == nil {
		return base
	}

	if a != nil && base == nil {
		return a
	}

	if a == nil && base == nil {
		return nil
	}

	target := BuildDefinition{
		Context:    base.Context,
		Dockerfile: base.Dockerfile,
		BuildArgs:  base.BuildArgs,
		CacheFrom:  base.CacheFrom,
		Tags:       base.Tags,
		Labels:     base.Labels,
		ShmSize:    base.ShmSize,
		Target:     base.Target,
	}

	if !kstrings.IsEmpty(&a.Context) {
		target.Context = a.Context
	}

	if !kstrings.IsEmpty(&a.Dockerfile) {
		target.Dockerfile = a.Dockerfile
	}

	if len(a.BuildArgs) > 0 {
		target.BuildArgs = kstrings.MergeMaps(base.BuildArgs, a.BuildArgs)
	}

	if len(a.CacheFrom) > 0 {
		target.CacheFrom = a.CacheFrom
	}

	if len(a.Tags) > 0 {
		target.Tags = a.Tags
	}

	if len(a.Labels) > 0 {
		target.Labels = a.Labels
	}

	if !kstrings.IsEmpty(&a.ShmSize) {
		target.ShmSize = a.ShmSize
	}

	if !kstrings.IsEmpty(&a.Target) {
		target.Target = a.Target
	}

	return &target
}
