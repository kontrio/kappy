package pkg

import (
	"github.com/apex/log"
	"github.com/kontrio/kappy/pkg/model"
	viper "github.com/spf13/viper"
)

func initConfig() (*viper.Viper, error) {
	v := viper.New()
	v.AddConfigPath(".")
	v.SetConfigName(".kappy")

	err := v.ReadInConfig()

	if err != nil {
		return nil, err
	}

	return v, nil
}

func LoadConfig() (*model.Config, error) {
	v, errConfig := initConfig()

	if errConfig != nil {
		return nil, errConfig
	}

	config := model.Config{}
	errMarshal := v.Unmarshal(&config)
	if errMarshal != nil {
		return nil, errMarshal
	}

	log.Debug("Loaded config")
	errValidation := validateConfig(&config)
	return &config, errValidation
}

func validateConfig(config *model.Config) error {
	for serviceName, serviceDef := range config.Services {
		serviceDef.Name = serviceName
	}

	return nil
}
