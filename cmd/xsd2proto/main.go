package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/i-icc/xsd2proto"
	"github.com/i-icc/xsd2proto/internal/converter"
	"github.com/i-icc/xsd2proto/internal/generator"
	"github.com/i-icc/xsd2proto/internal/parser"
)

const usageText = `xsd2proto - Convert XSD files to Protocol Buffer definitions

Usage:
  xsd2proto [options] <input.xsd>

Options:
  -o, --output string     Output file path (default: input filename with .proto extension)
  -p, --package string    Go package option for generated proto file
  -v, --verbose           Enable verbose output
  -h, --help             Show this help message
      --version          Show version information
      --no-header        Disable auto-generation header comment

Examples:
  xsd2proto schema.xsd                          # Convert schema.xsd to schema.proto
  xsd2proto -o output.proto schema.xsd         # Convert with custom output path
  xsd2proto -p "example.com/proto" schema.xsd  # Convert with go_package option
  xsd2proto -v schema.xsd                       # Convert with verbose output
  xsd2proto --no-header schema.xsd             # Convert without header comment
`

func main() {
	var (
		outputPath = flag.String("o", "", "Output file path")
		goPackage  = flag.String("p", "", "Go package option")
		verbose    = flag.Bool("v", false, "Enable verbose output")
		help       = flag.Bool("h", false, "Show help")
		version    = flag.Bool("version", false, "Show version")
		noHeader   = flag.Bool("no-header", false, "Disable auto-generation header comment")
	)

	// Custom usage function
	flag.Usage = func() {
		fmt.Print(usageText)
	}

	flag.Parse()

	// Handle version flag
	if *version {
		fmt.Printf("xsd2proto version %s\n", xsd2proto.GetVersion())
		os.Exit(0)
	}

	// Handle help flag
	if *help {
		flag.Usage()
		os.Exit(0)
	}

	// Check if input file is provided
	args := flag.Args()
	if len(args) != 1 {
		fmt.Fprintf(os.Stderr, "Error: Please provide exactly one XSD input file\n\n")
		flag.Usage()
		os.Exit(1)
	}

	inputPath := args[0]

	// Check if input file exists
	if _, err := os.Stat(inputPath); os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "Error: Input file '%s' does not exist\n", inputPath)
		os.Exit(1)
	}

	// Perform conversion
	if err := convertXSD(inputPath, *outputPath, *goPackage, *verbose, !*noHeader); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	if !*verbose {
		fmt.Printf("Successfully converted %s\n", inputPath)
	}
}

func convertXSD(inputPath, outputPath, goPackage string, verbose, includeHeader bool) error {
	if verbose {
		fmt.Printf("Converting %s to protobuf...\n", inputPath)
	}

	// Create instances
	p := parser.New()
	conv := converter.New()
	gen := generator.New()

	// Configure generator
	gen.SetHeaderOptions(includeHeader, xsd2proto.GetVersion())

	// Parse XSD file with imports/includes
	schema, err := p.ParseFileWithImports(inputPath)
	if err != nil {
		return fmt.Errorf("failed to parse XSD file: %w", err)
	}

	// Validate parsed schema
	if err := p.Validate(schema); err != nil {
		return fmt.Errorf("schema validation failed: %w", err)
	}

	if verbose {
		fmt.Printf("Successfully parsed XSD schema with %d elements, %d complex types, %d simple types\n",
			len(schema.Elements), len(schema.ComplexTypes), len(schema.SimpleTypes))
	}

	// Convert to protobuf model
	protoFile, err := conv.Convert(schema)
	if err != nil {
		return fmt.Errorf("failed to convert schema: %w", err)
	}

	// Add go_package option if specified
	if goPackage != "" {
		protoFile.Options["go_package"] = goPackage
	}

	// Generate protobuf content
	content, err := gen.Generate(protoFile)
	if err != nil {
		return fmt.Errorf("failed to generate protobuf: %w", err)
	}

	// Determine output path
	finalOutputPath := outputPath
	if finalOutputPath == "" {
		dir := filepath.Dir(inputPath)
		base := filepath.Base(inputPath)
		ext := filepath.Ext(base)
		name := strings.TrimSuffix(base, ext)
		finalOutputPath = filepath.Join(dir, name+".proto")
	}

	// Write to output file
	if err := writeToFile(finalOutputPath, content); err != nil {
		return fmt.Errorf("failed to write output file: %w", err)
	}

	if verbose {
		fmt.Printf("Successfully generated %s\n", finalOutputPath)
	}

	return nil
}

func writeToFile(path, content string) error {
	// Create directory if it doesn't exist
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Write content to file
	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	if _, err := file.WriteString(content); err != nil {
		return fmt.Errorf("failed to write content: %w", err)
	}

	return nil
}
