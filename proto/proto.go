package proto

import (
	"errors"
	"os"
	"path/filepath"

	"github.com/frostornge/solgen/deployment"
)

type option map[string]string
type Options map[string]option

type binder struct {
	deployments deployment.Deployments
	typeOptions Options
	contracts   []contract
}

func (bind *binder) parseContracts(deployments deployment.Deployments) {
	for name, deployment := range deployments {
		contract := &contract{typeOptions: bind.typeOptions, contractName: name}
		contract.parseContract(deployment)
		bind.contracts = append(bind.contracts, *contract)
	}
}

func GenerateBind(path string, deployments deployment.Deployments, typeOptions Options) error {
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

	bind := &binder{
		deployments: deployments,
		typeOptions: typeOptions,
	}
	bind.parseContracts(deployments)

	for _, contract := range bind.contracts {
		file := filepath.Join(path, contract.PackageName+".proto")
		if err := RenderFile(file, contract); err != nil {
			return err
		}
	}
	return nil
}
