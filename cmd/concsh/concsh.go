package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"syscall"
)

// all args after the first -- are then considered as a --- (notice the extra -) separated
// list of commands to be run concurrently
//
// output is interleaved on a line-by-line basis
//
// there is no shell evaluation of arguments
//
// TODO could support shell evalulation?
// TODO improve panics; some situations we might be able to better detect/handle?
//
// exit code is 0 if all commands succeed without error; one of the non-zero exit codes otherwise

type result struct {
	exitCode int
	lines    []outLine
}

type outLine struct {
	err  bool
	line string
}

func main() {
	s := 0

	for i, v := range os.Args {
		if v == "--" {
			s = i
			break
		}
	}

	results := make(chan result)
	nr := 0

	var args []string

	for _, v := range os.Args[s+1:] {
		if v == "---" {
			nr += runCmd(args, results)
			args = nil
		} else {
			args = append(args, v)
		}
	}

	// in case we did not have a final ---
	nr += runCmd(args, results)

	if nr == 0 {
		fmt.Println("No commands to execute")
		return
	}

	exitCode := 0

	for res := range results {

		for _, v := range res.lines {
			if v.err {
				fmt.Fprint(os.Stderr, v.line)
			} else {
				fmt.Fprint(os.Stdout, v.line)
			}
		}

		if res.exitCode != 0 {
			exitCode = res.exitCode
		}

		nr--
		if nr == 0 {
			break
		}
	}

	os.Exit(exitCode)
}

func runCmd(args []string, results chan result) int {
	res := 0

	if len(args) > 0 {
		res = 1
		go runCmdImpl(args, results)
	}

	return res
}

func runCmdImpl(args []string, results chan result) {

	cmd := exec.Command(args[0], args[1:]...)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		panic(err)
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		panic(err)
	}

	readDone := make(chan struct{})

	outres := make(chan string)
	errres := make(chan string)

	go read(stdout, outres, readDone)
	go read(stderr, errres, readDone)

	// because we have one process, two pipes to read from
	pc := 2

	err = cmd.Start()
	if err != nil {
		panic(err)
	}

	var lines []outLine

	for {
		select {
		case <-readDone:
			pc--
		case s := <-outres:
			lines = append(lines, outLine{
				line: s,
			})
		case s := <-errres:
			lines = append(lines, outLine{
				err:  true,
				line: s,
			})
		}

		if pc == 0 {
			break
		}
	}

	res := result{
		lines: lines,
	}

	err = cmd.Wait()
	if err != nil {
		ee, ok := err.(*exec.ExitError)
		if !ok {
			panic(err)
		}

		exitCode := 0

		switch ws := ee.Sys().(type) {
		case syscall.WaitStatus:
			exitCode = ws.ExitStatus()
		default:
			panic(fmt.Errorf("Need to add case for %T", ws))
		}

		res.exitCode = exitCode
	}

	results <- res
}

func read(in io.ReadCloser, res chan string, done chan struct{}) {
	b := bufio.NewReader(in)

	for {
		line, err := b.ReadString('\n')

		if err != nil {
			if err == io.EOF {
				if line != "" {
					res <- line + "\n"
				}
			}

			// notice we are ignoring io errors... because any other
			// errors will be handled by the Wait on the cmd with a non-zero
			// exit code

			break
		}

		res <- line
	}

	done <- struct{}{}
}
