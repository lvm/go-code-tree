package codetree

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/mod/modfile"
)

type Module string

func GetModule(dir string) (*Module, error) {
	file, err := os.Open(filepath.Join(dir, "go.mod"))
	if err != nil {
		return nil, err
	}
	defer file.Close()

	b := &bytes.Buffer{}
	io.Copy(b, file)

	f, err := modfile.Parse("go.mod", b.Bytes(), nil)
	if err != nil {
		return nil, err
	}

	mod := Module(f.Module.Mod.Path)
	return &mod, nil
}

func (m *Module) String() string {
	return string(*m)
}

func (m *Module) Basename() string {
	mods := strings.Split(m.String(), "/")
	return mods[len(mods)-1]
}
