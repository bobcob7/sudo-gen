package layerbroker

import (
	"fmt"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/bobcob7/sudo-gen/internal/codegen"
	"github.com/bobcob7/sudo-gen/internal/codegen/copy"
	"github.com/bobcob7/sudo-gen/internal/codegen/equals"
	"github.com/bobcob7/sudo-gen/internal/codegen/merge"
)

// Subtool implements the layerbroker code generator.
type Subtool struct{}

// Name returns the subtool name.
func (s *Subtool) Name() string { return "layerbroker" }

// Description returns the subtool description.
func (s *Subtool) Description() string {
	return "Generate thread-safe LayerBroker with ordered layers and subscriptions (no reflection)"
}

// Run executes the layerbroker code generation.
// It automatically generates the required dependencies (merge, copy, and equals).
func (s *Subtool) Run(cfg codegen.GeneratorConfig) error {
	// Generate dependencies first
	mergeTool := &merge.Subtool{}
	if err := mergeTool.Run(cfg); err != nil {
		return fmt.Errorf("generating merge dependency: %w", err)
	}
	copyTool := &copy.Subtool{MethodName: "Copy"}
	if err := copyTool.Run(cfg); err != nil {
		return fmt.Errorf("generating copy dependency: %w", err)
	}
	equalsTool := &equals.Subtool{MethodName: "Equal"}
	if err := equalsTool.Run(cfg); err != nil {
		return fmt.Errorf("generating equals dependency: %w", err)
	}
	info, err := codegen.ParseStruct(cfg.SourceDir, cfg.SourceFile, cfg.TypeName)
	if err != nil {
		return fmt.Errorf("parsing struct: %w", err)
	}
	if err := generateLayerBrokerFile(cfg, info); err != nil {
		return err
	}
	if cfg.GenerateTest {
		return generateLayerBrokerTestFile(cfg, info)
	}
	return nil
}

func generateLayerBrokerFile(cfg codegen.GeneratorConfig, info *codegen.StructInfo) error {
	baseName := strings.TrimSuffix(cfg.SourceFile, ".go")
	outputFile := filepath.Join(cfg.OutputDir, baseName+"_layerbroker.go")
	needsTime := false
	// Collect external package imports (excluding "time" which is handled separately)
	externalImports := collectExternalImports(info)
	for _, f := range info.Fields {
		if f.TypePkg == "time" {
			needsTime = true
		}
	}
	data := templateData{
		Package:            cfg.OutputPkg,
		TypeName:           info.Name,
		Fields:             info.Fields,
		NeedsTimeImport:    needsTime,
		NeedsReflectImport: false, // No longer using reflect.DeepEqual
		GenerateJSON:       cfg.GenerateJSON,
		ExternalImports:    externalImports,
	}
	gen := codegen.NewTemplateGenerator(templateFuncs())
	return gen.GenerateFile(outputFile, layerBrokerTemplate, data)
}

// collectExternalImports gathers imports for external packages used by fields.
func collectExternalImports(info *codegen.StructInfo) []codegen.ImportInfo {
	// Build a map of package name to import info
	importMap := make(map[string]codegen.ImportInfo)
	for _, imp := range info.Imports {
		pkgName := imp.Alias
		if pkgName == "" {
			pkgName = filepath.Base(imp.Path)
		}
		importMap[pkgName] = imp
	}

	// Find which external packages are used by fields (excluding "time")
	usedPkgs := make(map[string]bool)
	for _, f := range info.Fields {
		if f.TypePkg != "" && f.TypePkg != "time" {
			usedPkgs[f.TypePkg] = true
		}
	}

	var imports []codegen.ImportInfo
	for pkgName := range usedPkgs {
		if imp, ok := importMap[pkgName]; ok {
			imports = append(imports, imp)
		}
	}
	return imports
}

type templateData struct {
	Package            string
	TypeName           string
	Fields             []codegen.FieldInfo
	NeedsTimeImport    bool
	NeedsReflectImport bool
	GenerateJSON       bool
	ExternalImports    []codegen.ImportInfo
}

func templateFuncs() template.FuncMap {
	return template.FuncMap{
		"lower":         strings.ToLower,
		"partialType":   func(name string) string { return name + "Partial" },
		"isLocalStruct": isLocalStruct,
		"isExported":    isExported,
		"brokerType":    brokerTypeName,
		"layerType":     layerTypeName,
		"newBroker":     newBrokerName,
	}
}

func isExported(name string) bool {
	if len(name) == 0 {
		return false
	}
	return name[0] >= 'A' && name[0] <= 'Z'
}

func brokerTypeName(typeName string) string {
	if isExported(typeName) {
		return typeName + "LayerBroker"
	}
	return typeName + "LayerBroker"
}

func layerTypeName(typeName string) string {
	if isExported(typeName) {
		return typeName + "Layer"
	}
	return typeName + "Layer"
}

func newBrokerName(typeName string) string {
	if isExported(typeName) {
		return "New" + typeName + "LayerBroker"
	}
	// For unexported types, capitalize the first letter after "new"
	return "new" + strings.ToUpper(typeName[:1]) + typeName[1:] + "LayerBroker"
}

func isLocalStruct(f codegen.FieldInfo) bool {
	return f.IsStruct && f.TypePkg == "" && !f.IsSlice && !f.IsMap
}

func generateLayerBrokerTestFile(cfg codegen.GeneratorConfig, info *codegen.StructInfo) error {
	baseName := strings.TrimSuffix(cfg.SourceFile, ".go")
	outputFile := filepath.Join(cfg.OutputDir, baseName+"_layerbroker_test.go")

	// Find first string and int fields for test examples
	var stringField, intField string
	for _, f := range info.Fields {
		if stringField == "" && f.TypeName == "string" && !f.IsPointer && !f.IsSlice && !f.IsMap {
			stringField = f.Name
		}
		if intField == "" && (f.TypeName == "int" || f.TypeName == "int32" || f.TypeName == "int64") && !f.IsPointer && !f.IsSlice && !f.IsMap {
			intField = f.Name
		}
	}

	// Check if time.Time field exists
	needsTime := false
	for _, f := range info.Fields {
		if f.TypePkg == "time" {
			needsTime = true
			break
		}
	}

	data := testTemplateData{
		Package:      cfg.OutputPkg,
		TypeName:     info.Name,
		StringField:  stringField,
		IntField:     intField,
		Fields:       info.Fields,
		GenerateJSON: cfg.GenerateJSON,
		NeedsTime:    needsTime,
	}
	gen := codegen.NewTemplateGenerator(templateFuncs())
	return gen.GenerateFile(outputFile, layerBrokerTestTemplate, data)
}

type testTemplateData struct {
	Package      string
	TypeName     string
	StringField  string
	IntField     string
	Fields       []codegen.FieldInfo
	GenerateJSON bool
	NeedsTime    bool
}
