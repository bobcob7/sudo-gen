// Package manager implements the manager code generation subtool.
package manager

import (
	"strings"

	"github.com/bobcob7/merge-config/internal/codegen"
)

// PathInfo represents a subscribable path in the config tree.
type PathInfo struct {
	Path              string   // Full dot path: "Database.Host"
	PathConst         string   // For const/method names: "DatabaseHost"
	Segments          []string // ["Database", "Host"]
	FieldName         string   // Final field name: "Host"
	FieldType         string   // Go type: "string"
	IsPointer         bool     // Field itself is a pointer
	IsSlice           bool     // Field is a slice
	IsMap             bool     // Field is a map
	IsLocalStruct     bool     // Field is a local struct type (needs partial conversion)
	ParentIsPtr       bool     // Parent struct is accessed via pointer
	NilCheckPath      string   // Path for nil check: "m.config.Database"
	AccessorExpr      string   // Full accessor: "m.config.Database.Host"
	ZeroValue         string   // Zero value for the type
	ParentTypeName    string   // For nested: "DatabaseConfig" (the struct name)
	NeedsTimeImport   bool     // Whether this field uses time package
	SkipTransactionSet bool    // Skip generating transaction setter (e.g., for struct pointers)
}

// BuildPaths creates all subscribable paths from struct info.
func BuildPaths(info *codegen.StructInfo, nested []*codegen.StructInfo) []PathInfo {
	nestedMap := make(map[string]*codegen.StructInfo, len(nested))
	for _, n := range nested {
		nestedMap[n.Name] = n
	}
	var paths []PathInfo
	buildPathsRecursive(info, nestedMap, "", "m.config", "", false, &paths)
	return paths
}

func buildPathsRecursive(info *codegen.StructInfo, nestedMap map[string]*codegen.StructInfo, prefix, accessorPrefix, parentTypeName string, parentIsPtr bool, paths *[]PathInfo) {
	for _, field := range info.Fields {
		path := field.Name
		if prefix != "" {
			path = prefix + "." + field.Name
		}
		accessor := accessorPrefix + "." + field.Name
		pathConst := strings.ReplaceAll(path, ".", "")
		nilCheckPath := ""
		if parentIsPtr {
			nilCheckPath = accessorPrefix
		}
		isLocalStruct := field.IsStruct && field.TypePkg == "" && !field.IsSlice && !field.IsMap
		pi := PathInfo{
			Path:               path,
			PathConst:          pathConst,
			Segments:           strings.Split(path, "."),
			FieldName:          field.Name,
			FieldType:          fieldTypeString(field),
			IsPointer:          field.IsPointer,
			IsSlice:            field.IsSlice,
			IsMap:              field.IsMap,
			IsLocalStruct:      isLocalStruct,
			ParentIsPtr:        parentIsPtr,
			NilCheckPath:       nilCheckPath,
			AccessorExpr:       accessor,
			ZeroValue:          zeroValueFor(field),
			ParentTypeName:     parentTypeName,
			NeedsTimeImport:    field.TypePkg == "time",
			SkipTransactionSet: field.IsPointer && isLocalStruct,
		}
		*paths = append(*paths, pi)
		// Recurse into nested structs (local package only)
		if field.IsStruct && field.TypePkg == "" && !field.IsSlice && !field.IsMap {
			structName := field.TypeName
			if nestedInfo, ok := nestedMap[structName]; ok {
				buildPathsRecursive(nestedInfo, nestedMap, path, accessor, structName, field.IsPointer || parentIsPtr, paths)
			}
		}
	}
}

func fieldTypeString(f codegen.FieldInfo) string {
	if f.IsPointer {
		if f.TypePkg != "" {
			return "*" + f.TypePkg + "." + f.TypeName
		}
		return "*" + f.TypeName
	}
	if f.IsSlice {
		return f.TypeName
	}
	if f.IsMap {
		return f.TypeName
	}
	if f.TypePkg != "" {
		return f.TypePkg + "." + f.TypeName
	}
	return f.TypeName
}

func zeroValueFor(f codegen.FieldInfo) string {
	if f.IsPointer || f.IsSlice || f.IsMap {
		return "nil"
	}
	if f.IsStruct {
		if f.TypePkg != "" {
			return f.TypePkg + "." + f.TypeName + "{}"
		}
		return f.TypeName + "{}"
	}
	switch f.TypeName {
	case "string":
		return `""`
	case "bool":
		return "false"
	case "int", "int8", "int16", "int32", "int64",
		"uint", "uint8", "uint16", "uint32", "uint64",
		"float32", "float64", "byte", "rune":
		return "0"
	default:
		if f.TypePkg == "time" && f.TypeName == "Time" {
			return "time.Time{}"
		}
		return f.TypeName + "{}"
	}
}
