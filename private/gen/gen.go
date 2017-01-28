package gen

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"reflect"
	"strings"
	"unicode"

	"github.com/jjeffery/errors"
	"github.com/jjeffery/sqlrow/private/colname"
	"github.com/jjeffery/sqlrow/private/column"
)

// Model contains all of the information required by the template
// to generate code.
type Model struct {
	CommandLine string
	Package     string
	Imports     []*Import
	QueryTypes  []*QueryType
}

// Import describes a single import line required for the generated file.
type Import struct {
	Name string // Local name, or blank
	Path string
}

func (imp *Import) String() string {
	if imp.Name != "" {
		return fmt.Sprintf("%s %s", imp.Name, imp.Path)
	}
	return imp.Path
}

// QueryType contains all the information the template needs
// about a struct type for which methods are generated for
// DB queries.
type QueryType struct {
	TypeName        string
	QuotedTableName string // Table name in quotes
	Singular        string // Describes one instance in error msg
	Plural          string // Describes multiple instances in error msg
	DBField         string // Name of the field of type sqlrow.DB (probably db)
	SchemaField     string // Name of the schema field of type sqlrow.Schema (probably schema)
	RowType         *RowType
	Method          struct {
		Get       bool
		Select    bool
		SelectOne bool
		Insert    bool
		Update    bool
		Delete    bool
		Upsert    bool
	}
}

// RowType contains all the information the template needs about
// a struct type that is used to represent a single DB table row.
type RowType struct {
	Name      string
	IDArgs    string   // for function arguments specifying primary key ID field(s)
	IDParams  string   // for function parameters specifying primary key ID field(s)
	IDKeyvals string   // for log messages specifying primary key ID field(s)
	LogProps  []string // for error messages
}

// Parse the file, and any other related files and build the
// model, which can be used to generate the code.
func Parse(filename string) (*Model, error) {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, filename, nil, 0)
	if err != nil {
		return nil, errors.Wrap(err, "cannot parse file").With(
			"filename", filename,
		)
	}

	model := &Model{
		Package: file.Name.Name,
	}
	ir, err := newImportResolver(file.Imports)
	if err != nil {
		return nil, err
	}

	for _, decl := range file.Decls {
		if genDecl, ok := decl.(*ast.GenDecl); ok {
			for _, spec := range genDecl.Specs {
				if typeSpec, ok := spec.(*ast.TypeSpec); ok {
					structType, ok := typeSpec.Type.(*ast.StructType)
					if !ok {
						continue
					}
					queryType, err := newQueryType(file, ir, typeSpec, structType)
					if err != nil {
						return nil, err
					}
					if queryType != nil {
						model.QueryTypes = append(model.QueryTypes, queryType)
					}
				}
			}
		}
	}
	model.Imports = ir.Imports()

	return model, nil
}

func newQueryType(file *ast.File, ir *importResolver, typeSpec *ast.TypeSpec, structType *ast.StructType) (*QueryType, error) {
	var rowTypeField *ast.Field
	var dbField *ast.Field
	var schemaField *ast.Field
	var methods string
	var tableName string
	var singular string
	var plural string

	const rowTypeFieldName = "rowType"
	const dbTypeName = "sqlrow.DB"
	const schemaTypeName = "sqlrow.Schema"

	// Use a local import resolver for resolving field type names that will not
	// be used in generated code.
	localIR, err := newImportResolver(file.Imports)
	if err != nil {
		return nil, err
	}

	for _, field := range structType.Fields.List {
		fieldTypeName := localIR.exprString(field.Type)
		if field.Tag != nil {
			tag := reflect.StructTag(stripQuotes(field.Tag.Value))
			if v := tag.Get("methods"); v != "" {
				methods = v
			}
			if v := tag.Get("table"); v != "" {
				tableName = v
			}
			if v := tag.Get("singular"); v != "" {
				singular = v
			}
			if v := tag.Get("plural"); v != "" {
				plural = v
			}
		}
		if fieldTypeName == dbTypeName {
			dbField = field
			continue
		}
		if fieldTypeName == schemaTypeName {
			schemaField = field
			continue
		}
		for _, name := range field.Names {
			if name.Name == rowTypeFieldName {
				rowTypeField = field
				continue
			}
		}
	}
	if rowTypeField == nil {
		// not a struct defining a set of queries
		return nil, nil
	}
	if dbField == nil {
		return nil, errors.New("missing field").With(
			"struct", typeSpec.Name.Name,
			"type", dbTypeName,
		)
	}
	if schemaField == nil {
		return nil, errors.New("missing field").With(
			"struct", typeSpec.Name.Name,
			"type", schemaTypeName,
		)
	}

	if tableName == "" {
		tableName = colname.Snake.ColumnName(typeSpec.Name.Name)
		tableName = toPlural(tableName)
	}

	if singular == "" {
		singular = typeSpec.Name.Name
	}

	if plural == "" {
		plural = toPlural(singular)
	}

	if methods == "" {
		methods = "get,select,selectOne,insert,update,delete,upsert"
	}

	// at this point we have a struct that describes a query type
	queryType := &QueryType{
		TypeName:        typeSpec.Name.Name,
		QuotedTableName: fmt.Sprintf("%q", tableName),
		Singular:        singular,
		Plural:          plural,
		DBField:         dbField.Names[0].Name,
		SchemaField:     schemaField.Names[0].Name,
	}
	for _, method := range strings.Split(methods, ",") {
		lmethod := strings.ToLower(strings.TrimSpace(method))
		switch lmethod {
		case "get":
			queryType.Method.Get = true
		case "select":
			queryType.Method.Select = true
		case "selectone":
			queryType.Method.SelectOne = true
		case "insert":
			queryType.Method.Insert = true
		case "update":
			queryType.Method.Update = true
		case "upsert":
			queryType.Method.Upsert = true
		case "delete":
			queryType.Method.Delete = true
		default:
			return nil, errors.New("unknown method").With(
				"method", method,
			)
		}
	}

	rowType, err := newRowType(file, ir, rowTypeField.Type)
	if err != nil {
		return nil, err
	}

	queryType.RowType = rowType
	return queryType, nil
}

