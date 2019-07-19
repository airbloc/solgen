package proto

import (
	"log"
	"testing"
)

func TestGetDeploymentsFromUrl(t *testing.T) {
	deployments, err := GetDeploymentsFromUrl("http://localhost:8500")
	if err != nil {
		panic(err)
	}
	contracts := parseContracts(deployments)

	for _, contract := range contracts {
		log.Println(Render("./test/"+contract.PackageName+".proto", contract))
	}
}
