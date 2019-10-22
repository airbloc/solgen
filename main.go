package main

import (
	"os"
	"path"
	"path/filepath"

	"github.com/frostornge/solgen/bind"
	"github.com/frostornge/solgen/deployment"
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

func writeFile(path, filename string, contract deployment.Deployment, option bind.Customs) error {
	bindFile, err := openFile(filepath.Join(path, filename+".go"))
	if err != nil {
		return err
	}
	defer bindFile.Close()

	wrapFile, err := openFile(filepath.Join(path, filename+"_wrapper.go"))
	if err != nil {
		return err
	}
	defer wrapFile.Close()

	// Generate the contract binding
	if err := bind.Bind(
		bindFile, wrapFile,
		filename, contract.RawABI, "adapter",
		option, bind.PlatformKlay, bind.LangGo,
	); err != nil {
		return err
	}
	return nil
}

func main() {
	contracts, err := deployment.GetDeploymentsFrom("http://localhost:8500")
	if err != nil {
		panic(err)
	}

	options := map[string]bind.Customs{
		"Accounts": {
			Structs: map[string]string{"(address,uint8,address,address)": "types.Account"},
			Methods: map[string]bool{
				"create":                     true,
				"createTemporary":            true,
				"unlockTemporary":            true,
				"setController":              true,
				"getAccount":                 true,
				"getAccountByIdentityHash":   true,
				"getAccountId":               true,
				"getAccountIdByIdentityHash": true,
				"getAccountIdFromSignature":  true,
				"isTemporary":                true,
				"isControllerOf":             true,
				"exists":                     true,
			},
		},
		"AppRegistry": {
			Structs: map[string]string{"(string,address,address)": "types.App"},
			Methods: map[string]bool{
				"register":         true,
				"unregister":       true,
				"get":              true,
				"exists":           true,
				"isOwner":          true,
				"transferAppOwner": true,
			},
		},
		"ControllerRegistry": {
			Structs: map[string]string{"(address,uint256)": "types.DataController"},
			Methods: map[string]bool{
				"register": true,
				"get":      true,
				"exists":   true,
			},
		},
		"Consents": {
			Structs: map[string]string{"(uint8,string,bool)": "types.ConsentData", "(uint8,string,bool)[]": "[]types.ConsentData"},
			Methods: map[string]bool{
				"consent":                       true,
				"consentMany":                   true,
				"consentByController":           true,
				"consentManyByController":       true,
				"modifyConsentByController":     true,
				"modifyConsentManyByController": true,
				"isAllowed":                     true,
				"isAllowedAt":                   true,
			},
		},
		"DataTypeRegistry": {
			Structs: map[string]string{"(string,address,bytes32)": "types.DataType"},
			Methods: map[string]bool{
				"register":   true,
				"unregister": true,
				"get":        true,
				"exists":     true,
				"isOwner":    true,
			},
		},
		"Exchange": {
			Structs: map[string]string{
				"(string,address,bytes20[],uint256,uint256,(address,bytes4,bytes),uint8)": "types.Offer",
				"(address,bytes4,bytes)":                                                  "types.Escrow",
			},
			Methods: map[string]bool{
				"prepare":         true,
				"addDataIds":      true,
				"order":           true,
				"cancel":          true,
				"settle":          true,
				"reject":          true,
				"offerExists":     true,
				"getOffer":        true,
				"getOfferMembers": true,
			},
		},
	}

	if err := os.RemoveAll("./test/bind"); err != nil {
		panic(err)
	}

	if err := os.MkdirAll("./test/bind", os.ModePerm); err != nil {
		panic(err)
	}

	for name, contract := range contracts {
		fn := utils.ToSnakeCase(name)
		fp, _ := filepath.Abs(path.Join("./test", "bind"))

		if err := writeFile(fp, fn, contract, options[name]); err != nil {
			panic(err)
		}
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
		case BindTypeProto:
			if err := proto.GenerateBind(rootFlags.outputPath, deployments, proto.Options{}); err != nil {
				panic(err)
			}
		}
	}
}
