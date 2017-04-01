package codegen

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/jjeffery/errors"
)

type packageInfo struct {
	importSpec  *ast.ImportSpec
	packagePath string
	pkg         *ast.Package
}

type importResolver struct {
	packages map[string]*packageInfo
	imports  []*ast.ImportSpec
	used     map[string]*Import
}

func (r *importResolver) Resolve(name string) *Import {
	if imp, ok := r.used[name]; ok {
		return imp
	}

	if pkgInfo, ok := r.packages[name]; ok {
		imp := &Import{
			Path: pkgInfo.importSpec.Path.Value,
		}
		if pkgInfo.importSpec.Name != nil {
			imp.Name = pkgInfo.importSpec.Name.Name
		}

		r.used[name] = imp
		return imp
	}

	return nil
}

func (r *importResolver) ParsePackage(name string) (*ast.Package, error) {
	pkgInfo := r.packages[name]
	if pkgInfo == nil {
		return nil, errors.New("unknown package").With(
			"selector", name,
		)
	}
	if pkgInfo.pkg == nil {
		fset := token.NewFileSet()
		pkgs, err := parser.ParseDir(fset, pkgInfo.packagePath, filter, 0)
		if err != nil {
			return nil, err
		}
		for name, pkg := range pkgs {
			if name == "main" || strings.HasSuffix(name, "_test") {
				continue
			}
			pkgInfo.pkg = pkg
			break
		}
	}
	if pkgInfo.pkg == nil {
		return nil, errors.New("cannot parse package").With(
			"selector", name,
		)
	}
	return pkgInfo.pkg, nil
}

func (r *importResolver) exprString(t ast.Expr) string {
	if t == nil {
		return ""
	}
	switch v := t.(type) {
	case *ast.BadExpr:
		return "<bad-expr>"
	case *ast.Ident:
		return v.Name
	case *ast.Ellipsis:
		return fmt.Sprintf("...%s", r.exprString(v.Elt))
	case *ast.BasicLit:
		// does not appear in method declarations
		return v.Value
	case *ast.FuncLit:
		notExpecting("FuncLit")
	case *ast.CompositeLit:
		notExpecting("CompositeLit")
	case *ast.ParenExpr:
		return fmt.Sprintf("(%s)", r.exprString(v.X))
	case *ast.SelectorExpr:
		r.Resolve(r.exprString(v.X))
		return fmt.Sprintf("%s.%s", r.exprString(v.X), v.Sel.Name)
	case *ast.IndexExpr:
		return fmt.Sprintf("%s[%s]", r.exprString(v.X), r.exprString(v.Index))
	case *ast.SliceExpr:
		if v.Slice3 {
			return fmt.Sprintf("%s[%s:%s]", r.exprString(v.X), r.exprString(v.Low), r.exprString(v.High))
		}
		return fmt.Sprintf("%s[%s:%s:%s]", r.exprString(v.X), r.exprString(v.Low), r.exprString(v.High), r.exprString(v.Max))
	case *ast.TypeAssertExpr:
		notExpecting("TypeAssertExpr")
	case *ast.CallExpr:
		notExpecting("CallExpr")
	case *ast.StarExpr:
		return fmt.Sprintf("*%s", r.exprString(v.X))
	case *ast.UnaryExpr:
		return fmt.Sprintf("%s%s", v.Op.String(), r.exprString(v.X))
	case *ast.BinaryExpr:
		return fmt.Sprintf("%s %s %s", r.exprString(v.X), v.Op.String(), r.exprString(v.Y))
	case *ast.KeyValueExpr:
		return fmt.Sprintf("%s: %s", r.exprString(v.Key), r.exprString(v.Value))
	case *ast.ArrayType:
		return fmt.Sprintf("[%s]%s", r.exprString(v.Len), r.exprString(v.Elt))
	case *ast.StructType:
		notImplemented("StructType")
	case *ast.FuncType:
		notImplemented("FuncType")
	case *ast.InterfaceType:
		notImplemented("InterfaceType")
	case *ast.MapType:
		return fmt.Sprintf("map[%s]%s", r.exprString(v.Key), r.exprString(v.Value))
	case *ast.ChanType:
		switch v.Dir {
		case ast.SEND:
			return fmt.Sprintf("chan<- %s", r.exprString(v.Value))
		case ast.RECV:
			return fmt.Sprintf("<-chan %s", r.exprString(v.Value))
		default:
			return fmt.Sprintf("chan %s", r.exprString(v.Value))
		}
	}

	panic(fmt.Sprintf("unknown ast.Expr: %v", t))
}

