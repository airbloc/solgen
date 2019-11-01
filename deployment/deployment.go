package deployment

import (
	"bytes"
	"encoding/json"
	"io"
	"math/big"
	"net/http"
	"os"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/pkg/errors"
)

type Deployment struct {
	Address   common.Address           `json:"address"`
	TxHash    common.Hash              `json:"tx_hash"`
	CreatedAt *big.Int                 `json:"created_at"`
	ParsedABI []map[string]interface{} `json:"abi"`
	EvmABI    abi.ABI                  `json:"-"`
	RawABI    []byte                   `json:"-"`
}

type Deployments map[string]Deployment

func fromUrl(path string) (io.ReadCloser, error) {
	resp, err := http.Get(path)
	if err != nil {
		return nil, err
	}
	return resp.Body, err
}

func fromFile(path string) (io.ReadCloser, error) {
	return os.OpenFile(path, os.O_RDONLY, os.ModePerm)
}

func GetDeploymentsFrom(path string) (Deployments, error) {
	reader, err := func(path string) (io.ReadCloser, error) {
		if strings.HasPrefix(path, "http://") || strings.HasPrefix(path, "https://") {
			return fromUrl(path)
		}
		return fromFile(path)
	}(path)
	if err != nil {
		return nil, err
	}
	defer reader.Close()

	deployments := make(Deployments)
	if err := json.NewDecoder(reader).Decode(&deployments); err != nil {
		return nil, err
	}

	for contractName, deployment := range deployments {
		rawABI, err := json.Marshal(deployment.ParsedABI)
		if err != nil {
			return nil, errors.Wrap(err, "parse to raw abi")
		}

		evmABI, err := abi.JSON(bytes.NewReader(rawABI))
		if err != nil {
			return nil, errors.Wrap(err, "parse to evm abi")
		}

		deployment.RawABI = rawABI
		deployment.EvmABI = evmABI
		deployments[contractName] = deployment
	}
	return deployments, nil
}
