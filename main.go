package main

import (
	"github.com/frostornge/solgen/bind"
	"github.com/frostornge/solgen/deployment"
	"github.com/frostornge/solgen/proto"
	"github.com/spf13/cobra"
)

var (
	rootCmd = &cobra.Command{
		Use:     "solgen",
		Short:   "solgen is solidity bind generator",
		Long:    "Solgen is a tool for generate solidity binds.\n" + "This application helps to generate go/proto bind of solidity.\n",
		Version: "0.0.0",
		Run:     run(),
	}

	rootFlags struct {
		typeFlag   string
		inputPath  string
		outputPath string
	}
)

func init() {
	rflags := rootCmd.PersistentFlags()
	rflags.StringVarP(&rootFlags.typeFlag, "type", "t", "go", "Bind type")
	rflags.StringVarP(&rootFlags.inputPath, "input", "i", "http://localhost:8500", "Input path")
	rflags.StringVarP(&rootFlags.outputPath, "output", "o", "./out/", "Output path")
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		panic(err)
	}
}

func run() func(*cobra.Command, []string) {
	return func(cmd *cobra.Command, args []string) {
		deployments, err := deployment.GetDeploymentsFrom(rootFlags.inputPath)
		if err != nil {
			panic(err)
		}

		switch rootFlags.typeFlag {
		case "go":
			if err := bind.GenerateBind(rootFlags.outputPath, deployments); err != nil {
				panic(err)
			}
		case "proto":
			if err := proto.GenerateBind(rootFlags.outputPath, deployments); err != nil {
				panic(err)
			}
		}
	}
}
