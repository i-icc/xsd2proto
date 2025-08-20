package converter

import (
	"fmt"
	"strings"

	"github.com/i-icc/xsd2proto/internal/model"
)

// Converter converts XSD schemas to Protobuf definitions
type Converter struct {
	typeMapper        *TypeMapper
	fieldCounter      int
	usedEnumValues    map[string]bool
	enumValueCounters map[string]int
	usedMessageNames  map[string]bool   // Track used message names
	usedEnumNames     map[string]bool   // Track used enum names
	typeRenameMap     map[string]string // Map from original type name to renamed type name
	useCamelCase      bool              // Use camelCase for field names instead of snake_case
	usePascalCase     bool              // Use PascalCase for field names instead of snake_case
	currentSchema     *model.Schema     // Reference to current schema for ArrayOf optimization
}

// New creates a new converter instance
func New() *Converter {
	return &Converter{
		typeMapper:        NewTypeMapper(),
		fieldCounter:      1,
		usedEnumValues:    make(map[string]bool),
		enumValueCounters: make(map[string]int),
		usedMessageNames:  make(map[string]bool),
		usedEnumNames:     make(map[string]bool),
		typeRenameMap:     make(map[string]string),
		useCamelCase:      false,
		usePascalCase:     false,
	}
}

// SetFieldNamingStyle sets the field naming style
func (c *Converter) SetFieldNamingStyle(useCamelCase, usePascalCase bool) {
	c.useCamelCase = useCamelCase
	c.usePascalCase = usePascalCase
}

// Convert converts an XSD schema to a Protobuf file model
func (c *Converter) Convert(schema *model.Schema) (*model.ProtoFile, error) {
	// Store schema reference for ArrayOf optimization
	c.currentSchema = schema

	protoFile := &model.ProtoFile{
		Syntax:  "proto3",
		Package: c.generatePackageName(schema.TargetNamespace),
		Options: make(map[string]string),
	}

	// First pass: convert all simple types (enums)
	for _, simpleType := range schema.SimpleTypes {
		if simpleType.Restriction != nil && len(simpleType.Restriction.Enumerations) > 0 {
			enum := c.convertSimpleTypeToEnum(&simpleType)
			protoFile.Enums = append(protoFile.Enums, *enum)
		}
	}

	// Second pass: convert all complex types (messages)
	for _, complexType := range schema.ComplexTypes {
		// Skip ArrayOf pattern types - they will be converted to direct repeated fields
		if c.isArrayOfPattern(&complexType) {
			continue
		}
		message, err := c.convertComplexType(&complexType)
		if err != nil {
			return nil, fmt.Errorf("failed to convert complex type %s: %w", complexType.Name, err)
		}
		protoFile.Messages = append(protoFile.Messages, *message)
	}

	// Third pass: convert all elements
	for _, element := range schema.Elements {
		if element.ComplexType != nil {
			message, err := c.convertElementToMessage(&element)
			if err != nil {
				return nil, fmt.Errorf("failed to convert element %s: %w", element.Name, err)
			}
			protoFile.Messages = append(protoFile.Messages, *message)
		}
	}

	var mappedTypes []string
	for _, message := range protoFile.Messages {
		for _, field := range message.Fields {
			mappedTypes = append(mappedTypes, field.Type)
		}
	}
	protoFile.Imports = c.typeMapper.GetRequiredImports(mappedTypes)

	return protoFile, nil
}

func (c *Converter) convertComplexType(complexType *model.ComplexType) (*model.ProtoMessage, error) {
	message := &model.ProtoMessage{
		Name: c.generateUniqueMessageName(complexType.Name),
	}

	c.fieldCounter = 1

	// Process sequence elements
	if complexType.Sequence != nil {
		for _, element := range complexType.Sequence.Elements {
			field, err := c.convertElementToField(&element)
			if err != nil {
				return nil, err
			}
			message.Fields = append(message.Fields, *field)
		}
	}

	if complexType.Choice != nil {
		for _, element := range complexType.Choice.Elements {
			field, err := c.convertElementToField(&element)
			if err != nil {
				return nil, err
			}
			field.Label = model.FieldLabelOptional
			message.Fields = append(message.Fields, *field)
		}
	}

	// Process attributes as fields
	for _, attribute := range complexType.Attributes {
		field, err := c.convertAttributeToField(&attribute)
		if err != nil {
			return nil, err
		}
		message.Fields = append(message.Fields, *field)
	}

	return message, nil
}

