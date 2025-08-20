package test

import (
	"os"
	"strings"
	"testing"

	"github.com/i-icc/xsd2proto/internal/converter"
	"github.com/i-icc/xsd2proto/internal/generator"
	"github.com/i-icc/xsd2proto/internal/parser"
)

// TestDurationImport tests that duration.proto is imported when duration type is used
func TestDurationImport(t *testing.T) {
	setupTest(t)

	xsdContent := `<?xml version="1.0" encoding="UTF-8"?>
<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema"
           targetNamespace="http://example.com/duration"
           xmlns:tns="http://example.com/duration"
           elementFormDefault="qualified">

    <xs:complexType name="Timer">
        <xs:sequence>
            <xs:element name="name" type="xs:string"/>
            <xs:element name="interval" type="xs:duration"/>
            <xs:element name="timeout" type="xs:duration" minOccurs="0"/>
        </xs:sequence>
    </xs:complexType>

    <xs:complexType name="Event">
        <xs:sequence>
            <xs:element name="title" type="xs:string"/>
            <xs:element name="duration" type="xs:duration"/>
            <xs:element name="startTime" type="xs:dateTime"/>
        </xs:sequence>
    </xs:complexType>

</xs:schema>`

	// Write test XSD file
	tmpFile := "test_duration.xsd"
	if err := os.WriteFile(tmpFile, []byte(xsdContent), 0644); err != nil {
		t.Fatalf("Failed to write test XSD: %v", err)
	}
	defer os.Remove(tmpFile)

	// Parse XSD
	p := parser.New()
	schema, err := p.ParseFile(tmpFile)
	if err != nil {
		t.Fatalf("Failed to parse XSD: %v", err)
	}

	// Convert to proto
	conv := converter.New()
	protoFile, err := conv.Convert(schema)
	if err != nil {
		t.Fatalf("Failed to convert schema: %v", err)
	}

	// Generate proto content
	gen := generator.New()
	gen.SetHeaderOptions(false, "")
	content, err := gen.Generate(protoFile)
	if err != nil {
		t.Fatalf("Failed to generate proto: %v", err)
	}

	// Check that duration.proto is imported
	if !strings.Contains(content, `import "google/protobuf/duration.proto";`) {
		t.Error("Generated proto should import google/protobuf/duration.proto")
	}

	// Check that timestamp.proto is also imported (for dateTime)
	if !strings.Contains(content, `import "google/protobuf/timestamp.proto";`) {
		t.Error("Generated proto should import google/protobuf/timestamp.proto")
	}

	// Check that duration fields use the correct type
	if !strings.Contains(content, "google.protobuf.Duration interval") {
		t.Error("interval field should use google.protobuf.Duration type")
	}
	if !strings.Contains(content, "google.protobuf.Duration timeout") {
		t.Error("timeout field should use google.protobuf.Duration type")
	}
	if !strings.Contains(content, "google.protobuf.Duration duration") {
		t.Error("duration field should use google.protobuf.Duration type")
	}

	// Check that timestamp field uses the correct type
	if !strings.Contains(content, "google.protobuf.Timestamp start_time") {
		t.Error("start_time field should use google.protobuf.Timestamp type")
	}
}

// TestOnlyTimestampImport tests that only timestamp.proto is imported when only dateTime is used
func TestOnlyTimestampImport(t *testing.T) {
	setupTest(t)

	xsdContent := `<?xml version="1.0" encoding="UTF-8"?>
<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema"
           targetNamespace="http://example.com/timestamp"
           xmlns:tns="http://example.com/timestamp"
           elementFormDefault="qualified">

    <xs:complexType name="Log">
        <xs:sequence>
            <xs:element name="message" type="xs:string"/>
            <xs:element name="createdAt" type="xs:dateTime"/>
            <xs:element name="updatedAt" type="xs:dateTime"/>
        </xs:sequence>
    </xs:complexType>

</xs:schema>`

	// Write test XSD file
	tmpFile := "test_timestamp.xsd"
	if err := os.WriteFile(tmpFile, []byte(xsdContent), 0644); err != nil {
		t.Fatalf("Failed to write test XSD: %v", err)
	}
	defer os.Remove(tmpFile)

	// Parse XSD
	p := parser.New()
	schema, err := p.ParseFile(tmpFile)
	if err != nil {
		t.Fatalf("Failed to parse XSD: %v", err)
	}

	// Convert to proto
	conv := converter.New()
	protoFile, err := conv.Convert(schema)
	if err != nil {
		t.Fatalf("Failed to convert schema: %v", err)
	}

	// Generate proto content
	gen := generator.New()
	gen.SetHeaderOptions(false, "")
	content, err := gen.Generate(protoFile)
	if err != nil {
		t.Fatalf("Failed to generate proto: %v", err)
	}

	// Check that timestamp.proto is imported
	if !strings.Contains(content, `import "google/protobuf/timestamp.proto";`) {
		t.Error("Generated proto should import google/protobuf/timestamp.proto")
	}

	// Check that duration.proto is NOT imported
	if strings.Contains(content, `import "google/protobuf/duration.proto";`) {
		t.Error("Generated proto should NOT import google/protobuf/duration.proto when not needed")
	}
}
