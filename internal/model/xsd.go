package model

import "encoding/xml"

// Schema represents the root element of an XSD document
type Schema struct {
	XMLName              xml.Name      `xml:"http://www.w3.org/2001/XMLSchema schema"`
	TargetNamespace      string        `xml:"targetNamespace,attr"`
	ElementFormDefault   string        `xml:"elementFormDefault,attr"`
	AttributeFormDefault string        `xml:"attributeFormDefault,attr"`
	Imports              []Import      `xml:"import"`
	Includes             []Include     `xml:"include"`
	Elements             []Element     `xml:"element"`
	ComplexTypes         []ComplexType `xml:"complexType"`
	SimpleTypes          []SimpleType  `xml:"simpleType"`

	ImportedSchemas []*Schema `xml:"-"`
}

// Element represents an XSD element definition
type Element struct {
	Name        string       `xml:"name,attr"`
	Type        string       `xml:"type,attr"`
	MinOccurs   string       `xml:"minOccurs,attr"`
	MaxOccurs   string       `xml:"maxOccurs,attr"`
	ComplexType *ComplexType `xml:"complexType"`
	SimpleType  *SimpleType  `xml:"simpleType"`
}

// ComplexType represents an XSD complex type definition
type ComplexType struct {
	Name       string      `xml:"name,attr"`
	Sequence   *Sequence   `xml:"sequence"`
	Choice     *Choice     `xml:"choice"`
	Attributes []Attribute `xml:"attribute"`
}

// SimpleType represents an XSD simple type definition
type SimpleType struct {
	Name        string       `xml:"name,attr"`
	Restriction *Restriction `xml:"restriction"`
	Union       *Union       `xml:"union"`
	List        *List        `xml:"list"`
}

// Sequence represents an ordered group of elements
type Sequence struct {
	Elements  []Element `xml:"element"`
	MinOccurs string    `xml:"minOccurs,attr"`
	MaxOccurs string    `xml:"maxOccurs,attr"`
}

// Choice represents a choice between multiple elements
type Choice struct {
	Elements  []Element `xml:"element"`
	MinOccurs string    `xml:"minOccurs,attr"`
	MaxOccurs string    `xml:"maxOccurs,attr"`
}

// Attribute represents an XSD attribute
type Attribute struct {
	Name string `xml:"name,attr"`
	Type string `xml:"type,attr"`
	Use  string `xml:"use,attr"`
}

// Restriction represents type restrictions
type Restriction struct {
	Base         string        `xml:"base,attr"`
	Enumerations []Enumeration `xml:"enumeration"`
	Pattern      *Pattern      `xml:"pattern"`
	MinLength    *Length       `xml:"minLength"`
	MaxLength    *Length       `xml:"maxLength"`
}

// Pattern represents a pattern restriction
type Pattern struct {
	Value string `xml:"value,attr"`
}

// Length represents length restrictions
type Length struct {
	Value int `xml:"value,attr"`
}

// Enumeration represents an enumeration value
type Enumeration struct {
	Value string `xml:"value,attr"`
}

// Union represents a union of types
type Union struct {
	MemberTypes string `xml:"memberTypes,attr"`
}

// List represents a list type
type List struct {
	ItemType string `xml:"itemType,attr"`
}

// Import represents an XSD import directive
type Import struct {
	Namespace      string `xml:"namespace,attr"`
	SchemaLocation string `xml:"schemaLocation,attr"`
}

// Include represents an XSD include directive
type Include struct {
	SchemaLocation string `xml:"schemaLocation,attr"`
}
