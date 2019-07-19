package proto

import (
	"fmt"
	"strings"
)

type method struct {
	Name   string
	Input  string
	Output string
}

func (mtd method) printMethod(prefix string) string {
	return prefix + fmt.Sprintf("rpc %s(%s) returns (%s);", mtd.Name, mtd.Input, mtd.Output) + "\n"
}

type Service struct {
	Comment string
	Name    string
	Methods []method
}

func (srv Service) PrintService() string {
	return srv.printService("")
}

func (srv Service) printService(prefix string) string {
	var builder strings.Builder

	builder.WriteString(prefix + fmt.Sprintf("// %s", srv.Comment) + "\n")
	builder.WriteString(prefix + fmt.Sprintf("service %s {", srv.Name) + "\n")

	for _, mtd := range srv.Methods {
		builder.WriteString(mtd.printMethod(prefix + "\t"))
	}

	builder.WriteString(prefix + "}" + "\n")

	return builder.String()
}
