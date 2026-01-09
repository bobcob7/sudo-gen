// Package copy implements the deep copy code generation subtool.
package copy

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/bobcob7/sudo-gen/internal/codegen"
)

// Subtool implements the copy code generator.
type Subtool struct {
	MethodName string
}

// Name returns the subtool name.
func (s *Subtool) Name() string { return "copy" }

// Description returns the subtool description.
func (s *Subtool) Description() string {
	return "Generate deep copy methods for structs"
}

// Run executes the copy code generation.
func (s *Subtool) Run(cfg codegen.GeneratorConfig) error {
	methodName := s.MethodName
	if methodName == "" {
		methodName = "Copy"
	}
	g := &generator{
		cfg:        cfg,
		methodName: methodName,
		fset:       token.NewFileSet(),
		imports:    make(map[string]string),
		processed:  make(map[string]bool),
	}
	return g.run()
}

type generator struct {
	cfg        codegen.GeneratorConfig
	methodName string
	pkg        *ast.Package
	fset       *token.FileSet
	imports    map[string]string
	processed  map[string]bool
}

func (g *generator) run() error {
	if err := g.parsePackage(); err != nil {
		return err
	}
	return g.generateForType(g.cfg.TypeName)
}

func (g *generator) parsePackage() error {
	pkgs, err := parser.ParseDir(g.fset, g.cfg.SourceDir, nil, parser.ParseComments)
	if err != nil {
		return fmt.Errorf("parsing directory: %w", err)
	}
	for name, pkg := range pkgs {
		if !strings.HasSuffix(name, "_test") {
			g.pkg = pkg
			break
		}
	}
	if g.pkg == nil {
		return fmt.Errorf("no non-test package found in %s", g.cfg.SourceDir)
	}
	return nil
}

func (g *generator) generateForType(typeName string) error {
	structType, err := g.findStruct(typeName)
	if err != nil {
		return err
	}
	data, err := g.buildTemplateData(typeName, structType)
	if err != nil {
		return fmt.Errorf("building template data: %w", err)
	}
	return g.writeOutput(typeName, data)
}

func (g *generator) findStruct(typeName string) (*ast.StructType, error) {
	var structType *ast.StructType
	for _, file := range g.pkg.Files {
		ast.Inspect(file, func(n ast.Node) bool {
			ts, ok := n.(*ast.TypeSpec)
			if !ok || ts.Name.Name != typeName {
				return true
			}
			if st, ok := ts.Type.(*ast.StructType); ok {
				structType = st
				g.collectFileImports(file)
			}
			return false
		})
		if structType != nil {
			break
		}
	}
	if structType == nil {
		return nil, fmt.Errorf("type %s not found or is not a struct", typeName)
	}
	return structType, nil
}

func (g *generator) collectFileImports(file *ast.File) {
	for _, imp := range file.Imports {
		path := strings.Trim(imp.Path.Value, `"`)
		alias := ""
		if imp.Name != nil {
			alias = imp.Name.Name
		}
		g.imports[path] = alias
	}
}

func (g *generator) buildTemplateData(typeName string, st *ast.StructType) (templateData, error) {
	g.processed[typeName] = true
	fields := g.analyzeFields(st)
	imports := g.collectRequiredImports(fields)
	nestedTypes, err := g.collectNestedTypes(fields)
	if err != nil {
		return templateData{}, err
	}
	return templateData{
		Package:     g.pkg.Name,
		TypeName:    typeName,
		MethodName:  g.methodName,
		Fields:      fields,
		Imports:     imports,
		NestedTypes: nestedTypes,
	}, nil
}

func (g *generator) analyzeFields(st *ast.StructType) []fieldInfo {
	fields := make([]fieldInfo, 0, len(st.Fields.List))
	for _, field := range st.Fields.List {
		if len(field.Names) == 0 {
			continue
		}
		for _, name := range field.Names {
			if !ast.IsExported(name.Name) {
				continue
			}
			fi := fieldInfo{
				Name:     name.Name,
				Type:     exprToString(field.Type),
				TypeExpr: field.Type,
			}
			g.analyzeType(field.Type, &fi)
			fields = append(fields, fi)
		}
	}
	return fields
}

func (g *generator) analyzeType(expr ast.Expr, fi *fieldInfo) {
	switch t := expr.(type) {
	case *ast.StarExpr:
		fi.IsPointer = true
		fi.ElemType = exprToString(t.X)
		if ident, ok := t.X.(*ast.Ident); ok && !isBasicType(ident.Name) {
			fi.StructTypeName = ident.Name
			fi.NeedsDeep = true
		} else {
			fi.NeedsDeep = needsDeepCopy(t.X)
		}
	case *ast.ArrayType:
		fi.IsSlice = true
		fi.ElemType = exprToString(t.Elt)
		switch elt := t.Elt.(type) {
		case *ast.Ident:
			if !isBasicType(elt.Name) {
				fi.StructTypeName = elt.Name
				fi.NeedsDeep = true
			}
		case *ast.StarExpr:
			if ident, ok := elt.X.(*ast.Ident); ok && !isBasicType(ident.Name) {
				fi.StructTypeName = ident.Name
				fi.SliceElemIsPtr = true
				fi.NeedsDeep = true
			}
		default:
			fi.NeedsDeep = needsDeepCopy(t.Elt)
		}
	case *ast.MapType:
		fi.IsMap = true
		fi.KeyType = exprToString(t.Key)
		fi.ValueType = exprToString(t.Value)
		if fi.ValueType == "any" || fi.ValueType == "interface{}" {
			fi.NeedsDeep = true
			return
		}
		switch val := t.Value.(type) {
		case *ast.Ident:
			if !isBasicType(val.Name) {
				fi.StructTypeName = val.Name
				fi.NeedsDeep = true
			}
		case *ast.StarExpr:
			if ident, ok := val.X.(*ast.Ident); ok && !isBasicType(ident.Name) {
				fi.StructTypeName = ident.Name
				fi.NeedsDeep = true
			}
		default:
			fi.NeedsDeep = needsDeepCopy(t.Value)
		}
	case *ast.StructType:
		fi.IsStruct = true
	case *ast.Ident:
		if !isBasicType(t.Name) {
			fi.IsStruct = true
			fi.StructTypeName = t.Name
		}
	case *ast.SelectorExpr:
		pkg, ok := t.X.(*ast.Ident)
		if !ok {
			fi.IsStruct = true
			return
		}
		if pkg.Name == "time" && t.Sel.Name == "Time" {
			return
		}
		fi.IsStruct = true
	}
}

