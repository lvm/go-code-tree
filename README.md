# `go-code-tree`

The goal of this project is to provide a visual overview of a Go project using a DOT diagram. Particularly, this tool only analyses directories and the files contained, then parses the imports of each go file and builds a map of imports (local and third party). By default it ignores third party imports, mock and test files.

## Usage

```
go-code-tree -h
Usage of go-code-tree:
  -dir string
        Directory of the Go project to scan for imports (default "./")
  -mocks
        Scan mock files
  -tests
        Scan test files
  -third
        Show third-party imports
```

## Build

```
go build -ldflags "-s -w" -o go-code-tree .
```

## Example 

When running the script allowing third party dependencies (see [Suggested dependencies](#suggested-dependencies)): 
```
$ go-code-tree -dir go-code-tree -third | dot -Tpng -ogct.gv.png
```

It'll generate this diagram:

![](media/gct.gv.png)

## Suggested dependencies

* [Graphviz](https://graphviz.org/)

It's not required to have it installed to use this tool, because `go-code-tree` only prints diagrams, but it's useful to have `graphviz` installed to _see_ the diagram.

### macOS

```
brew install graphviz
```

### GNU/Linux (Debian based distros)

```
apt install graphviz
```

## LICENSE 

See [LICENSE](LICENSE)

## Related

* [go-callvis](https://github.com/ondrajz/go-callvis)
* [go-dep-graph](https://github.com/paetzke/go-dep-graph)