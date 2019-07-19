package main

import (
	"github.com/frostornge/solgen/bind"
	"github.com/frostornge/solgen/deployments"
	"github.com/frostornge/solgen/proto"
)

func main() {
	contracts, err := deployments.GetDeploymentsFromUrl("http://localhost:8500")
	if err != nil {
		panic(err)
	}

	if err := proto.GenerateBind("test/proto", contracts); err != nil {
		panic(err)
	}

	if err := bind.GenerateBind("test/bind", contracts); err != nil {
		panic(err)
	}
}