func (c *Converter) convertElementToMessage(element *model.Element) (*model.ProtoMessage, error) {
	if element.ComplexType == nil {
		return nil, fmt.Errorf("element %s has no complex type", element.Name)
	}

	messageName := element.Name
	if element.ComplexType.Name != "" {
		messageName = element.ComplexType.Name
	}

	element.ComplexType.Name = messageName
	return c.convertComplexType(element.ComplexType)
}

func (c *Converter) convertElementToField(element *model.Element) (*model.ProtoField, error) {
	protoType, err := c.typeMapper.MapXSDType(element.Type)
	if err != nil {
		return nil, err
	}

	// Check if this field references an ArrayOf pattern type
	arrayElementType := c.getArrayOfElementType(element.Type)
	if arrayElementType != "" {
		// Convert ArrayOf reference to direct repeated field
		field := &model.ProtoField{
			Name:   c.formatFieldName(element.Name),
			Type:   arrayElementType,
			Number: c.fieldCounter,
			Label:  model.FieldLabelRepeated,
		}
		c.fieldCounter++
		return field, nil
	}

	// If the type has been renamed, use the new name
	if !c.typeMapper.IsBuiltInType(element.Type) {
		// For custom types, check if they have been renamed
		cleanType := c.typeMapper.CleanTypeName(element.Type)

		// First, check with the cleaned type name
		if renamedType, exists := c.typeRenameMap[cleanType]; exists {
			protoType = renamedType
		} else {
			// If not found, check if the Pascal case version has been renamed
			// This handles case-insensitive type references
			pascalCaseType := c.toPascalCase(cleanType)
			if renamedType, exists := c.typeRenameMap[pascalCaseType]; exists {
				protoType = renamedType
			}
		}
	}

	field := &model.ProtoField{
		Name:   c.formatFieldName(element.Name),
		Type:   protoType,
		Number: c.fieldCounter,
		Label:  c.determineFieldLabel(element.MinOccurs, element.MaxOccurs),
	}

	c.fieldCounter++
	return field, nil
}

func (c *Converter) convertAttributeToField(attribute *model.Attribute) (*model.ProtoField, error) {
	protoType, err := c.typeMapper.MapXSDType(attribute.Type)
	if err != nil {
		return nil, err
	}

	// If the type has been renamed, use the new name
	if !c.typeMapper.IsBuiltInType(attribute.Type) {
		// For custom types, check if they have been renamed
		cleanType := c.typeMapper.CleanTypeName(attribute.Type)

		// First, check with the cleaned type name
		if renamedType, exists := c.typeRenameMap[cleanType]; exists {
			protoType = renamedType
		} else {
			// If not found, check if the Pascal case version has been renamed
			// This handles case-insensitive type references
			pascalCaseType := c.toPascalCase(cleanType)
			if renamedType, exists := c.typeRenameMap[pascalCaseType]; exists {
				protoType = renamedType
			}
		}
	}

	field := &model.ProtoField{
		Name:   c.formatFieldName(attribute.Name),
		Type:   protoType,
		Number: c.fieldCounter,
		Label:  c.determineAttributeLabel(attribute.Use),
	}

	c.fieldCounter++
	return field, nil
}

func (c *Converter) convertSimpleTypeToEnum(simpleType *model.SimpleType) *model.ProtoEnum {
	uniqueEnumName := c.generateUniqueEnumName(simpleType.Name)
	enum := &model.ProtoEnum{
		Name: uniqueEnumName,
	}

	// First, add the UNSPECIFIED value at index 0
	unspecifiedValue := model.ProtoEnumValue{
		Name:   c.generateUniqueEnumValueName(uniqueEnumName, "UNSPECIFIED", true),
		Number: 0,
	}
	enum.Values = append(enum.Values, unspecifiedValue)

	// Then add all the actual enum values starting from index 1
	for i, enumeration := range simpleType.Restriction.Enumerations {
		enumValue := model.ProtoEnumValue{
			Name:   c.generateUniqueEnumValueName(uniqueEnumName, enumeration.Value, false),
			Number: i + 1,
		}
		enum.Values = append(enum.Values, enumValue)
	}

	return enum
}

func (c *Converter) formatMessageName(name string) string {
	return c.toPascalCase(name)
}

func (c *Converter) formatFieldName(name string) string {
	if c.usePascalCase {
		return c.toPascalCase(name)
	}
	if c.useCamelCase {
		return c.toCamelCase(name)
	}
	return c.toSnakeCase(name)
}

func (c *Converter) formatEnumName(name string) string {
	return c.toPascalCase(name)
}

