package main

import (
	"flag"
	"fmt"
	"log"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/juju/errgo"
	"github.com/kr/fs"

	"golang.org/x/exp/inotify"
)

// TODO implement timeout for killing long-running process

var fIgnorePaths ignorePaths

var fQuiet = flag.Duration("q", time.Millisecond, "the duration of the 'quiet' window; format is 1s, 10us etc. Min 1 millisecond")
var fPath = flag.String("p", "", "the path to watch; default is CWD [*]")
var fFollow = flag.Bool("f", false, "whether to follow symlinks or not (recursively) [*]")
var fDie = flag.Bool("d", false, "die on first notification; only consider -p and -f flags")
var fClearScreen = flag.Bool("c", true, "clear the screen before running the command")
var fInitial = flag.Bool("i", true, "run command at time zero; only applies when -d not supplied")
var fTimeout = flag.Duration("t", 0, "the timeout after which a process is killed; not valid with -k")
var fKill = flag.Bool("k", true, "whether to kill the running command on a new notification; ensures contiguous command calls")

const (
	GitDir = ".git"
)

var GloballyIgnoredDirs = []string{GitDir}

func init() {
	flag.Var(&fIgnorePaths, "I", "Paths to ignore. Absolute paths are absolute to the path; relative paths can match anywhere in the tree")
}

type ignorePaths []string

func (i *ignorePaths) Set(value string) error {
	*i = append(*i, value)
	return nil
}

func (i *ignorePaths) String() string {
	return fmt.Sprint(*i)
}

func showUsage() {
	fmt.Fprintf(os.Stderr, "Command mode:\n\t%v [-q duration] [-p /path/to/watch] [-i] [-f] [-c] [-k] CMD ARG1 ARG2...\n\nDie mode:\n\t%v -d [-p /path/to/watch] [-f]\n\n", os.Args[0], os.Args[0])
	flag.PrintDefaults()
	fmt.Fprintf(os.Stderr, "\nOnly options marked with [*] are valid in die mode\n")
	os.Exit(1)
}

func main() {
	flag.Usage = showUsage
	flag.Parse()

	path := *fPath
	if path == "" {
		path = "."
	}
	path, err := filepath.Abs(path)
	if err != nil {
		panic(err)
	}
	_, err = os.Stat(path)
	if err != nil {
		log.Fatalf("Could not stat -p supplied path [%v]: %v\n", path, err)
	}

	if !*fDie {
		if *fQuiet < 0 {
			log.Fatalf("Quiet window duration [%v] must be positive\n", *fQuiet)
		}
		if *fTimeout < 0 {
			log.Fatalf("Command timeout duration [%v] must be positive\n", *fTimeout)
		}
	}

	if !*fDie && *fQuiet < time.Millisecond {
		log.Fatalln("Quiet time period must be at least 1 millisecond")
	}

	watchFlags := inotify.IN_ATTRIB | inotify.IN_CREATE | inotify.IN_DELETE | inotify.IN_DELETE_SELF |
		inotify.IN_MODIFY | inotify.IN_MOVE_SELF | inotify.IN_MOVE

	if !*fFollow {
		watchFlags |= inotify.IN_DONT_FOLLOW
	}

	w, err := newWatcher()
	if err != nil {
		log.Fatalf("Could not create a watcher: %v\n", err)
	}
	defer w.close()

	w.dirFlags = watchFlags
	w.fileFlags = watchFlags
	w.kill = *fKill
	w.timeout = *fTimeout
	w.quiet = *fQuiet
	w.initial = *fInitial
	w.command = flag.Args()
	w.clearScreen = *fClearScreen
	w.ignorePaths = append(fIgnorePaths, GloballyIgnoredDirs...)
	w.absPath = path

	if *fDie {
		w.watchOnce(path)
	} else {
		w.watchLoop(path)
	}
}

type watcher struct {
	iwatcher    *inotify.Watcher
	dirFlags    uint32
	fileFlags   uint32
	kill        bool
	clearScreen bool
	command     []string
	ignorePaths []string
	absPath     string
	initial     bool
	timeout     time.Duration
	quiet       time.Duration
}

func newWatcher() (*watcher, error) {
	res := &watcher{}
	w, err := inotify.NewWatcher()
	if err != nil {
		return nil, errgo.NoteMask(err, "Could not create a watcher")
	}
	res.iwatcher = w
	return res, nil
}

func (w *watcher) close() error {
	err := w.iwatcher.Close()
	if err != nil {
		return errgo.NoteMask(err, "Could not close inotify watcher")
	}
	return nil
}

