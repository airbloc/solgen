package bind

import (
	"github.com/airbloc/solgen/bind/language"
	"github.com/airbloc/solgen/bind/platform"
)

type Option struct {
	ABI      string
	Customs  Customs
	Platform platform.Platform
	Language language.Language
}
