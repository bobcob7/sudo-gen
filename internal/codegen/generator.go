package codegen

import (
	"bytes"
	"fmt"
	"go/format"
	"os"
	"text/template"
)

// TemplateGenerator handles template-based code generation.
type TemplateGenerator struct {
	FuncMap template.FuncMap
}

// NewTemplateGenerator creates a new TemplateGenerator with optional custom functions.
func NewTemplateGenerator(customFuncs template.FuncMap) *TemplateGenerator {
	return &TemplateGenerator{FuncMap: customFuncs}
}

// GenerateFile executes a template and writes the formatted output to a file.
func (g *TemplateGenerator) GenerateFile(outputFile, tmplText string, data any) error {
	tmpl, err := template.New("gen").Funcs(g.FuncMap).Parse(tmplText)
	if err != nil {
		return fmt.Errorf("parsing template: %w", err)
	}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return fmt.Errorf("executing template: %w", err)
	}
	formatted, err := format.Source(buf.Bytes())
	if err != nil {
		_ = os.WriteFile(outputFile+".unformatted", buf.Bytes(), 0644)
		return fmt.Errorf("formatting generated code: %w (wrote unformatted to %s.unformatted)", err, outputFile)
	}
	if err := os.WriteFile(outputFile, formatted, 0644); err != nil {
		return fmt.Errorf("writing file: %w", err)
	}
	fmt.Printf("Generated: %s\n", outputFile)
	return nil
}

// Subtool defines the interface for code generation subtools.
type Subtool interface {
	Name() string
	Description() string
	Run(cfg GeneratorConfig) error
}
