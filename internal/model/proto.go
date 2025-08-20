package model

// ProtoFile represents a complete protobuf file
type ProtoFile struct {
	Syntax   string
	Package  string
	Imports  []string
	Options  map[string]string
	Messages []ProtoMessage
	Enums    []ProtoEnum
}

// ProtoMessage represents a protobuf message definition
type ProtoMessage struct {
	Name     string
	Fields   []ProtoField
	Messages []ProtoMessage // nested messages
	Enums    []ProtoEnum    // nested enums
}

// ProtoField represents a field in a protobuf message
type ProtoField struct {
	Name    string
	Type    string
	Number  int
	Label   FieldLabel // optional, required, repeated
	Options map[string]string
}

// ProtoEnum represents a protobuf enum definition
type ProtoEnum struct {
	Name   string
	Values []ProtoEnumValue
}

// ProtoEnumValue represents a value in a protobuf enum
type ProtoEnumValue struct {
	Name   string
	Number int
}

// FieldLabel represents the label of a protobuf field
type FieldLabel int

const (
	FieldLabelOptional FieldLabel = iota
	FieldLabelRequired
	FieldLabelRepeated
)

// String returns the string representation of FieldLabel
func (l FieldLabel) String() string {
	switch l {
	case FieldLabelOptional:
		return "optional"
	case FieldLabelRequired:
		return "required"
	case FieldLabelRepeated:
		return "repeated"
	default:
		return "optional"
	}
}
