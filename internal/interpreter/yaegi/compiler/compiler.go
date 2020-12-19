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

// Compile is used to import go source code in directory to a
// package and generate it to a single go file for yaegi.
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
	files := make(map[string]string, len(pkg.GoFiles))
	for i := 0; i < len(pkg.GoFiles); i++ {
		path := filepath.Join(dir, pkg.GoFiles[i])
		data, err := ioutil.ReadFile(path) // #nosec
		if err != nil {
			return "", err
		}
		files[path] = string(data)
	}
	// read imports, pkg.Imports not include package aliases
	// like "fmt" and fmt1 "fmt", so we need process it manually.
	// use map to remove duplicate package.
	imports := make(map[string]struct{}, len(pkg.Imports))
	for _, importPos := range pkg.ImportPos {
		for _, pos := range importPos {
			file := files[pos.Filename]
			end := strings.Index(file[pos.Offset:], "\n")
			line := file[pos.Offset : pos.Offset+end]
			imports[line] = struct{}{}
		}
	}
	// calculate offset of code
	offsets := calculateOffsets(pkg, dir, files)
	// generate source code
	code := bytes.NewBuffer(make([]byte, 0, 1024))
	// write package name
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
	// write code, use string slice for sort.
	for i := 0; i < len(pkg.GoFiles); i++ {
		filename := filepath.Join(dir, pkg.GoFiles[i])
		offset := offsets[filename]
		code.WriteString(files[filename][offset:])
		code.WriteString("\n")
	}
	// format source code
	src, err := format.Source(code.Bytes())
	if err != nil {
		return "", err
	}
	return string(src), err
}

func calculateOffsets(pkg *build.Package, dir string, files map[string]string) map[string]int {
	// store files to offsets map
	offsets := make(map[string]int, len(pkg.GoFiles))
	// calculate offset after package and import.
	for _, pos := range pkg.ImportPos {
		for i := 0; i < len(pos); i++ {
			offset := pos[i].Offset
			filename := pos[i].Filename
			if offset <= offsets[filename] {
				continue
			}
			file := files[filename]
			begin := strings.LastIndex(files[filename][:offset], "import")
			// find bracket if it is exist
			if strings.Contains(file[begin:offset], "(") {
				offsets[filename] = begin + strings.Index(file[begin:], ")") + 1
			} else {
				offsets[filename] = begin + strings.Index(file[begin:], "\n") + 1
			}
		}
	}
	for i := 0; i < len(pkg.GoFiles); i++ {
		filename := filepath.Join(dir, pkg.GoFiles[i])
		if offsets[filename] != 0 {
			continue
		}
		// if no import, first search package, then search newline
		content := files[filename]
		idx := strings.Index(content, "package")
		offsets[filename] = idx + strings.Index(content[idx:], "\n")
	}
	return offsets
}
