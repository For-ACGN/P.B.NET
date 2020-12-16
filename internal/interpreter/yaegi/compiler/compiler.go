package compiler

import (
	"bytes"
	"fmt"
	"go/build"
	"go/format"
	"io/ioutil"
	"path/filepath"
	"strings"
)

// Compile is used to import go source code in directory to a package
// and generate it to a single go file for yaegi.
func Compile(ctx *build.Context, dir string) (string, error) {
	pkg, err := ctx.ImportDir(dir, 0)
	if err != nil {
		return "", err
	}
	// check error in go files
	if len(pkg.InvalidGoFiles) != 0 {
		return "", fmt.Errorf("find error in file: %s", pkg.InvalidGoFiles[0])
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
	// store files to offsets map
	offsets := make(map[string]int, len(pkg.GoFiles))
	for i := 0; i < len(pkg.GoFiles); i++ {
		offsets[filepath.Join(dir, pkg.GoFiles[i])] = 0
	}
	for _, pos := range pkg.ImportPos {
		for i := 0; i < len(pos); i++ {
			offset := pos[i].Offset
			filename := pos[i].Filename
			if offset <= offsets[filename] {
				continue
			}
			// find bracket
			file := files[filename]
			begin := strings.LastIndex(files[filename][:offset], "import")
			if strings.Contains(file[begin:offset], "(") {
				offsets[filename] = begin + strings.Index(file[begin:], ")") + 1
			} else {
				offsets[filename] = begin + strings.Index(file[begin:], "\n") + 1
			}

		}
	}
	// generate source code
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
	// write code
	for filename, offset := range offsets {
		content := files[filename]
		// no import
		if offset == 0 {
			// first search package
			idx := strings.Index(content, "package")
			// then search newline
			offset = idx + strings.Index(content[idx:], "\n")
		}
		fmt.Println("===")
		fmt.Println(content[offset:])
		fmt.Println("===")
		code.WriteString(content[offset:])
		code.WriteString("\n")
	}
	// format
	src, err := format.Source(code.Bytes())
	if err != nil {
		return "", err
	}
	return string(src), err
}
