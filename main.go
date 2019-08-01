package main

import (
	"github.com/frostornge/solgen/deployment"
	"github.com/frostornge/solgen/ethereum"
	"github.com/frostornge/solgen/klaytn"
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
		optionPath string
	}
)

const (
	BindTypeEth   = "eth"
	BindTypeKlay  = "klay"
	BindTypeProto = "proto"
)

func init() {
	rflags := rootCmd.PersistentFlags()
	rflags.StringVarP(&rootFlags.typeFlag, "type", "t", BindTypeEth, "Bind type")
	rflags.StringVarP(&rootFlags.inputPath, "input", "i", "http://localhost:8500", "Input path")
	rflags.StringVarP(&rootFlags.outputPath, "output", "o", "./out/", "Output path")
	rflags.StringVarP(&rootFlags.optionPath, "option", "p", "./option.json", "Option path")
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
		case BindTypeEth:
			opts, err := ethereum.GetOption(rootFlags.optionPath)
			if err != nil {
				panic(err)
			}

			if err := ethereum.GenerateBind(rootFlags.outputPath, deployments, opts); err != nil {
				panic(err)
			}
		case BindTypeKlay:
			opts, err := klaytn.GetOption(rootFlags.optionPath)
			if err != nil {
				panic(err)
			}

			if err := klaytn.GenerateBind(rootFlags.outputPath, deployments, opts); err != nil {
				panic(err)
			}
		case BindTypeProto:
			if err := proto.GenerateBind(rootFlags.outputPath, deployments, proto.Options{}); err != nil {
				panic(err)
			}
		}
	}
}
