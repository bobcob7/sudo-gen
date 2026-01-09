// Package codegen provides shared types and utilities for code generation tools.
package codegen

import "go/ast"

// StructInfo holds information about a parsed struct type.
type StructInfo struct {
	Name    string
	Fields  []FieldInfo
	Imports []ImportInfo
}

// FieldInfo holds information about a struct field.
type FieldInfo struct {
	Name           string
	Type           string     // Full type string (e.g., "[]string", "map[string]any")
	TypeExpr       ast.Expr   // Original AST expression
	TypeName       string     // Base type name (e.g., "string", "Tag")
	TypePkg        string     // Package prefix if any (e.g., "time" for time.Time)
	IsPointer      bool       // Field is a pointer type
	IsSlice        bool       // Field is a slice
	IsMap          bool       // Field is a map
	IsStruct       bool       // Field is a named struct type (not basic)
	MapKeyType     string     // Key type for maps
	MapValType     string     // Value type for maps
	SliceType      string     // Element type for slices
	Tag            string     // Struct tag
	NeedsDeep      bool       // Requires deep copy (for copy generator)
	StructTypeName string     // Name of struct type for calling methods
	SliceElemIsPtr bool       // Slice element is pointer to struct
}

// ImportInfo holds information about an import.
type ImportInfo struct {
	Path  string
	Alias string
}

// GeneratorConfig holds common configuration for generators.
type GeneratorConfig struct {
	TypeName   string
	SourceFile string
	SourceDir  string
	SourcePkg  string
	OutputDir  string
	OutputPkg  string
}
