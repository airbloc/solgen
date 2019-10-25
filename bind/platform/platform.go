package platform

type Platform string

const (
	Ethereum Platform = "ethereum"
	Klaytn   Platform = "klaytn"
)

var Imports = map[Platform]map[string]string{
	Ethereum: MergeImports(AirblocDependencies, EthereumDependencies),
	Klaytn:   MergeImports(AirblocDependencies, KlaytnDependencies),
}

func ManagerImports(platform Platform) map[string]string {
	return map[string]string{
		"wrappers":   "github.com/airbloc/contract-sdk/bind/wrappers",
		"blockchain": "github.com/airbloc/contract-sdk/blockchain",
		"logger":     "github.com/airbloc/logger",
		"common":     Imports[platform]["common"],
		"errors":     "github.com/pkg/errors",
	}
}

func MergeImports(imports ...map[string]string) map[string]string {
	o := make(map[string]string)
	for _, src := range imports {
		for k, v := range src {
			o[k] = v
		}
	}
	return o
}
