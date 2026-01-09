// sudo-gen is a unified code generation tool for Go structs.
//
// Usage:
//
//	//go:generate sudo-gen merge
//	type Config struct { ... }
//
//	//go:generate sudo-gen copy
//	type Config struct { ... }
//
// Or with explicit type:
//
//	//go:generate sudo-gen merge -type=Config
//	//go:generate sudo-gen copy -type=Config
//
// Subcommands:
//
//	merge    Generate partial types and ApplyPartial methods for config merging
//	copy     Generate deep copy methods for structs
//
// Flags:
//
//	-type     The name of the struct type (inferred if directive is above the type)
//	-output   Output directory for generated files (default: same as source)
//	-package  Package name for generated files (default: same as source)
//	-method   For copy: name of the generated method (default: Copy)
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"merge-config/internal/codegen"
	"merge-config/internal/codegen/copy"
	"merge-config/internal/codegen/merge"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}
	subcommand := os.Args[1]
	if subcommand == "-h" || subcommand == "-help" || subcommand == "--help" {
		printUsage()
		os.Exit(0)
	}
	os.Args = append(os.Args[:1], os.Args[2:]...)
	var (
		typeName   string
		outputDir  string
		pkgName    string
		methodName string
	)
	flag.StringVar(&typeName, "type", "", "Name of the struct type (inferred if directive is above the type)")
	flag.StringVar(&outputDir, "output", "", "Output directory for generated files (default: same as source)")
	flag.StringVar(&pkgName, "package", "", "Package name for generated files (default: same as source)")
	flag.StringVar(&methodName, "method", "Copy", "For copy: name of the generated copy method")
	flag.Parse()
	sourceFile := os.Getenv("GOFILE")
	if sourceFile == "" {
		fmt.Fprintln(os.Stderr, "error: GOFILE environment variable not set (are you running via go generate?)")
		os.Exit(1)
	}
	sourceDir, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error getting working directory: %v\n", err)
		os.Exit(1)
	}
	if typeName == "" {
		typeName, err = detectTypeName(subcommand, sourceDir, sourceFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			fmt.Fprintln(os.Stderr, "hint: use -type=TypeName or place the directive directly above the struct")
			os.Exit(1)
		}
	}
	if outputDir == "" {
		outputDir = sourceDir
	}
	sourcePkg := os.Getenv("GOPACKAGE")
	if pkgName == "" {
		pkgName = sourcePkg
	}
	cfg := codegen.GeneratorConfig{
		TypeName:   typeName,
		SourceFile: sourceFile,
		SourceDir:  sourceDir,
		SourcePkg:  sourcePkg,
		OutputDir:  outputDir,
		OutputPkg:  pkgName,
	}
	if err := runSubcommand(subcommand, cfg, methodName); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func detectTypeName(subcommand, sourceDir, sourceFile string) (string, error) {
	generatorName := "sudo-gen " + subcommand
	typeName, err := codegen.FindTypeAfterGenerateDirective(sourceDir, sourceFile, generatorName)
	if err == nil {
		return typeName, nil
	}
	goLine := os.Getenv("GOLINE")
	if goLine != "" {
		lineNum, lineErr := strconv.Atoi(goLine)
		if lineErr == nil {
			return codegen.FindTypeAfterLine(filepath.Join(sourceDir, sourceFile), lineNum)
		}
	}
	return "", err
}

func runSubcommand(name string, cfg codegen.GeneratorConfig, methodName string) error {
	switch name {
	case "merge":
		subtool := &merge.Subtool{}
		return subtool.Run(cfg)
	case "copy":
		subtool := &copy.Subtool{MethodName: methodName}
		return subtool.Run(cfg)
	default:
		return fmt.Errorf("unknown subcommand: %s", name)
	}
}

func printUsage() {
	fmt.Fprintf(os.Stderr, `sudo-gen - Unified code generation tool for Go structs

Usage:
  //go:generate sudo-gen <subcommand> [flags]
  type Config struct { ... }

Subcommands:
  merge    Generate partial types and ApplyPartial methods for config merging
  copy     Generate deep copy methods for structs

Examples:
  //go:generate sudo-gen merge
  //go:generate sudo-gen copy
  //go:generate sudo-gen merge -type=Config
  //go:generate sudo-gen copy -method=Clone

Flags:
  -type string
        Name of the struct type (inferred if directive is above the type)
  -output string
        Output directory for generated files (default: same as source)
  -package string
        Package name for generated files (default: same as source)
  -method string
        For copy: name of the generated copy method (default: Copy)
  -help
        Show this help message

Generated Files:
  merge:
    {source}_partial.go  - Partial version of the type with pointer fields
    {source}_merge.go    - ApplyPartial method for merging partials
  copy:
    {type}_copy.go       - Deep copy method for the struct

`)
}
