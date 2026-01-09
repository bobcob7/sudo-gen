// Package merge implements the merge code generation subtool.
package merge

import (
	"fmt"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/bobcob7/sudo-gen/internal/codegen"
)

// Subtool implements the merge code generator.
type Subtool struct{}

// Name returns the subtool name.
func (s *Subtool) Name() string { return "merge" }

// Description returns the subtool description.
func (s *Subtool) Description() string {
	return "Generate partial types and ApplyPartial methods for config merging"
}

// Run executes the merge code generation.
func (s *Subtool) Run(cfg codegen.GeneratorConfig) error {
	info, err := codegen.ParseStruct(cfg.SourceDir, cfg.SourceFile, cfg.TypeName)
	if err != nil {
		return fmt.Errorf("parsing struct: %w", err)
	}
	nested, err := codegen.FindNestedStructs(cfg.SourceDir, cfg.SourceFile, info)
	if err != nil {
		return fmt.Errorf("finding nested structs: %w", err)
	}
	allStructs := append([]*codegen.StructInfo{info}, nested...)

	// Build map of external structs for template functions
	externalStructs := make(map[string]bool)
	for _, st := range allStructs {
		if st.Package != "" {
			externalStructs[st.Package+"."+st.Name] = true
		}
	}

	// Collect imports from all structs (root and nested)
	allImports := collectAllImports(allStructs)
	if err := generatePartialFile(cfg, allStructs, allImports, externalStructs); err != nil {
		return fmt.Errorf("generating partial file: %w", err)
	}
	// For merge file, only include imports for external struct types we generate helpers for
	mergeImports := collectMergeImports(allStructs, externalStructs)
	if err := generateMergeFile(cfg, allStructs, externalStructs, mergeImports); err != nil {
		return fmt.Errorf("generating merge file: %w", err)
	}
	if cfg.GenerateTest {
		if err := generateMergeTestFile(cfg, allStructs, externalStructs); err != nil {
			return fmt.Errorf("generating merge test file: %w", err)
		}
	}
	return nil
}

func generatePartialFile(cfg codegen.GeneratorConfig, structs []*codegen.StructInfo, imports []codegen.ImportInfo, externalStructs map[string]bool) error {
	baseName := strings.TrimSuffix(cfg.SourceFile, ".go")
	outputFile := filepath.Join(cfg.OutputDir, baseName+"_partial.go")
	data := struct {
		Package string
		Imports []codegen.ImportInfo
		Structs []*codegen.StructInfo
	}{
		Package: cfg.OutputPkg,
		Imports: imports,
		Structs: structs,
	}
	gen := codegen.NewTemplateGenerator(templateFuncs(externalStructs))
	return gen.GenerateFile(outputFile, partialTemplate, data)
}

func generateMergeFile(cfg codegen.GeneratorConfig, structs []*codegen.StructInfo, externalStructs map[string]bool, imports []codegen.ImportInfo) error {
	baseName := strings.TrimSuffix(cfg.SourceFile, ".go")
	outputFile := filepath.Join(cfg.OutputDir, baseName+"_merge.go")
	data := struct {
		Package string
		Structs []*codegen.StructInfo
		Imports []codegen.ImportInfo
	}{
		Package: cfg.OutputPkg,
		Structs: structs,
		Imports: imports,
	}
	gen := codegen.NewTemplateGenerator(templateFuncs(externalStructs))
	return gen.GenerateFile(outputFile, mergeTemplate, data)
}

func generateMergeTestFile(cfg codegen.GeneratorConfig, structs []*codegen.StructInfo, externalStructs map[string]bool) error {
	baseName := strings.TrimSuffix(cfg.SourceFile, ".go")
	outputFile := filepath.Join(cfg.OutputDir, baseName+"_merge_test.go")
	data := struct {
		Package string
		Structs []*codegen.StructInfo
	}{
		Package: cfg.OutputPkg,
		Structs: structs,
	}
	gen := codegen.NewTemplateGenerator(templateFuncs(externalStructs))
	return gen.GenerateFile(outputFile, mergeTestTemplate, data)
}

func templateFuncs(externalStructs map[string]bool) template.FuncMap {
	return template.FuncMap{
		"partialType":       partialTypeName,
		"pointerType":       pointerTypeNameFunc(externalStructs),
		"needsConversion":   needsConversionFunc(externalStructs),
		"isExternal":        isExternalFunc(externalStructs),
		"isExternalField":   isExternalFieldFunc(externalStructs),
		"externalPartial":   externalPartialNameFunc(externalStructs),
	}
}

func partialTypeName(s *codegen.StructInfo) string {
	if s.Package != "" {
		// External package struct: prefix with capitalized package name
		return capitalize(s.Package) + s.Name + "Partial"
	}
	return s.Name + "Partial"
}