func (g *generator) collectNestedTypes(fields []fieldInfo) ([]templateData, error) {
	var nested []templateData
	seen := make(map[string]bool)
	for _, f := range fields {
		if f.StructTypeName == "" || seen[f.StructTypeName] || g.processed[f.StructTypeName] {
			continue
		}
		seen[f.StructTypeName] = true
		st, err := g.findStruct(f.StructTypeName)
		if err != nil {
			continue
		}
		data, err := g.buildTemplateData(f.StructTypeName, st)
		if err != nil {
			return nil, err
		}
		data.IsNestedType = true
		nested = append(nested, data)
		// Flatten: also add nested types from this type
		nested = append(nested, data.NestedTypes...)
		data.NestedTypes = nil // Clear to avoid duplication in template
	}
	return nested, nil
}

func (g *generator) collectRequiredImports(fields []fieldInfo) []codegen.ImportInfo {
	needed := make(map[string]string)
	for _, f := range fields {
		if f.IsSlice || f.IsMap {
			g.collectImportsFromType(f.TypeExpr, needed)
		}
	}
	for _, f := range fields {
		if f.IsMap && !f.NeedsDeep {
			needed["maps"] = ""
			break
		}
	}
	imports := make([]codegen.ImportInfo, 0, len(needed))
	for path, alias := range needed {
		imports = append(imports, codegen.ImportInfo{Path: path, Alias: alias})
	}
	return imports
}

func (g *generator) collectImportsFromType(expr ast.Expr, needed map[string]string) {
	switch t := expr.(type) {
	case *ast.SelectorExpr:
		pkg, ok := t.X.(*ast.Ident)
		if !ok {
			return
		}
		for path, alias := range g.imports {
			pkgName := alias
			if pkgName == "" {
				pkgName = filepath.Base(path)
			}
			if pkgName == pkg.Name {
				needed[path] = alias
				break
			}
		}
	case *ast.StarExpr:
		g.collectImportsFromType(t.X, needed)
	case *ast.ArrayType:
		g.collectImportsFromType(t.Elt, needed)
	case *ast.MapType:
		g.collectImportsFromType(t.Key, needed)
		g.collectImportsFromType(t.Value, needed)
	}
}

func (g *generator) writeOutput(typeName string, data templateData) error {
	baseName := strings.TrimSuffix(g.cfg.SourceFile, ".go")
	outputFile := filepath.Join(g.cfg.OutputDir, baseName+"_copy.go")
	gen := codegen.NewTemplateGenerator(templateFuncs())
	if err := gen.GenerateFile(outputFile, copyTemplate, data); err != nil {
		return err
	}
	if g.cfg.GenerateTest {
		testFile := filepath.Join(g.cfg.OutputDir, baseName+"_copy_test.go")
		return gen.GenerateFile(testFile, copyTestTemplate, data)
	}
	return nil
}

type templateData struct {
	Package      string
	TypeName     string
	MethodName   string
	Fields       []fieldInfo
	Imports      []codegen.ImportInfo
	NestedTypes  []templateData
	IsNestedType bool
}

type fieldInfo struct {
	Name           string
	Type           string
	TypeExpr       ast.Expr
	IsPointer      bool
	IsSlice        bool
	IsMap          bool
	IsStruct       bool
	ElemType       string
	KeyType        string
	ValueType      string
	NeedsDeep      bool
	StructTypeName string
	SliceElemIsPtr bool
}

func templateFuncs() template.FuncMap {
	return template.FuncMap{
		"lower": strings.ToLower,
	}
}

func isBasicType(name string) bool {
	switch name {
	case "bool", "string",
		"int", "int8", "int16", "int32", "int64",
		"uint", "uint8", "uint16", "uint32", "uint64", "uintptr",
		"float32", "float64",
		"complex64", "complex128",
		"byte", "rune", "any":
		return true
	}
	return false
}

func needsDeepCopy(expr ast.Expr) bool {
	switch expr.(type) {
	case *ast.StructType, *ast.ArrayType, *ast.MapType, *ast.StarExpr:
		return true
	}
	return false
}

func exprToString(expr ast.Expr) string {
	switch t := expr.(type) {
	case *ast.Ident:
		return t.Name
	case *ast.StarExpr:
		return "*" + exprToString(t.X)
	case *ast.ArrayType:
		return "[]" + exprToString(t.Elt)
	case *ast.MapType:
		return "map[" + exprToString(t.Key) + "]" + exprToString(t.Value)
	case *ast.SelectorExpr:
		return exprToString(t.X) + "." + t.Sel.Name
	case *ast.InterfaceType:
		if t.Methods == nil || len(t.Methods.List) == 0 {
			return "any"
		}
		return "interface{}"
	}
	return ""
}
