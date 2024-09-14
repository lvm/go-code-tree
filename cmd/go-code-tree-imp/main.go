package main

import (
	"flag"
	"fmt"
	"go-code-tree/pkg/codetree"
	"log"
)

func main() {
	dir := flag.String("dir", "./", "Directory of the Go project to scan for imports")
	showThirdParty := flag.Bool("third", false, "Show third-party imports")
	scanMocks := flag.Bool("mocks", false, "Scan mock files")
	scanTests := flag.Bool("tests", false, "Scan test files")
	flag.Parse()

	mod, err := codetree.GetModule(*dir)
	if err != nil {
		log.Print("Error reading module name:", err)
		return
	}

	ct := codetree.NewImpTree(*dir, *mod)

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
