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

func ManagerImports(plat Platform) map[string]string {
	return MergeImports(map[string]string{
		"wrappers": "github.com/airbloc/airbloc-go/bind/wrappers",
		"common":   Imports[plat]["common"],
		"errors":   "github.com/pkg/errors",
	}, AirblocDependencies)
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
