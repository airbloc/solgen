package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"

	"github.com/airbloc/solgen/bind"
	"github.com/airbloc/solgen/bind/language"
	"github.com/airbloc/solgen/bind/platform"
	"github.com/airbloc/solgen/deployment"
	"github.com/airbloc/solgen/utils"

	"github.com/spf13/cobra"
)

var (
	config    = NewConfig()
	cmdConfig = Config{}

	rootCmd = &cobra.Command{
		Use:   "solgen",
		Short: "Golang ABI bind generator for Airbloc",
		Long: "Solgen is a tool for generate solidity binds.\n" +
			"This application helps to generate go/proto bind of solidity.",
		Version: "v0.1.3",
		Run:     func(cmd *cobra.Command, args []string) { run() },
	}
)

func init() {
	cobra.OnInitialize(initConfig)

	flags := rootCmd.PersistentFlags()
	flags.StringVar(&cmdConfig.DeploymentPath, "deployment", "", "path of deployment (json)")
	flags.StringVar(&cmdConfig.OptionPath, "opt", "", "path of custom bind options")
	flags.StringVar(&cmdConfig.OutputPath, "out", "./build", "path of generated output")
}

func initConfig() {
	// merge config
	if config.DeploymentPath == "" || cmdConfig.DeploymentPath != "" {
		config.DeploymentPath = cmdConfig.DeploymentPath
	}
	if config.OptionPath == "" || cmdConfig.OptionPath != "" {
		config.OptionPath = cmdConfig.OptionPath
	}
	if config.OutputPath == "" || cmdConfig.OutputPath != "" {
		config.OutputPath = cmdConfig.OutputPath
	}

	if config.DeploymentPath == "" {
		panic("deployment path needed")
	}
}

func run() {
	deployments, err := deployment.GetDeploymentsFrom(config.DeploymentPath)
	if err != nil {
		panic(err)
	}

	customs := make(map[string]bind.Customs)
	if config.OptionPath != "" {
		opt, err := ioutil.ReadFile("option_bind_airbloc.json")
		if err != nil {
			panic(err)
		}

		if err := json.Unmarshal(opt, &customs); err != nil {
			panic(err)
		}
	}

	for _, mode := range bind.Modes {
		if err := os.MkdirAll(path.Clean(path.Join(config.OutputPath, string(mode))), os.ModePerm); err != nil {
			panic(err)
		}
	}

	for name, contract := range deployments {
		codes, err := bind.Bind(
			name, contract,
			bind.Option{
				Customs:  customs[name],
				Platform: platform.Klaytn,
				Language: language.Go,
			},
		)
		if err != nil {
			log.Println(err)
			continue
		}

		for _, mode := range bind.Modes {
			code, ok := codes[mode]
			if !ok {
				log.Println(mode, "not found")
				continue
			}

			filename := filepath.Clean(filepath.Join(config.OutputPath, string(mode), utils.ToSnakeCase(name)+".go"))
			if err := ioutil.WriteFile(filename, code, os.ModePerm); err != nil {
				log.Println(err)
			}
		}
	}
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		panic(err)
	}
}
