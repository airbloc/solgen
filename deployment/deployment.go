package deployment

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/ethereum/go-ethereum/accounts/abi"
)

type Deployments map[string]abi.ABI
type rawData map[string]map[string]interface{}

func parseDeployments(rawData rawData) (Deployments, error) {
	deployments := make(Deployments, len(rawData))

	for contractName, contractInfo := range rawData {
		rawAbi, err := json.Marshal(contractInfo["abi"])
		if err != nil {
			return nil, err
		}

		parsedAbi, err := abi.JSON(bytes.NewReader(rawAbi))
		if err != nil {
			return nil, err
		}

		deployments[contractName] = parsedAbi
	}

	return deployments, nil
}

func GetDeploymentsFromFile(path string) (Deployments, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	rawData := make(rawData)
	if err := json.Unmarshal(data, &rawData); err != nil {
		return nil, err
	}

	return parseDeployments(rawData)
}

func GetDeploymentsFromUrl(url string) (Deployments, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	rawData := make(rawData)
	if err := json.NewDecoder(resp.Body).Decode(&rawData); err != nil {
		return nil, err
	}

	return parseDeployments(rawData)
}
