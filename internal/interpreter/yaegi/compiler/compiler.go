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
// If use unsafe.Offsetof function, it will process it specially.
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
	buf := bytes.NewBuffer(make([]byte, 0, 1024))
	// write package name
	buf.WriteString("package ")
	buf.WriteString(pkg.Name)
	buf.WriteString("\n\n")
	// write import
	if len(imports) != 0 {
		buf.WriteString("import (\n")
		for importLine := range imports {
			buf.WriteString("\t")
			buf.WriteString(importLine)
			buf.WriteString("\n")
		}
		buf.WriteString(")\n\n")
	}
	// write code, use string slice for sort.
	for i := 0; i < len(pkg.GoFiles); i++ {
		filename := filepath.Join(dir, pkg.GoFiles[i])
		offset := offsets[filename]
		buf.WriteString(files[filename][offset:])
		buf.WriteString("\n")
	}
	// format source code
	b, err := format.Source(buf.Bytes())
	if err != nil {
		return "", err
	}
	code := string(b)
	// check need process unsafe
	for i := 0; i < len(pkg.Imports); i++ {
		if pkg.Imports[i] == "unsafe" {
			code = ProcessUnsafeOffsetof(code)
			break
		}
	}
	return code, nil
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

// ProcessUnsafeOffsetof is used to process yaegi code that use unsafe.Offsetof().
// It will replace "unsafe.Offsetof(T{}.A)" to "unsafe.Offsetof(T{}, "A")"
func ProcessUnsafeOffsetof(src string) string {
	const flag = "unsafe.Offsetof("
	buf := strings.Builder{}
	offset := 0
	for {
		index := strings.Index(src[offset:], flag)
		if index == -1 {
			buf.WriteString(src[offset:])
			break
		}
		// write code before flag
		buf.WriteString(src[offset : offset+index+len(flag)])
		// update offset for simplify code
		offset = offset + index + len(flag)
		// get field name
		begin := strings.Index(src[offset:], ".")
		end := strings.Index(src[offset:], ")")
		fieldName := src[offset+begin+1 : offset+end]
		buf.WriteString(src[offset : offset+begin])
		buf.WriteString(", \"")
		// if structure field is unexported,
		// yaegi will add "X" before field name.
		if fieldName[0] >= 'a' && fieldName[0] <= 'z' {
			buf.WriteString("X")
		}
		buf.WriteString(fieldName)
		buf.WriteString("\")")
		// update global offset
		offset = offset + end + 1
	}
	return buf.String()
}
