package codegen

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"go/types"
	"os"
	"path/filepath"
	"strings"
)

// ParseStruct parses a Go source file and extracts struct information.
func ParseStruct(dir, filename, typeName string) (*StructInfo, error) {
	fset := token.NewFileSet()
	fullPath := filepath.Join(dir, filename)
	f, err := parser.ParseFile(fset, fullPath, nil, parser.ParseComments)
	if err != nil {
		return nil, fmt.Errorf("parsing file: %w", err)
	}
	imports := collectImports(f)
	targetStruct, targetName, err := findStructType(f, typeName)
	if err != nil {
		return nil, err
	}
	fields := parseStructFields(targetStruct, imports)
	return &StructInfo{
		Name:    targetName,
		Fields:  fields,
		Imports: imports,
	}, nil
}

func collectImports(f *ast.File) []ImportInfo {
	imports := make([]ImportInfo, 0, len(f.Imports))
	for _, imp := range f.Imports {
		path := strings.Trim(imp.Path.Value, `"`)
		alias := ""
		if imp.Name != nil {
			alias = imp.Name.Name
		}
		imports = append(imports, ImportInfo{Path: path, Alias: alias})
	}
	return imports
}

func findStructType(f *ast.File, typeName string) (*ast.StructType, string, error) {
	for _, decl := range f.Decls {
		genDecl, ok := decl.(*ast.GenDecl)
		if !ok || genDecl.Tok != token.TYPE {
			continue
		}
		for _, spec := range genDecl.Specs {
			typeSpec, ok := spec.(*ast.TypeSpec)
			if !ok || typeSpec.Name.Name != typeName {
				continue
			}
			structType, ok := typeSpec.Type.(*ast.StructType)
			if !ok {
				return nil, "", fmt.Errorf("type %s is not a struct", typeName)
			}
			return structType, typeSpec.Name.Name, nil
		}
	}
	return nil, "", fmt.Errorf("type %s not found", typeName)
}

func parseStructFields(st *ast.StructType, imports []ImportInfo) []FieldInfo {
	fields := make([]FieldInfo, 0, len(st.Fields.List))
	for _, field := range st.Fields.List {
		if len(field.Names) == 0 {
			continue // Skip embedded fields
		}
		for _, name := range field.Names {
			if !ast.IsExported(name.Name) {
				continue
			}
			fi := parseFieldType(field.Type, imports)
			fi.Name = name.Name
			fi.TypeExpr = field.Type
			fi.Type = exprToString(field.Type)
			if field.Tag != nil {
				fi.Tag = field.Tag.Value
			}
			fields = append(fields, fi)
		}
	}
	return fields
}

func parseFieldType(expr ast.Expr, imports []ImportInfo) FieldInfo {
	fi := FieldInfo{}
	switch t := expr.(type) {
	case *ast.Ident:
		fi.TypeName = t.Name
		fi.IsStruct = !isBasicType(t.Name)
		if fi.IsStruct {
			fi.StructTypeName = t.Name
		}
	case *ast.SelectorExpr:
		if pkg, ok := t.X.(*ast.Ident); ok {
			fi.TypePkg = pkg.Name
			fi.TypeName = t.Sel.Name
			fi.IsStruct = true
		}
	case *ast.StarExpr:
		fi = parseFieldType(t.X, imports)
		fi.IsPointer = true
		fi.NeedsDeep = fi.IsStruct || fi.IsSlice || fi.IsMap
	case *ast.ArrayType:
		fi.IsSlice = true
		elemInfo := parseFieldType(t.Elt, imports)
		if elemInfo.TypePkg != "" {
			fi.SliceType = elemInfo.TypePkg + "." + elemInfo.TypeName
		} else {
			fi.SliceType = elemInfo.TypeName
		}
		fi.TypeName = "[]" + fi.SliceType
		if !isBasicType(elemInfo.TypeName) && elemInfo.TypePkg == "" {
			fi.StructTypeName = elemInfo.TypeName
			fi.NeedsDeep = true
		}
		if elemInfo.IsPointer && elemInfo.IsStruct {
			fi.SliceElemIsPtr = true
			fi.NeedsDeep = true
		}
	case *ast.MapType:
		fi.IsMap = true
		keyInfo := parseFieldType(t.Key, imports)
		valInfo := parseFieldType(t.Value, imports)
		if keyInfo.TypePkg != "" {
			fi.MapKeyType = keyInfo.TypePkg + "." + keyInfo.TypeName
		} else {
			fi.MapKeyType = keyInfo.TypeName
		}
		if valInfo.TypePkg != "" {
			fi.MapValType = valInfo.TypePkg + "." + valInfo.TypeName
		} else {
			fi.MapValType = valInfo.TypeName
		}
		fi.TypeName = fmt.Sprintf("map[%s]%s", fi.MapKeyType, fi.MapValType)
		if fi.MapValType == "any" || fi.MapValType == "interface{}" {
			fi.NeedsDeep = true
		} else if !isBasicType(valInfo.TypeName) && valInfo.TypePkg == "" {
			fi.StructTypeName = valInfo.TypeName
			fi.NeedsDeep = true
		}
	case *ast.InterfaceType:
		fi.TypeName = "any"
	}
	return fi
}

func isBasicType(name string) bool {
	switch name {
	case "bool", "string",
		"int", "int8", "int16", "int32", "int64",
		"uint", "uint8", "uint16", "uint32", "uint64", "uintptr",
		"byte", "rune", "any", "error",
		"float32", "float64",
		"complex64", "complex128":
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
	default:
		return types.ExprString(expr)
	}
}

