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
	if err := generatePartialFile(cfg, allStructs, info.Imports); err != nil {
		return fmt.Errorf("generating partial file: %w", err)
	}
	if err := generateMergeFile(cfg, allStructs); err != nil {
		return fmt.Errorf("generating merge file: %w", err)
	}
	if cfg.GenerateTest {
		if err := generateMergeTestFile(cfg, allStructs); err != nil {
			return fmt.Errorf("generating merge test file: %w", err)
		}
	}
	return nil
}

func generatePartialFile(cfg codegen.GeneratorConfig, structs []*codegen.StructInfo, imports []codegen.ImportInfo) error {
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
	gen := codegen.NewTemplateGenerator(templateFuncs())
	return gen.GenerateFile(outputFile, partialTemplate, data)
}

func generateMergeFile(cfg codegen.GeneratorConfig, structs []*codegen.StructInfo) error {
	baseName := strings.TrimSuffix(cfg.SourceFile, ".go")
	outputFile := filepath.Join(cfg.OutputDir, baseName+"_merge.go")
	data := struct {
		Package string
		Structs []*codegen.StructInfo
	}{
		Package: cfg.OutputPkg,
		Structs: structs,
	}
	gen := codegen.NewTemplateGenerator(templateFuncs())
	return gen.GenerateFile(outputFile, mergeTemplate, data)
}

func generateMergeTestFile(cfg codegen.GeneratorConfig, structs []*codegen.StructInfo) error {
	baseName := strings.TrimSuffix(cfg.SourceFile, ".go")
	outputFile := filepath.Join(cfg.OutputDir, baseName+"_merge_test.go")
	data := struct {
		Package string
		Structs []*codegen.StructInfo
	}{
		Package: cfg.OutputPkg,
		Structs: structs,
	}
	gen := codegen.NewTemplateGenerator(templateFuncs())
	return gen.GenerateFile(outputFile, mergeTestTemplate, data)
}

func templateFuncs() template.FuncMap {
	return template.FuncMap{
		"partialType":     partialTypeName,
		"pointerType":     pointerTypeName,
		"needsConversion": needsConversion,
	}
}

func partialTypeName(name string) string {
	return name + "Partial"
}

func pointerTypeName(f codegen.FieldInfo) string {
	if f.IsPointer {
		if f.IsStruct && f.TypePkg == "" {
			return "*" + f.TypeName + "Partial"
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
	if f.TypePkg != "" {
		return "*" + f.TypePkg + "." + f.TypeName
	}
	return "*" + f.TypeName
}

func needsConversion(f codegen.FieldInfo) bool {
	return f.IsStruct && f.TypePkg == "" && !f.IsSlice && !f.IsMap
}
