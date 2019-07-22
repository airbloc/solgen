package deployment

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
)

type Deployments map[string]abi.ABI
type rawData map[string]map[string]interface{}

func parseDeployments(rawData rawData) (Deployments, error) {
	deployments := make(Deployments, len(rawData))

	for contractName, contractInfo := range rawData {
		rawAbi, _ := json.Marshal(contractInfo["abi"])
		parsedAbi, err := abi.JSON(bytes.NewReader(rawAbi))
		if err != nil {
			return nil, err
		}

		deployments[contractName] = parsedAbi
	}

	return deployments, nil
}

func fetchFromUrl(path string) (Deployments, error) {
	resp, err := http.Get(path)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	rawData := make(rawData)
	decoder := json.NewDecoder(resp.Body)
	if err := decoder.Decode(&rawData); err != nil {
		return nil, err
	}

	return parseDeployments(rawData)
}

func fetchFromFile(path string) (Deployments, error) {
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

func GetDeploymentsFrom(path string) (Deployments, error) {
	if strings.HasPrefix(path, "http://") || strings.HasPrefix(path, "https://") {
		return fetchFromUrl(path)
	}
	return fetchFromFile(path)
}