func (w *watcher) recursiveWatchAdd(p string) error {
	// p is a path; may or may not be a directory

	walker := fs.Walk(p)
WalkLoop:
	for walker.Step() {
		if err := walker.Err(); err != nil {
			// TODO better than silently continue?
			continue
		}
		s := walker.Stat()
		flags := w.fileFlags
		if s.IsDir() {
			flags = w.dirFlags

			for _, s := range w.ignorePaths {
				rel, _ := filepath.Rel(w.absPath, walker.Path())

				if filepath.IsAbs(s) {
					nonAbs := strings.TrimPrefix(s, "/")

					if nonAbs == rel {
						walker.SkipDir()
						continue WalkLoop
					}

				} else {
					if strings.HasSuffix(rel, s) {
						walker.SkipDir()
						continue WalkLoop
					}
				}
			}
		}
		err := w.iwatcher.AddWatch(walker.Path(), flags)
		if err != nil {
			// TODO anything better to do that just swallow it?
		}
	}
	return nil
}

func (w *watcher) recursiveWatchRemove(p string) error {
	// TODO make this recursive if needs be?
	err := w.iwatcher.RemoveWatch(p)
	if err != nil {
		// TODO anything better to do that just swallow it?
	}
	return nil
}

func (w *watcher) watchOnce(p string) {
	w.recursiveWatchAdd(p)

	retVal := 0

	select {
	case _ = <-w.iwatcher.Event:
		// TODO handle the queue overflow? probably not needed
	case _ = <-w.iwatcher.Error:
		// TODO handle the queue overflow
		retVal = 1
	}
	os.Exit(retVal)
}

func (w *watcher) watchLoop(p string) {
	w.recursiveWatchAdd(p)

	comm := w.commandLoop()

	var timeout <-chan time.Time

	for {
		select {
		case <-timeout:
			timeout = nil
			comm <- struct{}{}
		case e := <-w.iwatcher.Event:
			// TODO handle the queue overflow... this could happen
			// if we do get queue overflow, might need to look at putting
			// subscriptions on another goroutine, buffering the adds/
			// removes somehow
			if e.Mask&inotify.IN_CREATE != 0 {
				err := w.recursiveWatchAdd(e.Name)
				if err != nil {
					// TODO anything better to do that just swallow it?
				}
			} else if e.Mask&(inotify.IN_DELETE|inotify.IN_DELETE_SELF) != 0 {
				err := w.recursiveWatchRemove(e.Name)
				if err != nil {
					// TODO anything better to do that just swallow it?
				}
			}

			// whatever the type of event, now time to fire across to the
			// command goroutine
			if timeout == nil {
				timeout = time.After(time.Millisecond * 200)
			}
		case _ = <-w.iwatcher.Error:
			// TODO handle the queue overflow
		}
	}
}

func (w *watcher) commandLoop() chan struct{} {
	c := make(chan struct{})
	var t *time.Timer

	if w.initial {
		t = time.NewTimer(0)
	} else if w.quiet > 0 {
		t = time.NewTimer(math.MaxInt64)
	}

	go func() {
		args := []string{"-O", "globstar", "-c", "--"}
		args = append(args, w.command...)
		var command *exec.Cmd
		cmdDone := make(chan struct{})

		runCmd := func() {
			if command != nil {
				if !w.kill {
					// command is still running and we were told not to kill it
					return
				}
				err := command.Process.Kill()
				if err != nil {
					// TODO assume this would only fail if the process has
					// already died... hence silently ignore
				}
			}
			command = exec.Command("bash", args...)
			command.Stdout = os.Stdout
			command.Stderr = os.Stderr
			if w.clearScreen {
				fmt.Printf("\033[2J")
			}
			err := command.Start()
			if err != nil {
				log.Fatalf("We could not run the command provided: %v\n", err)
			}
			go func(c *exec.Cmd) {
				err := c.Wait()
				if err != nil {
					// TODO we need to handle this better... silently ignore?
				}
				cmdDone <- struct{}{}
			}(command)
		}

		for {
			select {
			case <-cmdDone:
				command = nil
			case <-t.C:
				// we had a timeout at the end of a quiet window
				// we need to call the command (and kill if required)
				runCmd()
			case <-c:
				// we got a tick
				if w.quiet > 0 {
					// we need to obey the quiet window
					t.Reset(w.quiet)
				} else {
					// we need to call the command (and kill if required)
					runCmd()
				}
			}
		}
	}()

	return c
}
