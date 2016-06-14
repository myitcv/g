package main

import "github.com/myitcv/g/protobuf/ast"

type nodeSort []ast.Node

func (a nodeSort) Len() int      { return len(a) }
func (a nodeSort) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a nodeSort) Less(i, j int) bool {
	return a[i].Pos().Offset < a[j].Pos().Offset
}