func capitalize(s string) string {
	if s == "" {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}

func pointerTypeNameFunc(externalStructs map[string]bool) func(f codegen.FieldInfo) string {
	return func(f codegen.FieldInfo) string {
		if f.IsPointer {
			if f.IsStruct && f.TypePkg == "" {
				return "*" + f.TypeName + "Partial"
			}
			// Check if this is an external struct we're generating partials for
			if f.TypePkg != "" && externalStructs[f.TypePkg+"."+f.TypeName] {
				return "*" + capitalize(f.TypePkg) + f.TypeName + "Partial"
			}
			if f.TypePkg != "" {
				return "*" + f.TypePkg + "." + f.TypeName
			}
			return "*" + f.TypeName
		}
		if f.IsSlice || f.IsMap {
			return f.TypeName
		}
		if f.IsStruct && f.TypePkg == "" {
			return "*" + f.TypeName + "Partial"
		}
		// Check if this is an external struct we're generating partials for
		if f.TypePkg != "" && externalStructs[f.TypePkg+"."+f.TypeName] {
			return "*" + capitalize(f.TypePkg) + f.TypeName + "Partial"
		}
		if f.TypePkg != "" {
			return "*" + f.TypePkg + "." + f.TypeName
		}
		return "*" + f.TypeName
	}
}

func needsConversionFunc(externalStructs map[string]bool) func(f codegen.FieldInfo) bool {
	return func(f codegen.FieldInfo) bool {
		if f.IsSlice || f.IsMap {
			return false
		}
		// Local struct
		if f.IsStruct && f.TypePkg == "" {
			return true
		}
		// External struct we're generating partials for
		if f.TypePkg != "" && externalStructs[f.TypePkg+"."+f.TypeName] {
			return true
		}
		return false
	}
}

func isExternalFunc(externalStructs map[string]bool) func(s *codegen.StructInfo) bool {
	return func(s *codegen.StructInfo) bool {
		return s.Package != ""
	}
}

func isExternalFieldFunc(externalStructs map[string]bool) func(f codegen.FieldInfo) bool {
	return func(f codegen.FieldInfo) bool {
		if f.TypePkg == "" {
			return false
		}
		return externalStructs[f.TypePkg+"."+f.TypeName]
	}
}

func externalPartialNameFunc(externalStructs map[string]bool) func(f codegen.FieldInfo) string {
	return func(f codegen.FieldInfo) string {
		if f.TypePkg != "" && externalStructs[f.TypePkg+"."+f.TypeName] {
			return capitalize(f.TypePkg) + f.TypeName + "Partial"
		}
		return f.TypeName + "Partial"
	}
}

// collectMergeImports gathers imports needed for the merge file (only external struct packages).
func collectMergeImports(structs []*codegen.StructInfo, externalStructs map[string]bool) []codegen.ImportInfo {
	// Build a map of all available imports
	allImports := make(map[string]codegen.ImportInfo)
	for _, s := range structs {
		for _, imp := range s.Imports {
			pkgName := imp.Alias
			if pkgName == "" {
				pkgName = filepath.Base(imp.Path)
			}
			allImports[pkgName] = imp
		}
	}

	// For merge file, we only need imports for external structs we're generating Apply helpers for
	usedPkgs := make(map[string]bool)
	for _, s := range structs {
		for _, f := range s.Fields {
			if f.TypePkg != "" && externalStructs[f.TypePkg+"."+f.TypeName] {
				usedPkgs[f.TypePkg] = true
			}
		}
	}

	imports := make([]codegen.ImportInfo, 0)
	for pkgName := range usedPkgs {
		if imp, ok := allImports[pkgName]; ok {
			imports = append(imports, imp)
		}
	}
	return imports
}

// collectAllImports gathers imports from all structs that are actually used by fields.
func collectAllImports(structs []*codegen.StructInfo) []codegen.ImportInfo {
	// Build a map of all available imports
	allImports := make(map[string]codegen.ImportInfo)
	for _, s := range structs {
		for _, imp := range s.Imports {
			pkgName := imp.Alias
			if pkgName == "" {
				pkgName = filepath.Base(imp.Path)
			}
			allImports[pkgName] = imp
		}
	}

	// Find which packages are actually used by fields
	usedPkgs := make(map[string]bool)
	for _, s := range structs {
		for _, f := range s.Fields {
			if f.TypePkg != "" {
				usedPkgs[f.TypePkg] = true
			}
		}
	}

	// Only include imports that are used
	imports := make([]codegen.ImportInfo, 0)
	for pkgName := range usedPkgs {
		if imp, ok := allImports[pkgName]; ok {
			imports = append(imports, imp)
		}
	}
	return imports
}