// generateUniqueEnumValueName adds prefix to enum value names to avoid duplicates
func (c *Converter) generateUniqueEnumValueName(enumName, valueName string, isFirstValue bool) string {
	// Convert enum name to snake_case for the prefix (e.g., FixtureType -> fixture_type -> FIXTURE_TYPE)
	enumSnakeCase := c.toSnakeCase(enumName)
	enumPrefix := strings.ToUpper(enumSnakeCase)

	// For the first value (index 0), add _UNSPECIFIED suffix according to protobuf style guide
	var prefixedName string
	if isFirstValue {
		prefixedName = enumPrefix + "_UNSPECIFIED"
	} else {
		basicValueName := strings.ToUpper(valueName)
		prefixedName = enumPrefix + "_" + basicValueName
	}

	if !c.usedEnumValues[prefixedName] {
		c.usedEnumValues[prefixedName] = true
		return prefixedName
	}

	baseKey := prefixedName
	counter := c.enumValueCounters[baseKey]
	for {
		counter++
		candidateName := fmt.Sprintf("%s%d", prefixedName, counter)
		if !c.usedEnumValues[candidateName] {
			c.usedEnumValues[candidateName] = true
			c.enumValueCounters[baseKey] = counter
			return candidateName
		}
	}
}

// generateUniqueMessageName ensures message names are unique
func (c *Converter) generateUniqueMessageName(originalName string) string {
	formattedName := c.formatMessageName(originalName)

	// Check if the formatted name is already used by either messages or enums
	if !c.usedMessageNames[formattedName] && !c.usedEnumNames[formattedName] {
		c.usedMessageNames[formattedName] = true
		c.typeRenameMap[originalName] = formattedName
		return formattedName
	}

	// If the name is already used, append a number
	counter := 2
	for {
		candidateName := fmt.Sprintf("%s%d", formattedName, counter)
		if !c.usedMessageNames[candidateName] && !c.usedEnumNames[candidateName] {
			c.usedMessageNames[candidateName] = true
			c.typeRenameMap[originalName] = candidateName
			return candidateName
		}
		counter++
	}
}

// generateUniqueEnumName ensures enum names are unique
func (c *Converter) generateUniqueEnumName(originalName string) string {
	formattedName := c.formatEnumName(originalName)

	// Check if the formatted name is already used by either messages or enums
	if !c.usedEnumNames[formattedName] && !c.usedMessageNames[formattedName] {
		c.usedEnumNames[formattedName] = true
		c.typeRenameMap[originalName] = formattedName
		return formattedName
	}

	// If the name is already used, append a number
	counter := 2
	for {
		candidateName := fmt.Sprintf("%s%d", formattedName, counter)
		if !c.usedEnumNames[candidateName] && !c.usedMessageNames[candidateName] {
			c.usedEnumNames[candidateName] = true
			c.typeRenameMap[originalName] = candidateName
			return candidateName
		}
		counter++
	}
}

func (c *Converter) generatePackageName(targetNamespace string) string {
	if targetNamespace == "" {
		return "generated"
	}

	if strings.HasPrefix(targetNamespace, "http://") || strings.HasPrefix(targetNamespace, "https://") {
		path := strings.TrimPrefix(targetNamespace, "http://")
		path = strings.TrimPrefix(path, "https://")

		parts := strings.Split(path, "/")

		// For simple cases like "example.com/simple", use only the last part
		if len(parts) == 2 {
			return parts[1]
		}

		// For complex domains, use last meaningful part for simple schemas
		if len(parts) > 1 {
			lastPart := parts[len(parts)-1]
			if lastPart != "" {
				return lastPart
			}
		}

		// Fallback to domain-based naming for complex cases
		domain := parts[0]
		var packageParts []string

		if domain != "" {
			domainParts := strings.Split(domain, ".")
			packageParts = append(packageParts, domainParts...)
		}

		if len(parts) > 1 {
			for _, part := range parts[1:] {
				if part != "" {
					packageParts = append(packageParts, part)
				}
			}
		}

		return strings.Join(packageParts, ".")
	}

	if strings.HasPrefix(targetNamespace, "./") {
		path := strings.TrimPrefix(targetNamespace, "./")
		parts := strings.Split(path, "/")

		var packageParts []string
		for _, part := range parts {
			if part != "" {
				packageParts = append(packageParts, part)
			}
		}

		return strings.Join(packageParts, ".")
	}

	if strings.HasPrefix(targetNamespace, "urn:") {
		parts := strings.Split(targetNamespace, ":")
		if len(parts) > 0 {
			return parts[len(parts)-1]
		}
	}

	result := strings.ReplaceAll(targetNamespace, "/", ".")
	result = strings.ReplaceAll(result, "_", ".")
	result = strings.TrimPrefix(result, ".")
	return result
}

