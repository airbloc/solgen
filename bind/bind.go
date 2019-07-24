package bind

import (
	"errors"
	"os"
	"path/filepath"

	"github.com/frostornge/solgen/deployment"
)

func GenerateBind(path string, deployments deployment.Deployments) error {
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
		data := parseData(contractName, contractAbi, "adapter")
		file := filepath.Join(path, contractName+".go")
		if err = RenderFile(file, data); err != nil {
			return err
		}
	}

	return nil
}
