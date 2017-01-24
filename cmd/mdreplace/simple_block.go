package main

// import (
// 	"fmt"
// 	"io"
// 	"os/exec"
// 	"strings"
// 	"unicode"
// )

// type transFn func(io.Reader) (interface{}, error)

// type simpleBlock struct {
// 	name  string
// 	trans transFn

// 	hPrefix string

// 	// these are effectively state
// 	inCode    bool
// 	cmdString string
// 	cmd       []string
// 	tmpl      string
// }

// func newSimpleBlock(name string, trans transFn) *simpleBlock {
// 	return &simpleBlock{
// 		name:    name,
// 		hPrefix: commStart + " " + name + ":",
// 		trans:   trans,
// 	}
// }

// func (b *simpleBlock) isHeader(line string) bool {
// 	return strings.HasPrefix(line, b.hPrefix)
// }

// func (b *simpleBlock) isEnd(line string) bool {
// 	return strings.TrimRightFunc(line, unicode.IsSpace) == endBlock
// }

// func (b *simpleBlock) parseHeader(line string) error {
// 	cmdStr := strings.TrimSpace(line)
// 	cmdStr = strings.TrimPrefix(cmdStr, b.hPrefix)
// 	cmdStr = strings.TrimSuffix(cmdStr, b.hSuffix)

// 	b.cmdStr = cmdStr

// 	cmd, err := split(cmdStr)
// 	if err != nil {
// 		return err
// 	}

// 	b.cmd = cmd

// 	return nil
// }

// func (b *simpleBlock) reset() {
// 	b.cmdStr = ""
// 	b.cmd = nil
// }

// func (s *simpleBlock) explode() (string, error) {
// 	defer s.basicBlock.reset()

// 	cmd := exec.Command(s.cmd[0], s.cmd[1:]...)
// 	out, err := cmd.CombinedOutput()
// 	if err != nil {
// 		return "", fmt.Errorf("could not execute cmd %v: %v\n%v", s.cmdStr, err, string(out))
// 	}

// 	sp := ""
// 	w := s.wrap(string(out))
// 	if w[len(w)-1] != '\n' {
// 		sp = "\n"
// 	}

// 	return fmt.Sprintf("%v%v%v\n%v%v%v", s.hPrefix, s.cmdStr, s.hSuffix, w, sp, s.end), nil
// }
