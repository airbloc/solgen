package bind

type Customs struct {
	Structs map[string]string `json:"structs"`
	Imports map[string]string `json:"imports"`
	Methods map[string]bool   `json:"methods"`
}
