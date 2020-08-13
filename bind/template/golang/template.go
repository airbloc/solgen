package golang

import (
	"strings"

	"github.com/airbloc/solgen/bind/template/golang/contracts"
	"github.com/airbloc/solgen/bind/template/golang/managers"
)

func GetContractTemplate() string {
	return strings.Join([]string{
		contracts.Contract,
		contracts.Caller,
		contracts.Transactor,
		contracts.Filterer,
	}, "\n")
}

func GetManagerTamplate() string {
	return managers.Managers
}
