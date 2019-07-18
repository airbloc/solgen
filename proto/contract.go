package proto

type rpc struct {
	Name   string
	Input  string
	Output string
}

type service struct {
	Name string
	Rpcs []rpc
}

type arg struct {
	Name     string
	Repeated bool
	Type     string
	Count    int
}

type message struct {
	Comment string
	Name    string
	Args    []arg
}

type contract struct {
	PackageName string
	Services    []service
	Messages    []message
}