func toPlural(s string) string {
	return s + "s"
}

func stripRowTypeExpr(expr ast.Expr) ast.Expr {
	for {
		switch v := expr.(type) {
		case *ast.Ident:
			return v
		case *ast.SelectorExpr:
			return v
		case *ast.StarExpr:
			expr = v.X
		case *ast.ParenExpr:
			expr = v.X
		default:
			return nil
		}
	}
}

func newRowType(file *ast.File, ir *importResolver, typeExpr ast.Expr) (*RowType, error) {
	typeExpr = stripRowTypeExpr(typeExpr)
	if typeExpr == nil {
		return nil, errors.New("unexpected type for rowType")
	}
	if _, ok := typeExpr.(*ast.SelectorExpr); ok {
		return nil, errors.New("doesn't handle row types in other packages yet, row type needs to be in the same package and file")
	}
	rowTypeIdent, ok := typeExpr.(*ast.Ident)
	if !ok {
		// should not get here, checked earlier
		return nil, errors.New("unexpected row type")
	}

	rowTypeName := rowTypeIdent.Name
	structType := findStructType(file, rowTypeName)
	if structType == nil {
		return nil, errors.New("cannot find row type").With(
			"name", rowTypeName,
		)
	}

	var pkParams []string
	var pkKeyvals []string
	var pkArgs []string
	var kvArgs []string

	for _, field := range structType.Fields.List {
		var tagInfo column.TagInfo
		if field.Tag != nil {
			tag := reflect.StructTag(stripQuotes(field.Tag.Value))
			tagInfo = column.ParseTag(tag)
		}
		if tagInfo.Ignore {
			continue
		}
		if tagInfo.PrimaryKey {
			for _, fieldName := range field.Names {
				paramName := lowerCaseField(fieldName.Name)
				pkArgs = append(pkArgs, paramName)
				pkKeyvals = append(pkKeyvals, fmt.Sprintf("%q", paramName))
				pkKeyvals = append(pkKeyvals, paramName)
				kvArgs = append(kvArgs, fieldName.Name)
				typeName := ir.exprString(field.Type)
				pkParams = append(pkParams, fmt.Sprintf("%s %s", paramName, typeName))
			}
		}
		if tagInfo.NaturalKey {
			for _, ident := range field.Names {
				kvArgs = append(kvArgs, ident.Name)
			}
		}
	}

	rowType := &RowType{
		Name:      rowTypeName,
		IDParams:  strings.Join(pkParams, ", "),
		IDArgs:    strings.Join(pkArgs, ", "),
		IDKeyvals: strings.Join(pkKeyvals, ", ") + ",",
		LogProps:  kvArgs,
	}

	return rowType, nil
}

func findStructType(file *ast.File, name string) *ast.StructType {
	for _, decl := range file.Decls {
		if genDecl, ok := decl.(*ast.GenDecl); ok {
			for _, spec := range genDecl.Specs {
				if typeSpec, ok := spec.(*ast.TypeSpec); ok {
					structType, ok := typeSpec.Type.(*ast.StructType)
					if !ok {
						continue
					}
					if typeSpec.Name.Name == name {
						return structType
					}
				}
			}
		}
	}
	return nil
}

func stripQuotes(s string) string {
	return strings.Trim(s, "`")
}

func lowerCaseField(s string) string {
	var buf bytes.Buffer
	var metLower bool
	for _, ch := range s {
		if !metLower && unicode.IsUpper(ch) {
			buf.WriteRune(unicode.ToLower(ch))
			continue
		}
		metLower = true
		buf.WriteRune(ch)
	}
	return buf.String()
}
