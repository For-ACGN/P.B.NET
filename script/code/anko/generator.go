package anko

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"project/internal/system"
)

func exportDeclaration(root, path, dir, init string) (string, error) {
	// package file path
	packages, err := parseDir(filepath.Join(root, strings.ReplaceAll(dir, "$", "")))
	if err != nil {
		return "", err
	}
	name := getPackageName(packages)
	// walk files
	constants := make(map[string]struct{})
	variables := make(map[string]struct{})
	types := make(map[string]struct{})
	functions := make(map[string]struct{})
	for _, file := range packages[name].Files {
		for _, decl := range file.Decls {
			switch decl := decl.(type) {
			case *ast.GenDecl:
				switch decl.Tok {
				case token.CONST:
					exportValues(decl, constants)
				case token.VAR:
					exportValues(decl, variables)
				case token.TYPE:
					exportTypes(decl, types)
				}
			case *ast.FuncDecl:
				exportFunction(decl, functions)
			}
		}
	}
	if dir[0] == '$' {
		path = "project/" + path
	}
	cs := sortStringMap(constants)
	vs := sortStringMap(variables)
	ts := sortStringMap(types)
	fs := sortStringMap(functions)
	return generateCode(path, name, init, cs, vs, ts, fs), nil
}

func isGoFile(info os.FileInfo) bool {
	if info.IsDir() {
		return false
	}
	name := info.Name()
	if name == "fuzz.go" {
		return false
	}
	if strings.HasSuffix(name, "_test.go") {
		return false
	}
	if strings.HasPrefix(name, "example_") {
		return false
	}
	return true
}

func parseDir(dir string) (map[string]*ast.Package, error) {
	return parser.ParseDir(token.NewFileSet(), dir, isGoFile, parser.ParseComments)
}

func getPackageName(packages map[string]*ast.Package) string {
	var pkgName string
loop:
	for pn := range packages {
		switch {
		case pn == "main":
		case strings.HasSuffix(pn, "_test"):
		default:
			pkgName = pn
			break loop
		}
	}
	return pkgName
}

func isDeprecated(text string) bool {
	for _, item := range [...]string{
		"Deprecated:",
		"Deprecated.",
	} {
		if strings.Contains(text, item) {
			return true
		}
	}
	return false
}

func exportValues(decl *ast.GenDecl, m map[string]struct{}) {
	if isDeprecated(decl.Doc.Text()) {
		return
	}
	for _, spec := range decl.Specs {
		vs := spec.(*ast.ValueSpec)
		if isDeprecated(vs.Doc.Text()) {
			continue
		}
		for _, name := range vs.Names {
			// skip some special variables
			if name.Name == "ErrTrailingComma" {
				continue
			}
			if name.IsExported() {
				m[name.Name] = struct{}{}
			}
		}
	}
}

func exportTypes(decl *ast.GenDecl, m map[string]struct{}) {
	if isDeprecated(decl.Doc.Text()) {
		return
	}
	for _, spec := range decl.Specs {
		ts := spec.(*ast.TypeSpec)
		if isDeprecated(ts.Doc.Text()) {
			continue
		}
		if ts.Name.IsExported() {
			m[ts.Name.Name] = struct{}{}
		}
	}
}

func exportFunction(decl *ast.FuncDecl, m map[string]struct{}) {
	if isDeprecated(decl.Doc.Text()) {
		return
	}
	if decl.Recv != nil {
		return
	}
	if decl.Name.IsExported() {
		m[decl.Name.Name] = struct{}{}
	}
}

func sortStringMap(m map[string]struct{}) []string {
	s := make([]string, 0, len(m))
	for k := range m {
		s = append(s, k)
	}
	sort.Strings(s)
	return s
}

func generateCode(path, name, init string, constants, vars, types, fns []string) string {
	const template = `
func init%s() {
	env.Packages["%s"] = map[string]reflect.Value{
		// define constants
%s
		// define variables
%s
		// define functions
%s	}
	var (
%s	)
	env.PackageTypes["%s"] = map[string]reflect.Type{
%s	}
}
`
	const (
		tabs = "\t\t"

		// "Compare": reflect.ValueOf(bytes.Compare),
		valFormat = tabs + `"%s": reflect.ValueOf(%s.%s),` + "\n"

		// conn net.Conn
		vpFormat = tabs + "%s %s.%s\n"

		// "Conn": reflect.TypeOf(&conn).Elem(),
		typeFormat = tabs + `"%s": reflect.TypeOf(&%s).Elem(),` + "\n"
	)

	// constants
	buf := new(bytes.Buffer)
	for _, c := range constants {
		_, _ = fmt.Fprintf(buf, valFormat, c, name, c)
	}
	cs := buf.String()

	// variables
	buf.Reset()
	for _, v := range vars {
		_, _ = fmt.Fprintf(buf, valFormat, v, name, v)
	}
	vs := buf.String()

	// functions
	buf.Reset()
	for _, fn := range fns {
		_, _ = fmt.Fprintf(buf, valFormat, fn, name, fn)
	}
	fs := buf.String()

	// prepare var buffer for struct and interface
	vpBuf := new(bytes.Buffer)
	buf.Reset()
	for _, typ := range types {
		// "ReadWriter" -> "readWriter"
		varName := strings.ToLower(typ[0:1]) + typ[1:]
		_, _ = fmt.Fprintf(vpBuf, vpFormat, varName, name, typ)
		_, _ = fmt.Fprintf(buf, typeFormat, typ, varName)
	}
	vpStr := vpBuf.String()
	ts := buf.String()
	return fmt.Sprintf(template, init, path, cs, vs, fs, vpStr, path, ts)
}

func formatCodeAndSave(t *testing.T, code, path string) {
	data, err := format.Source([]byte(code))
	require.NoError(t, err)
	fmt.Println(string(data))
	err = system.WriteFile(path, data)
	require.NoError(t, err)
}
