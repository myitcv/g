package main

import (
	"fmt"
	"io"
	"strings"

	"github.com/myitcv/g/protobuf/ast"
	"github.com/myitcv/g/protobuf/parser"
)

type formatter struct {
	files []string
	importPaths

	indent int

	output io.Writer

	// TODO this is a bit gross - we can only be in one oneOf at any
	// point in time... seems hacky to store the state here (for indenting)
	oneOf *ast.Oneof
}

func (f *formatter) fmt() {

	fs, err := parser.ParseFiles(f.files, f.importPaths)
	if err != nil {
		panic(err)
	}

	var fmtFiles []*ast.File

	for _, astFile := range fs.Files {
		for _, file := range f.files {
			if file == astFile.Name {
				fmtFiles = append(fmtFiles, astFile)
			}
		}
	}

	for _, file := range fmtFiles {
		f.fmtFile(file)
	}
}

func (f *formatter) Println(a ...interface{}) {
	fmt.Fprintf(f.output, strings.Repeat("\t", f.indent))
	fmt.Fprintln(f.output, a...)
}

func (f *formatter) Printf(format string, a ...interface{}) {
	fmt.Fprintf(f.output, strings.Repeat("\t", f.indent)+format, a...)
}

func (f *formatter) NoIndentPrintf(format string, a ...interface{}) {
	fmt.Fprintf(f.output, format, a...)
}
