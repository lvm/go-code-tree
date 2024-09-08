package main

import (
	"bufio"
	"flag"
	"fmt"
	"go/parser"
	"go/token"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type (
	ICodeTree interface {
		GetImports(scanMocks, scanTests bool) (Relation, error)
		GenerateGraph(showThirdParty bool) (string, error)
	}

	Relation map[string][]string
	Module   string
	CodeTree struct {
		Root    string
		Module  Module
		Imports Relation
	}
)

func getModule(dir string) (*Module, error) {
	file, err := os.Open(filepath.Join(dir, "go.mod"))
	if err != nil {
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()

		if strings.HasPrefix(line, "module ") {
			mod := strings.TrimSpace(strings.TrimPrefix(line, "module "))
			module := Module(mod)
			return &module, nil
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return nil, nil
}

func parseImports(filePath string) ([]string, error) {
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, filePath, nil, parser.ImportsOnly)
	if err != nil {
		return nil, err
	}

	var imports []string = make([]string, len(node.Imports))
	for i, imp := range node.Imports {
		imports[i] = strings.Trim(imp.Path.Value, `"`)
	}

	return imports, nil
}

func NewCodeTree(root string, module Module) *CodeTree {
	return &CodeTree{
		Root:   root,
		Module: module,
	}
}

func (m *Module) String() string {
	return string(*m)
}

func (m *Module) getRepo() string {
	mods := strings.Split(m.String(), "/")
	return mods[len(mods)-1]
}

func (ct *CodeTree) GetImports(scanMocks, scanTests bool) (Relation, error) {
	imports := make(Relation)
	walker := func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if strings.HasSuffix(path, "_test.go") && !scanTests || strings.Contains(path, "_mock.go") && !scanMocks {
			return nil
		}

		if !info.IsDir() && filepath.Ext(path) == ".go" {
			imps, err := parseImports(path)
			if err != nil {
				return err
			}
			imports[path] = imps
		}

		return nil
	}

	if err := filepath.Walk(ct.Root, walker); err != nil {
		return nil, err
	}

	return imports, nil
}

func (ct *CodeTree) GenerateGraph(showThirdParty bool) (string, error) {
	var (
		sb          strings.Builder
		imports     Relation = ct.Imports
		mod         string   = ct.Module.String()
		repo        string   = ct.Module.getRepo()
		dirContent  Relation = make(Relation)
		fileImports Relation = make(Relation)
	)

	for _, line := range []string{
		fmt.Sprintf("digraph \"%s\" {\n", repo),
		"  rankdir=TB;\n",
		"  node [shape=box, color=\"burlywood\", style=\"filled\", fillcolor=\"seashell\"];\n",
		"  edge [color=\"burlywood\"];\n",
	} {
		sb.WriteString(line)
	}

	files := make([]string, 0, len(imports))
	for file := range imports {
		files = append(files, file)
	}
	sort.Strings(files)

	for _, file := range files {
		dir := filepath.Dir(file)
		dirContent[file] = []string{dir}

		imps := imports[file]
		localImports := make(map[string]struct{})
		thirdImports := make(map[string]struct{})

		for _, imp := range imps {
			isLocal := false

			if strings.HasPrefix(imp, mod) {
				imp = strings.Replace(imp, mod, repo, 1)
				localImports[imp] = struct{}{}
				isLocal = true
			}

			if !isLocal {
				thirdImports[imp] = struct{}{}
			}
		}

		for imp := range localImports {
			fileImports[file] = append(fileImports[file], imp)
		}

		if showThirdParty && len(thirdImports) > 0 {
			for imp := range thirdImports {
				fileImports[file] = append(fileImports[file], imp)
			}
		}
	}

	for dir, content := range dirContent {
		for _, file := range content {
			sb.WriteString(fmt.Sprintf("  \"%s\" [color=\"orange\", style=\"filled\", fillcolor=\"lightyellow\"];\n", file))
			sb.WriteString(fmt.Sprintf("  \"%s\" [color=\"seagreen\", style=\"filled\", fillcolor=\"mintcream\"];\n", dir))
			sb.WriteString(fmt.Sprintf("  \"%s\" -> \"%s\" [color=\"orange\"];\n", file, dir))
		}
	}

	for file, imports := range fileImports {
		for _, imp := range imports {
			sb.WriteString(fmt.Sprintf("  \"%s\" -> \"%s\" [color=\"seagreen\"];\n", file, imp))
		}
	}

	sb.WriteString("}")

	return sb.String(), nil
}

func main() {
	dir := flag.String("dir", "./", "Directory of the Go project to scan for imports")
	showThirdParty := flag.Bool("third", false, "Show third-party imports")
	scanMocks := flag.Bool("mocks", false, "Scan mock files")
	scanTests := flag.Bool("tests", false, "Scan test files")
	flag.Parse()

	mod, err := getModule(*dir)
	if err != nil {
		log.Print("Error reading module name:", err)
		return
	}

	ct := NewCodeTree(*dir, *mod)

	imports, err := ct.GetImports(*scanMocks, *scanTests)
	if err != nil {
		log.Print("Error getting imports:", err)
		return
	}
	ct.Imports = imports

	graph, err := ct.GenerateGraph(*showThirdParty)
	if err != nil {
		log.Print("Failed to generate graph:", err)
		return
	}

	fmt.Println(graph)
}
