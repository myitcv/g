package main

import "fmt"

type importPaths []string

func (i *importPaths) Set(value string) error {
	*i = append(*i, value)
	return nil
}

func (i *importPaths) String() string {
	return fmt.Sprint(*i)
}
