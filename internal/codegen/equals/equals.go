// Package equals implements the equals code generation subtool.
package equals

import (
	"fmt"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/bobcob7/sudo-gen/internal/codegen"
)

// Subtool implements the equals code generator.
type Subtool struct {
	MethodName string
}

// Name returns the subtool name.
func (s *Subtool) Name() string { return "equals" }

// Description returns the subtool description.
func (s *Subtool) Description() string {
	return "Generate type-safe equality comparison methods for structs"
}

// Run executes the equals code generation.
func (s *Subtool) Run(cfg codegen.GeneratorConfig) error {
	methodName := s.MethodName
	if methodName == "" {
		methodName = "Equal"
	}
	info, err := codegen.ParseStruct(cfg.SourceDir, cfg.SourceFile, cfg.TypeName)
	if err != nil {
		return fmt.Errorf("parsing struct: %w", err)
	}
	nested, err := codegen.FindNestedStructs(cfg.SourceDir, cfg.SourceFile, info)
	if err != nil {
		return fmt.Errorf("finding nested structs: %w", err)
	}
	// Filter out external package structs - we can't add methods to them
	allStructs := []*codegen.StructInfo{info}
	for _, st := range nested {
		if st.Package == "" {
			allStructs = append(allStructs, st)
		}
	}
	return generateEqualsFile(cfg, allStructs, methodName)
}

func generateEqualsFile(cfg codegen.GeneratorConfig, structs []*codegen.StructInfo, methodName string) error {
	baseName := strings.TrimSuffix(cfg.SourceFile, ".go")
	outputFile := filepath.Join(cfg.OutputDir, baseName+"_equals.go")
	data := templateData{
		Package:    cfg.OutputPkg,
		Structs:    structs,
		MethodName: methodName,
	}
	gen := codegen.NewTemplateGenerator(templateFuncs())
	if err := gen.GenerateFile(outputFile, equalsTemplate, data); err != nil {
		return err
	}
	if cfg.GenerateTest {
		testFile := filepath.Join(cfg.OutputDir, baseName+"_equals_test.go")
		return gen.GenerateFile(testFile, equalsTestTemplate, data)
	}
	return nil
}

type templateData struct {
	Package    string
	Structs    []*codegen.StructInfo
	MethodName string
}

func templateFuncs() template.FuncMap {
	return template.FuncMap{
		"isLocalStruct": isLocalStruct,
	}
}

func isLocalStruct(f codegen.FieldInfo) bool {
	return f.IsStruct && f.TypePkg == "" && !f.IsSlice && !f.IsMap
}
