### `concsh`

<!-- REPLACE: go list -f '{{.Doc}}' myitcv.io/g/cmd/concsh -->
`concsh` allows you to concurrently run commands from your shell.
<!-- END -->

```bash
go get myitcv.io/g/cmd/concsh
```

### Usage

<!-- CODE: concsh --help -->
```
concsh -- COMMAND1 ARGS1... --- COMMAND2 ARGS2... --- ...
```
<!-- END -->

All args after the first `--` are then considered as a `---`-separated (notice the extra `-`)
list of commands to be run concurrently. Output from each command (both stdout and stderr)
is output to the `concsh`'s stdout and stderr when a command finishes executing; output is not
interleaved between commands, that is to say output is grouped by command (although the distinction
between stdout and stderr is retained)

The exit code from `concsh` is `0` if all commands succeed without error, or one of the non-zero
exit codes otherwise

### Example

```bash
x=$(go list -f '{{.Dir}}' myitcv.io/g/cmd/concsh)/example
concsh -- $x/timer.sh 1 --- $x/timer.sh 2 --- $x/timer.sh 3 --- $x/timer.sh 4 --- $x/timer.sh 5
```

which gives output similar to:

```
Instance 4 iteration loop 1 (loop delay 0.5s)
Instance 4 iteration loop 2 (loop delay 0.7s)
Instance 4 iteration loop 3 (loop delay 0.0s)
Instance 4 iteration loop 4 (loop delay 0.5s)
Instance 4 iteration loop 5 (loop delay 0.1s)
Instance 3 iteration loop 1 (loop delay 0.2s)
Instance 3 iteration loop 2 (loop delay 0.0s)
Instance 3 iteration loop 3 (loop delay 0.8s)
Instance 3 iteration loop 4 (loop delay 0.1s)
Instance 3 iteration loop 5 (loop delay 0.7s)
Instance 5 iteration loop 1 (loop delay 0.3s)
Instance 5 iteration loop 2 (loop delay 0.5s)
Instance 5 iteration loop 3 (loop delay 0.7s)
Instance 5 iteration loop 4 (loop delay 0.5s)
Instance 5 iteration loop 5 (loop delay 0.2s)
Instance 1 iteration loop 1 (loop delay 0.7s)
Instance 1 iteration loop 2 (loop delay 0.5s)
Instance 1 iteration loop 3 (loop delay 0.5s)
Instance 1 iteration loop 4 (loop delay 0.5s)
Instance 1 iteration loop 5 (loop delay 0.2s)
Instance 2 iteration loop 1 (loop delay 0.7s)
Instance 2 iteration loop 2 (loop delay 0.2s)
Instance 2 iteration loop 3 (loop delay 0.8s)
Instance 2 iteration loop 4 (loop delay 0.6s)
Instance 2 iteration loop 5 (loop delay 0.4s)
```

See how the output is grouped per command.
