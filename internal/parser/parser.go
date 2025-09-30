package parser

import (
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/i-icc/xsd2proto/internal/model"
)

type Parser struct{}

func New() *Parser {
	return &Parser{}
}

func (p *Parser) ParseFile(filePath string) (*model.Schema, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open XSD file: %w", err)
	}
	defer file.Close()

	return p.Parse(file)
}

func (p *Parser) Parse(reader io.Reader) (*model.Schema, error) {
	var schema model.Schema

	decoder := xml.NewDecoder(reader)
	if err := decoder.Decode(&schema); err != nil {
		return nil, fmt.Errorf("failed to parse XSD: %w", err)
	}

	return &schema, nil
}

func (p *Parser) Validate(schema *model.Schema) error {
	if schema == nil {
		return fmt.Errorf("schema is nil")
	}

	elementNames := make(map[string]int)
	for _, element := range schema.Elements {
		if element.Name == "" {
			return fmt.Errorf("element with empty name found")
		}
		elementNames[element.Name]++
	}

	typeNames := make(map[string]int)
	for _, complexType := range schema.ComplexTypes {
		if complexType.Name == "" {
			return fmt.Errorf("complexType with empty name found")
		}
		typeNames[complexType.Name]++
	}

	for _, simpleType := range schema.SimpleTypes {
		if simpleType.Name == "" {
			return fmt.Errorf("simpleType with empty name found")
		}
		typeNames[simpleType.Name]++
	}

	return nil
}

func (p *Parser) ParseFileWithImports(filePath string) (*model.Schema, error) {
	processedFiles := make(map[string]bool)
	return p.parseFileRecursive(filePath, processedFiles)
}

func (p *Parser) parseFileRecursive(filePath string, processedFiles map[string]bool) (*model.Schema, error) {
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to get absolute path for %s: %w", filePath, err)
	}

	if processedFiles[absPath] {
		return nil, nil
	}
	processedFiles[absPath] = true

	schema, err := p.ParseFile(filePath)
	if err != nil {
		return nil, err
	}

	baseDir := filepath.Dir(filePath)
	for _, imp := range schema.Imports {
		var importPath string

		if imp.SchemaLocation != "" {
			importPath = filepath.Join(baseDir, imp.SchemaLocation)
		} else if imp.Namespace != "" {
			derivedPath := p.deriveFilePathFromNamespace(imp.Namespace, baseDir)
			if derivedPath != "" {
				importPath = derivedPath
			}
		}

		if importPath != "" {
			if _, err := os.Stat(importPath); err == nil {
				importedSchema, err := p.parseFileRecursive(importPath, processedFiles)
				if err != nil {
					return nil, fmt.Errorf("failed to process import %s: %w", importPath, err)
				}
				if importedSchema != nil {
					schema.ImportedSchemas = append(schema.ImportedSchemas, importedSchema)
				}
			}
		}
	}

	for _, inc := range schema.Includes {
		if inc.SchemaLocation != "" {
			includePath := filepath.Join(baseDir, inc.SchemaLocation)
			includedSchema, err := p.parseFileRecursive(includePath, processedFiles)
			if err != nil {
				return nil, fmt.Errorf("failed to process include %s: %w", inc.SchemaLocation, err)
			}
			if includedSchema != nil {
				schema.ImportedSchemas = append(schema.ImportedSchemas, includedSchema)
			}
		}
	}

	return schema, nil
}

func (p *Parser) deriveFilePathFromNamespace(namespace, baseDir string) string {
	if namespace == "" {
		return ""
	}

	if strings.HasPrefix(namespace, "./") {
		path := strings.TrimPrefix(namespace, "./")
		fileName := strings.ReplaceAll(path, "/", ".") + ".xsd"
		return filepath.Join(baseDir, fileName)
	}

	if strings.HasPrefix(namespace, "http://") || strings.HasPrefix(namespace, "https://") {
		path := strings.TrimPrefix(namespace, "http://")
		path = strings.TrimPrefix(path, "https://")
		fileName := strings.ReplaceAll(path, "/", ".") + ".xsd"
		return filepath.Join(baseDir, fileName)
	}

	fileName := strings.ReplaceAll(namespace, "/", ".") + ".xsd"
	return filepath.Join(baseDir, fileName)
}
