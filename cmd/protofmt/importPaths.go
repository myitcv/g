// Copyright (c) 2016 Paul Jolly <paul@myitcv.org.uk>, all rights reserved.
// Use of this document is governed by a license found in the LICENSE document.

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
