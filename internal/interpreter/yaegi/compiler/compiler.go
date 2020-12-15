package compiler

import (
	"bytes"
	"fmt"
	"go/build"
	"io/ioutil"
	"path/filepath"
	"strings"
)

// Load is used to import go source code in directory to a package and
// generate it to a single go file for yaegi.
func Compile(ctx *build.Context, dir string) (string, error) {
	pkg, err := ctx.ImportDir(dir, 0)
	if err != nil {
		return "", err
	}

	// read go files
	files := make(map[string]string)
	for i := 0; i < len(pkg.GoFiles); i++ {
		path := filepath.Join(dir, pkg.GoFiles[i])
		data, err := ioutil.ReadFile(path)
		if err != nil {
			return "", err
		}
		files[path] = string(data)
	}

	// ok
	fmt.Println(pkg.Name)
	fmt.Println(pkg.Imports)

	// get imports
	imports := make(map[string]struct{}, len(pkg.Imports))
	for _, posList := range pkg.ImportPos {
		for _, pos := range posList {
			file := files[pos.Filename]
			end := strings.Index(file[pos.Offset:], "\n")
			importLine := file[pos.Offset : pos.Offset+end]
			imports[importLine] = struct{}{}
		}
	}
	fmt.Println(imports)

	// store files to offsets map
	offsets := make(map[string]int, len(pkg.GoFiles))
	for i := 0; i < len(pkg.GoFiles); i++ {
		offsets[filepath.Join(dir, pkg.GoFiles[i])] = 0
	}
	for _, pos := range pkg.ImportPos {
		for i := 0; i < len(pos); i++ {
			if pos[i].Offset > offsets[pos[i].Filename] {
				offsets[pos[i].Filename] = pos[i].Offset
			}
		}
	}

	fmt.Println(offsets)

	code := bytes.NewBuffer(make([]byte, 0, 1024))
	// write package
	code.WriteString("package ")
	code.WriteString(pkg.Name)
	code.WriteString("\n\n")
	// write import
	if len(imports) != 0 {
		code.WriteString("import (\n")
		for importLine := range imports {
			code.WriteString("\t")
			code.WriteString(importLine)
			code.WriteString("\n")
		}
		code.WriteString(")\n\n")
	}
	// write codes

	return code.String(), err
}
