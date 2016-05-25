package main

import (
	"fmt"
	"sort"
	"strings"

	"github.com/myitcv/g/protobuf/ast"
)

// TODO
//
// We lose spacing (or no spacing) between fields in a message; it should be at most 1 space
// not an enforced 1 space;

func (f *formatter) fmtFile(file *ast.File) {
	f.fmtSyntax(file.Syntax)
	f.fmtPackage(file.Package)
	f.fmtOptions(file.Options)
	f.fmtImports(file.Imports)

	var nodes []ast.Node
	for _, m := range file.Messages {
		nodes = append(nodes, m)
	}
	for _, e := range file.Enums {
		nodes = append(nodes, e)
	}
	for _, s := range file.Services {
		nodes = append(nodes, s)
	}

	f.fmtNodes(nodes)
}

func (f *formatter) fmtSyntax(syntax string) {
	f.Printf("syntax = \"%v\";\n", syntax)
	f.Println()
}

func (f *formatter) fmtPackage(pkg []string) {
	f.Printf("package %v;\n", strings.Join(pkg, "."))

	if len(pkg) > 0 {
		f.Println()
	}
}

func (f *formatter) fmtOptions(options [][2]string) {
	for _, o := range options {
		f.Printf("option %v = %v;\n", o[0], o[1])
	}

	if len(options) > 0 {
		f.Println()
	}
}

func (f *formatter) fmtImports(imports []string) {
	for _, i := range imports {
		f.Printf("import \"%v\";\n", i)
	}

	if len(imports) > 0 {
		f.Println()
	}
}

func (f *formatter) fmtNodes(nodes []ast.Node) {
	sort.Sort(nodeSort(nodes))

	for i, n := range nodes {
		if i != 0 {
			f.Println()
		}
		f.fmtNode(n)
	}
}

func (f *formatter) fmtNode(node ast.Node) {
	switch node := node.(type) {
	case *ast.Message:
		f.fmtMessage(node)
	case *ast.Enum:
		f.fmtEnum(node)
	case *ast.Field:
		f.fmtField(node)
	case *ast.Service:
		f.fmtService(node)
	case *ast.Method:
		f.fmtMethod(node)
	default:
		panic(fmt.Errorf("No formatter for %T", node))
	}
}

func (f *formatter) fmtService(svc *ast.Service) {
	f.Printf("service %v {\n", svc.Name)
	f.indent++

	var nodes []ast.Node
	for _, m := range svc.Methods {
		nodes = append(nodes, m)
	}
	f.fmtNodes(nodes)

	f.indent--
	f.Println("}")
}

func (f *formatter) fmtMethod(meth *ast.Method) {
	f.Printf("rpc %v (%v) returns (%v)", meth.Name, meth.InTypeName, meth.OutTypeName)
	if len(meth.Options) > 0 {
		f.NoIndentPrintf(" {\n")
		f.indent++

		for _, o := range meth.Options {
			f.Printf("option (%v) = %v;\n", o[0], o[1])
		}

		f.indent--
		f.Println("}")
	} else {
		f.NoIndentPrintf(";\n")
	}
}

func (f *formatter) fmtMessage(message *ast.Message) {
	f.Printf("message %v {\n", message.Name)
	f.indent++

	var nodes []ast.Node
	for _, m := range message.Messages {
		nodes = append(nodes, m)
	}
	for _, e := range message.Enums {
		nodes = append(nodes, e)
	}
	for _, field := range message.Fields {
		nodes = append(nodes, field)
	}

	f.fmtNodes(nodes)

	f.indent--
	f.Println("}")
}

func (f *formatter) fmtEnum(enum *ast.Enum) {
	f.Printf("enum %v {\n", enum.Name)
	f.indent++

	for _, v := range enum.Values {
		f.Printf("%v = %v;\n", v.Name, v.Number)
	}

	f.indent--
	f.Println("}")
}

func (f *formatter) fmtField(field *ast.Field) {
	if field.Oneof != nil && f.oneOf == nil {
		f.oneOf = field.Oneof
		f.Printf("oneof %v {\n", field.Oneof.Name)
		f.indent++
	} else if field.Oneof == nil && f.oneOf != nil {
		f.indent--
		f.oneOf = nil
		f.Println("}")
	}

	if field.KeyTypeName != "" {
		f.Printf("map<%v, %v> %v = %v", field.KeyTypeName, field.TypeName, field.Name, field.Tag)
	} else if field.Repeated {
		f.Printf("repeated %v %v = %v", field.TypeName, field.Name, field.Tag)
	} else {
		f.Printf("%v %v = %v", field.TypeName, field.Name, field.Tag)
	}

	if len(field.Options) > 0 {
		f.NoIndentPrintf(" [")
		for i, o := range field.Options {
			if i > 0 {
				f.NoIndentPrintf(", ")
			}
			f.NoIndentPrintf("(%v)=%v", o[0], o[1])
		}
		f.NoIndentPrintf("];\n")
	} else {
		f.NoIndentPrintf(";\n")
	}
}