func (r *importResolver) Imports() []*Import {
	var imports []*Import
	for _, imp := range r.used {
		imports = append(imports, imp)
	}
	return imports
}

// parse the package enough to find out its name
func filter(fileInfo os.FileInfo) bool {
	if strings.HasSuffix(fileInfo.Name(), "_test.go") {
		return false
	}
	return true
}

func newImportResolver(imports []*ast.ImportSpec) (*importResolver, error) {
	resolver := &importResolver{
		packages: make(map[string]*packageInfo),
		imports:  imports,
		used:     make(map[string]*Import),
	}

	for _, importSpec := range imports {
		if importSpec.Name != nil && importSpec.Name.Name == "." {
			return nil, fmt.Errorf("dot imports are not supported: . %v", importSpec.Path.Value)
		}
		packagePath, err := findPackageDirectory(importSpec.Path.Value)
		if err != nil {
			return nil, err
		}
		fset := token.NewFileSet()
		pkgs, err := parser.ParseDir(fset, packagePath, filter, parser.PackageClauseOnly)
		if err != nil {
			return nil, errors.Wrap(err, "cannot parse package").With(
				"package", importSpec.Path.Value,
			)
		}
		// there should only be one item in the slice
		for name := range pkgs {
			if name == "main" {
				continue
			}
			pkgInfo := &packageInfo{
				importSpec:  importSpec,
				packagePath: packagePath,
			}
			if importSpec.Name != nil {
				resolver.packages[importSpec.Name.Name] = pkgInfo
			} else {
				resolver.packages[name] = pkgInfo
			}
		}
	}
	return resolver, nil
}

func notExpecting(nodeType string) {
	msg := fmt.Sprintf("not expecting node type of %s", nodeType)
	panic(msg)
}

func notImplemented(nodeType string) {
	msg := fmt.Sprintf("handling of node type not implemented: %s", nodeType)
	panic(msg)
}

// findPackageDirectory finds the directory on the GOPATH that
// corresponds with the directory specification.
// BUG(jpj): I have to think that there is a std library function for
// doing this.
func findPackageDirectory(importPath string) (string, error) {
	// strip leading and trailing '"'
	importPath = strings.TrimPrefix(strings.TrimSuffix(importPath, `"`), `"`)

	// directories for searching for the import path
	goDirs := []string{runtime.GOROOT()}
	{
		gopath := os.Getenv("GOPATH")
		if gopath == "" {
			var home string
			if runtime.GOOS == "windows" {
				home = os.Getenv("USERPROFILE")
			} else {
				home = os.Getenv("HOME")
			}
			if home != "" {
				gopath = filepath.Join(home, "go")
			}
		}
		if gopath != "" {
			goDirs = append(goDirs, filepath.SplitList(gopath)...)
		}
	}
	for _, goDir := range goDirs {
		importDir := filepath.Join(goDir, "src", filepath.FromSlash(importPath))
		fileInfo, err := os.Stat(importDir)
		if err != nil {
			if !os.IsNotExist(err) {
				// error other than does not exist
				return "", errors.Wrap(err, "cannot stat").With(
					"path", importDir,
				)
			}
			// file/directory does not exist
			fileInfo = nil
		}
		if fileInfo != nil && fileInfo.IsDir() {
			// found the directory
			return importDir, nil
		}
	}
	return "", &os.PathError{
		Op:   "findPackageDirectory",
		Path: importPath,
		Err:  os.ErrNotExist,
	}
}
