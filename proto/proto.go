package proto

import (
	"errors"
	"os"
	"path/filepath"

	"github.com/frostornge/solgen/deployments"
)

func GenerateBind(path string, deployments deployments.Deployments) error {
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

	contracts := parseContracts(deployments)

	for _, contract := range contracts {
		file := filepath.Join(path, contract.PackageName+".proto")
		if err := RenderFile(file, contract); err != nil {
			return err
		}
	}
	return nil
}
