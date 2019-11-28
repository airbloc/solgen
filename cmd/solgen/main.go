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
	}
)

func init() {
	solgenCmd := &cobra.Command{
		Use:   "run",
		Short: "Generate bind",
		Run:   func(cmd *cobra.Command, args []string) { run() },
	}

	cobra.OnInitialize(initConfig)

	flags := solgenCmd.PersistentFlags()
	flags.StringVar(&cmdConfig.DeploymentPath, "deployment", "http://localhost:8500", "endpoint of deployment")
	flags.StringVar(&cmdConfig.OptionPath, "opt", "", "path of custom bind options")
	flags.StringVar(&cmdConfig.OutputPath, "out", "", "path of generated output")

	rootCmd.AddCommand(solgenCmd)
}

func initConfig() {
	// merge config
}

func main() {
	deployments, err := deployment.GetDeploymentsFrom("cmd/solgen/deployment.test.json")
	//deployments, err := deployment.GetDeploymentsFrom("http://localhost:8500")
	if err != nil {
		panic(err)
	}

	customs := make(map[string]bind.Customs)
	opt, err := ioutil.ReadFile("option_bind_airbloc.json")
	if err != nil {
		panic(err)
	}

	if err := json.Unmarshal(opt, &customs); err != nil {
		panic(err)
	}

	base := "test/bind"

	if err := os.RemoveAll(base); err != nil {
		panic(err)
	}

	for _, mode := range bind.Modes {
		if err := os.MkdirAll(path.Join(base, string(mode)), os.ModePerm); err != nil {
			panic(err)
		}
	}

	for name, deployment := range deployments {
		codes, err := bind.Bind(
			name, deployment,
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

			func() {
				filename := filepath.Join(base, string(mode), utils.ToSnakeCase(name)+".go")
				if err := ioutil.WriteFile(filename, code, os.ModePerm); err != nil {
					log.Println(err)
					return
				}
			}()
		}
	}
	//bind.Bind()
}

func run() {

}
