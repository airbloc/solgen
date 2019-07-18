package main

import (
	"log"

	"github.com/frostornge/sol2proto/proto"
)

func main() {
	deployments, err := proto.GetDeploymentsFromUrl("http://localhost:8500")
	if err != nil {
		panic(err)
	}
	contracts := proto.Parse(deployments)

	for _, contract := range contracts {
		if contract.PackageName == "accounts" {
			log.Println(proto.Render(contract))
		}
	}
}
