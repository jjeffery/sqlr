package codegen

import (
	"fmt"
	"go/ast"
	"path"
	"strings"
)

type importResolver struct {
	imports []*ast.ImportSpec
	used    map[string]*Import
}

func (r *importResolver) Resolve(name string) *Import {
	if imp, ok := r.used[name]; ok {
		return imp
	}

	// strips the quotes from the import path and returns the base name
	pathBase := func(p string) string {
		return path.Base(strings.TrimPrefix(strings.TrimSuffix(p, `"`), `"`))
	}
	// strips the quotes from the import path and returns the base name without any extension
	// (good for import paths like "gopkg.in/xyz/abc.v1")
	pathBaseWithoutExtension := func(p string) string {
		p = pathBase(p)
		return strings.TrimSuffix(p, path.Ext(p))
	}

	tests := []func(*ast.ImportSpec) bool{
		// import has matching explicit name
		func(is *ast.ImportSpec) bool {
			return is.Name != nil && is.Name.Name == name
		},
		// import has matching import base name
		func(is *ast.ImportSpec) bool {
			if is.Name != nil {
				return false
			}
			return pathBase(is.Path.Value) == name
		},
		// import has matching import base name without extension
		func(is *ast.ImportSpec) bool {
			if is.Name != nil {
				return false
			}
			return pathBaseWithoutExtension(is.Path.Value) == name
		},
		// import base name contains the string somewhere
		func(is *ast.ImportSpec) bool {
			if is.Name != nil {
				return false
			}
			return strings.Contains(pathBase(is.Path.Value), name)
		},
	}
	// Search for an import whose name matches.
	for _, test := range tests {
		for _, importSpec := range r.imports {
			if test(importSpec) {
				imp := &Import{
					Path: importSpec.Path.Value,
				}
				if importSpec.Name != nil {
					imp.Name = importSpec.Name.Name
				}

				r.used[name] = imp
				return imp
			}
		}
	}
	return nil
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

func newImportResolver(imports []*ast.ImportSpec) (*importResolver, error) {
	for _, importSpec := range imports {
		if importSpec.Name != nil && importSpec.Name.Name == "." {
			return nil, fmt.Errorf("dot imports are not supported: . %v", importSpec.Path.Value)
		}
	}
	return &importResolver{
		imports: imports,
		used:    make(map[string]*Import),
	}, nil
}

func notExpecting(nodeType string) {
	msg := fmt.Sprintf("not expecting node type of %s", nodeType)
	panic(msg)
}

func notImplemented(nodeType string) {
	msg := fmt.Sprintf("handling of node type not implemented: %s", nodeType)
	panic(msg)
}
