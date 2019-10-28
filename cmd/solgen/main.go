package main

import (
	"encoding/json"
	"io/ioutil"
	"os"

	"github.com/airbloc/solgen/bind/language"
	"github.com/airbloc/solgen/bind/platform"

	"github.com/airbloc/solgen/bind"
	"github.com/airbloc/solgen/deployment"
)

func main() {
	deployments, err := deployment.GetDeploymentsFrom("http://localhost:8500")
	if err != nil {
		panic(err)
	}

	customs := make(map[string]bind.Customs)
	opt, err := ioutil.ReadFile("option_bind_airbloc.json")
	if err != nil {
		panic(err)
	}

	if err := json.Unmarshal(opt, &customs); err != nil {
		panic(err)
	}

	if err := os.RemoveAll("test/bind"); err != nil {
		panic(err)
	}

	if err := os.MkdirAll("test/bind/contracts", os.ModePerm); err != nil {
		panic(err)
	}
	if err := os.MkdirAll("test/bind/managers", os.ModePerm); err != nil {
		panic(err)
	}
	if err := os.MkdirAll("test/bind/wrappers", os.ModePerm); err != nil {
		panic(err)
	}

	for name, deployment := range deployments {
		_, _ = bind.Bind(name, bind.Option{
			ABI:      deployment.RawABI,
			Customs:  customs[name],
			Platform: platform.Klaytn,
			Language: language.Go,
		})
	}
	//bind.Bind()
}