func (c *Converter) determineFieldLabel(minOccurs, maxOccurs string) model.FieldLabel {
	// Handle repeated fields first
	if maxOccurs == "unbounded" || (maxOccurs != "" && maxOccurs != "1") {
		return model.FieldLabelRepeated
	}

	// Handle optional fields: minOccurs="0" means optional
	if minOccurs == "0" {
		return model.FieldLabelOptional
	}

	// Default case: minOccurs="1" or unspecified means required
	// In proto3, all fields are technically optional, but we track semantic meaning
	return model.FieldLabelRequired
}

func (c *Converter) determineAttributeLabel(use string) model.FieldLabel {
	// XSD attributes: use="required" means required, use="optional" or unspecified means optional
	if use == "required" {
		return model.FieldLabelRequired
	}
	return model.FieldLabelOptional
}

func (c *Converter) toPascalCase(s string) string {
	parts := strings.FieldsFunc(s, func(r rune) bool {
		return r == '_' || r == '-' || r == '.'
	})

	var finalParts []string
	for _, part := range parts {
		camelParts := c.splitCamelCase(part)
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

// isArrayOfPattern checks if a complex type follows the ArrayOf pattern
// Returns true if the type name starts with "ArrayOf" and contains only one repeated element
func (c *Converter) isArrayOfPattern(complexType *model.ComplexType) bool {
	cleanName := c.typeMapper.CleanTypeName(complexType.Name)

	// Check if name starts with "ArrayOf"
	if !strings.HasPrefix(cleanName, "ArrayOf") {
		return false
	}

	// Check if it has exactly one sequence element with unbounded occurrence
	if complexType.Sequence == nil || len(complexType.Sequence.Elements) != 1 {
		return false
	}

	element := complexType.Sequence.Elements[0]
	return element.MaxOccurs == "unbounded" || (element.MaxOccurs != "" && element.MaxOccurs != "1")
}

// getArrayOfElementType returns the element type from an ArrayOf type reference
// Returns empty string if not an ArrayOf pattern
func (c *Converter) getArrayOfElementType(typeName string) string {
	cleanType := c.typeMapper.CleanTypeName(typeName)

	// Check if name starts with "ArrayOf"
	if !strings.HasPrefix(cleanType, "ArrayOf") {
		return ""
	}

	// Find the corresponding complex type in the schema
	if c.currentSchema == nil {
		return ""
	}

	for _, complexType := range c.currentSchema.ComplexTypes {
		if c.typeMapper.CleanTypeName(complexType.Name) == cleanType {
			if c.isArrayOfPattern(&complexType) {
				// Extract the element type from the single repeated element
				element := complexType.Sequence.Elements[0]
				elementType := c.typeMapper.CleanTypeName(element.Type)

				// Return the properly formatted element type name
				if c.typeMapper.IsBuiltInType(element.Type) {
					protoType, _ := c.typeMapper.MapXSDType(element.Type)
					return protoType
				}

				// For custom types, return the Pascal case formatted name
				return c.toPascalCase(elementType)
			}
		}
	}

	return ""
}

func (c *Converter) splitCamelCase(s string) []string {
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

func (c *Converter) toSnakeCase(s string) string {
	var result strings.Builder
	for i, r := range s {
		if i > 0 && (r >= 'A' && r <= 'Z') {
			result.WriteRune('_')
		}
		if r == '-' || r == '.' {
			result.WriteRune('_')
		} else {
			result.WriteRune(r)
		}
	}
	return strings.ToLower(result.String())
}

func (c *Converter) toCamelCase(s string) string {
	parts := strings.FieldsFunc(s, func(r rune) bool {
		return r == '_' || r == '-' || r == '.'
	})

	var finalParts []string
	for _, part := range parts {
		camelParts := c.splitCamelCase(part)
		finalParts = append(finalParts, camelParts...)
	}

	var result strings.Builder
	for i, part := range finalParts {
		if len(part) > 0 {
			if i == 0 {
				// First part: lowercase first character (camelCase, not PascalCase)
				result.WriteString(strings.ToLower(part[:1]))
				if len(part) > 1 {
					result.WriteString(strings.ToLower(part[1:]))
				}
			} else {
				// Subsequent parts: uppercase first character
				result.WriteString(strings.ToUpper(part[:1]))
				if len(part) > 1 {
					result.WriteString(strings.ToLower(part[1:]))
				}
			}
		}
	}
	return result.String()
}
