package language

import (
	"github.com/airbloc/solgen/bind/template"
	"github.com/airbloc/solgen/utils"
	"github.com/ethereum/go-ethereum/accounts/abi"
)

type Lang string

const (
	Go   Lang = "golang"
	Java Lang = "java"
)

var BindType = map[Lang]func(kind abi.Type, structs map[string]*template.Struct) string{
	Go:   bindTypeGo,
	Java: bindTypeJava,
}

// bindTopicType is a set of type binders that convert Solidity types to some
// supported programming language topic types.
var BindTopicType = map[Lang]func(kind abi.Type, structs map[string]*template.Struct) string{
	Go:   bindTopicTypeGo,
	Java: bindTopicTypeJava,
}

// bindStructType is a set of type binders that convert Solidity tuple types to some supported
// programming language struct definition.
var BindStructType = map[Lang]func(kind abi.Type, structs map[string]*template.Struct) string{
	Go:   bindStructTypeGo,
	Java: bindStructTypeJava,
}

// namedType is a set of functions that transform language specific types to
// named versions that my be used inside method names.
var NamedType = map[Lang]func(string, abi.Type) string{
	Go:   func(string, abi.Type) string { panic("this shouldn't be needed") },
	Java: namedTypeJava,
}

// methodNormalizer is a name transformer that modifies Solidity method names to
// conform to target language naming concentions.
var MethodNormalizer = map[Lang]func(string) string{
	Go:   abi.ToCamelCase,
	Java: utils.Decapitalise,
}
