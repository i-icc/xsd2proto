package converter

import (
	"strings"
)

type TypeMapper struct {
	customMappings map[string]string
}

func NewTypeMapper() *TypeMapper {
	return &TypeMapper{
		customMappings: make(map[string]string),
	}
}

func (tm *TypeMapper) MapXSDType(xsdType string) (string, error) {
	cleanType := tm.CleanTypeName(xsdType)

	if protoType, exists := tm.customMappings[cleanType]; exists {
		return protoType, nil
	}
	switch cleanType {
	case "string", "normalizedString", "token", "NMTOKEN", "Name", "NCName", "ID", "IDREF":
		return "string", nil
	case "boolean":
		return "bool", nil
	case "int", "integer", "short", "byte", "unsignedByte":
		return "int32", nil
	case "long", "unsignedInt":
		return "int64", nil
	case "float":
		return "float", nil
	case "double", "decimal":
		return "double", nil
	case "dateTime", "date", "time":
		return "google.protobuf.Timestamp", nil
	case "duration":
		return "google.protobuf.Duration", nil
	case "anyURI":
		return "string", nil
	case "base64Binary", "hexBinary":
		return "bytes", nil
	case "unsignedLong":
		return "uint64", nil
	case "unsignedShort":
		return "uint32", nil
	default:
		// Return the cleaned type name as-is, without formatting
		// The converter will handle the formatting and uniqueness
		return cleanType, nil
	}
}

func (tm *TypeMapper) AddCustomMapping(xsdType, protoType string) {
	tm.customMappings[xsdType] = protoType
}

// CleanTypeName removes namespace prefix from type name
func (tm *TypeMapper) CleanTypeName(typeName string) string {
	if idx := strings.LastIndex(typeName, ":"); idx != -1 {
		return typeName[idx+1:]
	}
	return typeName
}

func (tm *TypeMapper) formatCustomTypeName(typeName string) string {
	return tm.toPascalCase(typeName)
}

func (tm *TypeMapper) toPascalCase(s string) string {
	parts := strings.FieldsFunc(s, func(r rune) bool {
		return r == '_' || r == '-' || r == '.'
	})

	var finalParts []string
	for _, part := range parts {
		camelParts := tm.splitCamelCase(part)
		finalParts = append(finalParts, camelParts...)
	}

	var result strings.Builder
	for _, part := range finalParts {
		if len(part) > 0 {
			result.WriteString(strings.ToUpper(part[:1]))
			if len(part) > 1 {
				result.WriteString(strings.ToLower(part[1:]))
			}
		}
	}
	return result.String()
}

func (tm *TypeMapper) splitCamelCase(s string) []string {
	if len(s) == 0 {
		return []string{}
	}

	var parts []string
	var current strings.Builder

	for i, r := range s {
		if i > 0 && r >= 'A' && r <= 'Z' {
			if current.Len() > 0 {
				parts = append(parts, current.String())
				current.Reset()
			}
		}
		current.WriteRune(r)
	}

	if current.Len() > 0 {
		parts = append(parts, current.String())
	}

	return parts
}

func (tm *TypeMapper) IsBuiltInType(typeName string) bool {
	cleanType := tm.CleanTypeName(typeName)
	builtInTypes := map[string]bool{
		"string": true, "normalizedString": true, "token": true, "NMTOKEN": true,
		"Name": true, "NCName": true, "ID": true, "IDREF": true,
		"boolean": true,
		"int":     true, "integer": true, "short": true, "byte": true, "unsignedByte": true,
		"long": true, "unsignedInt": true, "unsignedLong": true, "unsignedShort": true,
		"float": true, "double": true, "decimal": true,
		"dateTime": true, "date": true, "time": true, "duration": true,
		"anyURI": true, "base64Binary": true, "hexBinary": true,
	}
	return builtInTypes[cleanType]
}

func (tm *TypeMapper) GetRequiredImports(mappedTypes []string) []string {
	imports := make(map[string]bool)

	for _, protoType := range mappedTypes {
		switch protoType {
		case "google.protobuf.Timestamp":
			imports["google/protobuf/timestamp.proto"] = true
		case "google.protobuf.Duration":
			imports["google/protobuf/duration.proto"] = true
		case "google.protobuf.Any":
			imports["google/protobuf/any.proto"] = true
		case "google.protobuf.Empty":
			imports["google/protobuf/empty.proto"] = true
		case "google.protobuf.Struct":
			imports["google/protobuf/struct.proto"] = true
		case "google.protobuf.Value":
			imports["google/protobuf/struct.proto"] = true
		case "google.protobuf.ListValue":
			imports["google/protobuf/struct.proto"] = true
		case "google.protobuf.FieldMask":
			imports["google/protobuf/field_mask.proto"] = true
		}
	}

	var result []string
	for imp := range imports {
		result = append(result, imp)
	}

	return result
}
