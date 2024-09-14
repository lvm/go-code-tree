package main

import (
	"flag"
	"fmt"
	"go-code-tree/pkg/codetree"
	"log"
)

func main() {
	dir := flag.String("dir", "./", "Directory of the Go project to scan code")
	scanMocks := flag.Bool("mocks", false, "Scan mock files")
	scanTests := flag.Bool("tests", false, "Scan test files")
	flag.Parse()

	mod, err := codetree.GetModule(*dir)
	if err != nil {
		log.Print("Error reading module name:", err)
		return
	}

	ft := codetree.NewFuncTree(*dir, *mod)

	funcs, err := ft.GetFuncs(*scanMocks, *scanTests)
	if err != nil {
		log.Print("Error getting funcs:", err)
		return
	}
	ft.Funcs = funcs

	graph, err := ft.GenerateGraph()
	if err != nil {
		log.Print("Failed to generate graph:", err)
		return
	}

	fmt.Println(graph)
}
