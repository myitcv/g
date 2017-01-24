## `mdreplace`

<!-- __TEMPLATE: go run *.go -help
```
{{ . }}
```
-->

```
<!-- this contains a comment -->
First example
<!-- END -->
```
<!-- END -->

<!-- __JSON: go list -json
{{ .Doc }}
-->

Second example
<!-- END -->

<!-- __LINES: echo -e "hello\nworld"
{{ range . }}
* {{ . }}
{{ end }}
-->

Third example
<!-- END -->
