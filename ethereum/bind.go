package ethereum

import (
	"errors"
	"log"
	"os"
	"path/filepath"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/frostornge/solgen/deployment"
	"github.com/frostornge/solgen/utils"
)

type binder struct {
	data         *tmplData
	contractName string
	typeOptions  option
}

func (bind *binder) parseData(evmABI abi.ABI, pkg string) error {
	log.SetFlags(log.Llongfile)

	contract, err := parseContract(
		evmABI,
		bind.contractName,
		bind.typeOptions,
	)
	if err != nil {
		return err
	}

	bind.data = &tmplData{
		Package:  pkg,
		Contract: contract,
	}
	return nil
}

func GenerateBind(path string, deployments deployment.Deployments, typeOptions Options) error {
	log.SetFlags(log.Llongfile)

	stat, err := os.Stat(path)
	if os.IsNotExist(err) {
		if err = os.MkdirAll(path, os.ModePerm); err != nil {
			return err
		}
	} else {
		if !stat.IsDir() {
			return errors.New("is not directory")
		}
	}

	for contractName, contractAbi := range deployments {
		opt := make(option)
		opt.apply(typeOptions["default"])    // default
		opt.apply(typeOptions[contractName]) // override it

		bind := &binder{contractName: contractName, typeOptions: opt}
		if err := bind.parseData(contractAbi, "adapter"); err != nil {
			return err
		}

		file := filepath.Join(path, utils.ToSnakeCase(contractName)+".go")
		if err = RenderFile(file, bind.data); err != nil {
			return err
		}
	}

	return nil
}