// FindTypeAfterGenerateDirective finds the struct type declared immediately after a go:generate directive.
func FindTypeAfterGenerateDirective(dir, filename, generatorName string) (string, error) {
	fset := token.NewFileSet()
	fullPath := filepath.Join(dir, filename)
	f, err := parser.ParseFile(fset, fullPath, nil, parser.ParseComments)
	if err != nil {
		return "", fmt.Errorf("parsing file: %w", err)
	}
	for _, decl := range f.Decls {
		genDecl, ok := decl.(*ast.GenDecl)
		if !ok || genDecl.Tok != token.TYPE || genDecl.Doc == nil {
			continue
		}
		for _, comment := range genDecl.Doc.List {
			if strings.Contains(comment.Text, "go:generate") && strings.Contains(comment.Text, generatorName) {
				for _, spec := range genDecl.Specs {
					typeSpec, ok := spec.(*ast.TypeSpec)
					if !ok {
						continue
					}
					if _, ok := typeSpec.Type.(*ast.StructType); ok {
						return typeSpec.Name.Name, nil
					}
				}
			}
		}
	}
	return "", fmt.Errorf("no struct type found after go:generate %s directive", generatorName)
}

// FindTypeAfterLine finds the struct type declared immediately after the given line number.
func FindTypeAfterLine(filename string, lineNum int) (string, error) {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, filename, nil, parser.ParseComments)
	if err != nil {
		return "", fmt.Errorf("parsing file: %w", err)
	}
	for _, decl := range f.Decls {
		genDecl, ok := decl.(*ast.GenDecl)
		if !ok || genDecl.Tok != token.TYPE {
			continue
		}
		for _, spec := range genDecl.Specs {
			typeSpec, ok := spec.(*ast.TypeSpec)
			if !ok {
				continue
			}
			pos := fset.Position(typeSpec.Pos())
			if pos.Line > lineNum {
				if _, ok := typeSpec.Type.(*ast.StructType); ok {
					return typeSpec.Name.Name, nil
				}
			}
		}
	}
	return "", fmt.Errorf("no struct type found after line %d", lineNum)
}

// FindNestedStructs finds all struct types referenced by the given struct.
// It searches all .go files in the directory to find nested types.
func FindNestedStructs(dir, filename string, info *StructInfo) ([]*StructInfo, error) {
	var nested []*StructInfo
	seen := make(map[string]bool, len(info.Fields))
	seen[info.Name] = true
	for _, field := range info.Fields {
		typeName := field.StructTypeName
		if typeName == "" || field.TypePkg != "" || seen[typeName] {
			continue
		}
		nestedInfo, err := FindStructInPackage(dir, typeName)
		if err != nil {
			continue // Type might be external or not found
		}
		seen[typeName] = true
		nested = append(nested, nestedInfo)
		subNested, err := FindNestedStructs(dir, "", nestedInfo)
		if err == nil {
			for _, sub := range subNested {
				if !seen[sub.Name] {
					seen[sub.Name] = true
					nested = append(nested, sub)
				}
			}
		}
	}
	return nested, nil
}

// FindStructInPackage searches all .go files in the directory for a struct type.
func FindStructInPackage(dir, typeName string) (*StructInfo, error) {
	fset := token.NewFileSet()
	pkgs, err := parser.ParseDir(fset, dir, func(fi os.FileInfo) bool {
		return !strings.HasSuffix(fi.Name(), "_test.go")
	}, parser.ParseComments)
	if err != nil {
		return nil, fmt.Errorf("parsing directory: %w", err)
	}
	for _, pkg := range pkgs {
		for filename, f := range pkg.Files {
			imports := collectImports(f)
			for _, decl := range f.Decls {
				genDecl, ok := decl.(*ast.GenDecl)
				if !ok || genDecl.Tok != token.TYPE {
					continue
				}
				for _, spec := range genDecl.Specs {
					typeSpec, ok := spec.(*ast.TypeSpec)
					if !ok || typeSpec.Name.Name != typeName {
						continue
					}
					structType, ok := typeSpec.Type.(*ast.StructType)
					if !ok {
						continue
					}
					fields := parseStructFields(structType, imports)
					return &StructInfo{
						Name:    typeSpec.Name.Name,
						Fields:  fields,
						Imports: imports,
						// Store which file the struct was found in
						SourceFile: filepath.Base(filename),
					}, nil
				}
			}
		}
	}
	return nil, fmt.Errorf("type %s not found in package", typeName)
}

// CollectRequiredImports determines which imports are needed for generated code.
func CollectRequiredImports(fields []FieldInfo, fileImports []ImportInfo) []ImportInfo {
	needed := make(map[string]string, len(fileImports))
	importMap := make(map[string]string, len(fileImports))
	for _, imp := range fileImports {
		importMap[imp.Path] = imp.Alias
	}
	for _, f := range fields {
		collectImportsFromExpr(f.TypeExpr, importMap, needed)
	}
	imports := make([]ImportInfo, 0, len(needed))
	for path, alias := range needed {
		imports = append(imports, ImportInfo{Path: path, Alias: alias})
	}
	return imports
}

func collectImportsFromExpr(expr ast.Expr, importMap, needed map[string]string) {
	switch t := expr.(type) {
	case *ast.SelectorExpr:
		pkg, ok := t.X.(*ast.Ident)
		if !ok {
			return
		}
		for path, alias := range importMap {
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
		collectImportsFromExpr(t.X, importMap, needed)
	case *ast.ArrayType:
		collectImportsFromExpr(t.Elt, importMap, needed)
	case *ast.MapType:
		collectImportsFromExpr(t.Key, importMap, needed)
		collectImportsFromExpr(t.Value, importMap, needed)
	}
}
