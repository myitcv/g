watcher
=======

A Linux-based directory watcher for triggering commands:

```bash
$ ./watcher -help
Command mode:
        ./watcher [-q duration] [-p /path/to/watch] [-i] [-f] [-c] [-k] command_string

Die mode:
        ./watcher -d [-p /path/to/watch] [-f]

  -c=true: clear the screen before running the command
  -d=false: die on first notification; only consider -p and -f flags
  -f=false: whether to follow symlinks or not (recursively) [*]
  -i=true: run command at time zero; only applies when -d not supplied
  -k=true: whether to kill the running command on a new notification; ensures contiguous command calls
  -p="": the path to watch; default is CWD [*]
  -q=1ms: the duration of the 'quiet' window; format is 1s, 10us etc. Min 1 millisecond
  -t=0: the timeout after which a process is killed; not valid with -k

Only options marked with [*] are valid in die mode
```

### Die mode

Useful for one-off triggering of another command when a change is detected in some path.

The following command will block the watcher process until there is a file or directory change
somewhere beneath `/tmp` (recursive watching):

```bash
$ watcher -d -p /tmp && echo 'Hello world'
```

When a change is detected, `watcher` exits with a zero status (assuming there was no errors)
and `Hello world` is printed to the console.

### Command mode

Used to repeatedly run a command when a change is detected in some path.

The following command re-runs `go test` when a change is detected in the current working directory
(the default path assumed if `-p` is omitted):

```bash
$ watcher 'go test'
```

Notice the use of single quotes; the arguments to `watcher` in command mode are equivalent to the
command string one would pass to `bash -c`, indeed this is a good test:

```bash
$ bash -c -- 'go test'
```

Further options can be seen via `watcher -help`
