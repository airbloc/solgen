package main

import "github.com/kelseyhightower/envconfig"

type Config struct {
	DeploymentPath string `envconfig:"deployment_path"`
	OptionPath     string `envconfig:"option_path"`
	OutputPath     string `envconfig:"output_path"`
}

func NewConfig() (config Config) {
	envconfig.MustProcess("solgen", &config)
	return
}
