package main

import (
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/frostornge/solgen/bind"
	"github.com/frostornge/solgen/deployment"
	"github.com/frostornge/solgen/klaytn"
	"github.com/frostornge/solgen/proto"
	"github.com/frostornge/solgen/utils"
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
	BindTypeKlay  = "klay"
	BindTypeProto = "proto"
)

func init() {
	rflags := rootCmd.PersistentFlags()
	rflags.StringVarP(&rootFlags.typeFlag, "type", "t", BindTypeKlay, "Bind type")
	rflags.StringVarP(&rootFlags.inputPath, "input", "i", "http://localhost:8500", "Input path")
	rflags.StringVarP(&rootFlags.outputPath, "output", "o", "./out/", "Output path")
	rflags.StringVarP(&rootFlags.optionPath, "option", "p", "./option.json", "Option path")
}

func openFile(path string) (*os.File, error) {
	file, err := os.OpenFile(path, os.O_WRONLY, os.ModePerm)
	if os.IsNotExist(err) {
		file, err = os.Create(path)
		if err != nil {
			return nil, err
		}
	} else if err != nil {
		return nil, err
	}

	return file, nil
}

func writeFile(abiPath string, outPath string, abi io.Reader) (err error) {
	if _, err := openFile(outPath); err != nil {
		return err
	}

	file, err := openFile(abiPath)
	if err != nil {
		return err
	}
	defer func() { err = file.Close() }()

	if _, err := io.Copy(file, abi); err != nil {
		return err
	}
	return
}

func main() {
	tmp := os.TempDir()

	contracts, err := deployment.GetDeploymentsFrom("http://localhost:8500")
	if err != nil {
		panic(err)
	}

	for name, contract := range contracts {
		// If the entire solidity code was specified, build and bind based on that
		var (
			abis  []string
			bins  []string
			types []string
			sigs  []map[string]string
			libs  = make(map[string]string)
		)

		filename := utils.ToSnakeCase(name)
		abiPath, _ := filepath.Abs(filepath.Join(tmp, filename+".abi"))
		outPath, _ := filepath.Abs(filepath.Join("./test", "bind", filename+".go"))
		abi := strings.NewReader(contract.RawABI)
		if err := writeFile(abiPath, outPath, abi); err != nil {
			panic(err)
		}

		// Generate the contract binding
		code, err := bind.Bind(abi, "adapter", bind.LangGo)
		if err != nil {
			panic(err)
		}
		print(code)
	}

	//if err := rootCmd.Execute(); err != nil {
	//	panic(err)
	//}
}

func run() func(*cobra.Command, []string) {
	return func(cmd *cobra.Command, args []string) {
		deployments, err := deployment.GetDeploymentsFrom(rootFlags.inputPath)
		if err != nil {
			panic(err)
		}

		switch rootFlags.typeFlag {
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
