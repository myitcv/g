// Copyright (c) 2016 Paul Jolly <paul@myitcv.org.uk>, all rights reserved.
// Use of this document is governed by a license found in the LICENSE document.

package main // import "myitcv.io/g/cmd/newshell"

import (
	"bufio"
	"bytes"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"syscall"
)

var fP = flag.Int64("p", -1, "the PID of the process to walk for bash sub-processes")

const (
	GoVersion       = "GOVERSION"
	MustChangeToDir = "MUST_CHANGE_TO_DIR"
)

func main() {
	flag.Parse()

	runtime.LockOSThread()

	// we need to fail gracefully because the output (and exit code) from this
	// program will not be seen

	// if we have any remaining args they will be treated as the input to exec
	if len(flag.Args()) == 0 {
		return
	}

	if *fP != -1 {
		// we need to try and find the deepst bash child process of the process
		// and get the cwd of the process
		bestPid := uint64(0)
		pids := []uint64{uint64(*fP)}

		for len(pids) > 0 {
			pid := pids[0]
			pids = pids[1:]

			cmd := exec.Command("pgrep", "-x", "bash", "-P", strconv.FormatUint(pid, 10))
			output, err := cmd.CombinedOutput()
			if err != nil {
				ws := cmd.ProcessState.Sys().(syscall.WaitStatus)
				if ws.ExitStatus() != 1 {
					log.Fatalf("Could not run psgrep: %v, %v\n", err, output)
				}
			}

			lineReader := bytes.NewReader(output)
			scanner := bufio.NewScanner(lineReader)
			for scanner.Scan() {
				// each line is a pid
				cPid, err := strconv.ParseUint(scanner.Text(), 10, 64)
				if err != nil {
					log.Fatalf("Could not parseint: %v\n", err)
				}
				bestPid = cPid
				pids = append(pids, cPid)
			}

			if err := scanner.Err(); err != nil {
				log.Fatalf("Could not scan: %v\n", err)
			}
		}

		if bestPid != 0 {
			n, err := os.Readlink(fmt.Sprintf("/proc/%v/cwd", bestPid))
			if err != nil {
				log.Fatalf("Could not read cwd of best pid %v: %v", bestPid, err)
			}

			// we don't care if this fails
			os.Setenv(MustChangeToDir, n)

			gv, err := goVersion(bestPid)
			if err == nil {
				os.Setenv(GoVersion, gv)
			}
		}

	}

	cmd := flag.Args()
	syscall.Exec(cmd[0], cmd, os.Environ())
}

func goVersion(pid uint64) (string, error) {
	mi, err := os.Open(fmt.Sprintf("/proc/%d/mountinfo", pid))
	if err != nil {
		return "", err
	}
	defer mi.Close()

	root := ""

	sc := bufio.NewScanner(mi)

	for sc.Scan() {
		line := sc.Text()
		parts := strings.Fields(line)

		if parts[4] == "/home/myitcv/gos" {
			root = parts[3]
			break
		}
	}

	if strings.HasPrefix(root, "/home/myitcv/.gos/") {
		return strings.TrimPrefix(root, "/home/myitcv/.gos/"), nil
	}

	if root == "/home/myitcv/dev/go" {
		return "gotip", nil
	}

	return "", errors.New("Not mounted or unknown error")
}
