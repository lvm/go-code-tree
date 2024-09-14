package codetree

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
)

type (
	FuncTree struct {
		Root   string
		Module Module
		Funcs  Relation
	}
)

func NewFuncTree(root string, module Module) *FuncTree {
	return &FuncTree{
		Root:   root,
		Module: module,
		Funcs:  nil,
	}
}

func parseExpr(e ast.Expr) string {
	switch xpr := e.(type) {
	case *ast.ArrayType:
		return fmt.Sprintf("[]%v", parseExpr(xpr.Elt)) // TODO: complete if slice has len -> [N]arr
	case *ast.ChanType:
		var arr string
		switch xpr.Dir {
		case ast.SEND:
			arr = "chan<-"
		case ast.RECV:
			arr = "<-chan"
		default:
			arr = "chan"
		}
		return fmt.Sprintf("%v%v", arr, xpr.Value)
	case *ast.Ident:
		return fmt.Sprintf("%v", xpr.Name)
	case *ast.MapType:
		return fmt.Sprintf("map[%s]%s", parseExpr(xpr.Key), parseExpr(xpr.Value))
	case *ast.SelectorExpr:
		return fmt.Sprintf("%s.%s", parseExpr(xpr.X), parseExpr(xpr.Sel))
	case *ast.StarExpr:
		return fmt.Sprintf("*%v", xpr.X)
	default:
		return ""
	}

}

func parseFieldList(fl ast.FieldList) string {
	var sb strings.Builder
	var comma string = ""
	for i, field := range fl.List {
		var nsb strings.Builder
		var ncomma string = ""
		for j, name := range field.Names {
			if j > 0 {
				ncomma = ","
			}
			nsb.WriteString(fmt.Sprintf("%s%s ", ncomma, name.Name))
		}

		if i > 0 {
			comma = ", "
		}
		sb.WriteString(fmt.Sprintf("%s%s%s", comma, nsb.String(), parseExpr(field.Type)))
	}

	return sb.String()
}

func parseFuncs(filePath string) ([]string, error) {
	var (
		sb    strings.Builder
		funcs []string
	)

	inspector := func(n ast.Node) bool {
		if fn, ok := n.(*ast.FuncDecl); ok {
			sb.WriteString("func ")

			if fn.Recv != nil {
				sb.WriteString(fmt.Sprintf("(%s) ", parseFieldList(*fn.Recv)))
			}

			sb.WriteString(fn.Name.Name)

			if fn.Type.Params != nil {
				sb.WriteString(fmt.Sprintf("(%s)", parseFieldList(*fn.Type.Params)))
			}

			if fn.Type.Results != nil {
				sb.WriteString(fmt.Sprintf(" (%s)", parseFieldList(*fn.Type.Results)))
			}

			sb.WriteString("\n")
		}
		return true
	}

	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, filePath, nil, parser.AllErrors)
	if err != nil {
		return nil, err
	}

	ast.Inspect(node, inspector)

	for _, fn := range strings.Split(sb.String(), "\n") {
		if len(fn) > 0 {
			funcs = append(funcs, fn)
		}
	}

	return funcs, nil
}

func (ft *FuncTree) GetFuncs(scanMocks, scanTests bool) (Relation, error) {
	funcs := make(Relation)
	walker := func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if strings.HasSuffix(path, "_test.go") && !scanTests || strings.HasSuffix(path, "_mock.go") && !scanMocks {
			return nil
		}

		if !info.IsDir() && filepath.Ext(path) == ".go" {
			fns, err := parseFuncs(path)
			if err != nil {
				return err
			}
			funcs[path] = fns
		}

		return nil
	}

	if err := filepath.Walk(ft.Root, walker); err != nil {
		return nil, err
	}

	return funcs, nil
}

func (ft *FuncTree) GenerateGraph() (string, error) {
	var (
		sb        strings.Builder
		fileFuncs Relation        = ft.Funcs
		rels      map[string]bool = make(map[string]bool)
	)

	for _, line := range []string{
		fmt.Sprintf("digraph \"%s\" {\n", ft.Module.Basename()),
		"  rankdir=LR;\n",
		"  node [shape=box, color=\"burlywood\", style=\"filled\", fillcolor=\"seashell\"];\n",
		"  edge [color=\"burlywood\"];\n",
	} {
		sb.WriteString(line)
	}

	for file, funcs := range fileFuncs {
		path := strings.Split(file, string(filepath.Separator))
		basename := filepath.Base(file)
		for i := 0; i < len(path)-1; i++ {
			left := path[i]
			right := path[i+1]
			rels[fmt.Sprintf(" \"%s\" -> \"%s\" [color=\"orange\", style=\"filled\", fillcolor=\"lightyellow\"];\n", left, right)] = true
		}

		rels[fmt.Sprintf("  \"%s\" [color=\"seagreen\", style=\"filled\", fillcolor=\"mintcream\"];\n", basename)] = true
		for _, fn := range funcs {
			rels[fmt.Sprintf(" \"%s\" [color=\"dodgerblue4\", style=\"filled\", fillcolor=\"aliceblue\"];\n", fn)] = true
			rels[fmt.Sprintf(" \"%s\" -> \"%s\" [color=\"seagreen\", style=\"filled\", fillcolor=\"mintcream\"];\n", basename, fn)] = true
		}
	}

	for rel := range rels {
		sb.WriteString(rel)
	}

	sb.WriteString("}")

	return sb.String(), nil
}
