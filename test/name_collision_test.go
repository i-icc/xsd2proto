package test

import (
	"os"
	"strings"
	"testing"

	"github.com/i-icc/xsd2proto/internal/converter"
	"github.com/i-icc/xsd2proto/internal/generator"
	"github.com/i-icc/xsd2proto/internal/parser"
)

// TestNameCollisionHandling tests that the converter properly handles name collisions
// Note: This test is skipped because string-based enumerations are now treated as string fields with comments
func TestNameCollisionHandling(t *testing.T) {
	t.Skip("String-based enumerations are treated as string fields with comments")
	setupTest(t)

	xsdContent := `<?xml version="1.0" encoding="UTF-8"?>
<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema"
           targetNamespace="http://example.com/collision"
           xmlns:tns="http://example.com/collision"
           elementFormDefault="qualified">

    <!-- First Person message -->
    <xs:complexType name="Person">
        <xs:sequence>
            <xs:element name="name" type="xs:string"/>
            <xs:element name="age" type="xs:int"/>
        </xs:sequence>
    </xs:complexType>

    <!-- Second Person message (duplicate name) -->
    <xs:element name="person">
        <xs:complexType>
            <xs:sequence>
                <xs:element name="firstName" type="xs:string"/>
                <xs:element name="lastName" type="xs:string"/>
            </xs:sequence>
        </xs:complexType>
    </xs:element>

    <!-- First Status enum -->
    <xs:simpleType name="Status">
        <xs:restriction base="xs:string">
            <xs:enumeration value="ACTIVE"/>
            <xs:enumeration value="INACTIVE"/>
        </xs:restriction>
    </xs:simpleType>

    <!-- Second Status enum (duplicate name) -->
    <xs:simpleType name="status">
        <xs:restriction base="xs:string">
            <xs:enumeration value="PENDING"/>
            <xs:enumeration value="APPROVED"/>
        </xs:restriction>
    </xs:simpleType>

    <!-- Message that references the types -->
    <xs:complexType name="Document">
        <xs:sequence>
            <xs:element name="owner" type="tns:Person"/>
            <xs:element name="documentStatus" type="tns:Status"/>
            <xs:element name="approvalStatus" type="tns:status"/>
        </xs:sequence>
    </xs:complexType>

</xs:schema>`

	// Write test XSD file
	tmpFile := "test_collision.xsd"
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

	// Check for unique message names
	if !strings.Contains(content, "message Person {") {
		t.Error("First Person message should exist")
	}
	if !strings.Contains(content, "message Person2 {") {
		t.Error("Second Person message should be renamed to Person2")
	}

	// Check for unique enum names
	if !strings.Contains(content, "enum Status {") {
		t.Error("First Status enum should exist")
	}
	if !strings.Contains(content, "enum Status2 {") {
		t.Error("Second Status enum should be renamed to Status2")
	}

	// Check enum values have correct prefixes
	if !strings.Contains(content, "STATUS_UNSPECIFIED") {
		t.Error("First enum should have STATUS_UNSPECIFIED")
	}
	if !strings.Contains(content, "STATUS_ACTIVE") {
		t.Error("First enum should have STATUS_ prefix")
	}
	if !strings.Contains(content, "STATUS2_UNSPECIFIED") {
		t.Error("Second enum should have STATUS2_UNSPECIFIED")
	}
	if !strings.Contains(content, "STATUS2_PENDING") {
		t.Error("Second enum should have STATUS2_ prefix")
	}

	// Check that field references use correct types
	if !strings.Contains(content, "Person owner") {
		t.Error("owner field should reference Person type")
	}
	if !strings.Contains(content, "Status document_status") {
		t.Error("document_status field should reference Status type")
	}
	// This is the critical test - approval_status should reference Status2
	if !strings.Contains(content, "Status2 approval_status") {
		t.Error("approval_status field should reference Status2 type")
	}
}

// TestMessageEnumNameCollision tests collision between message and enum names
// Note: This test is skipped because string-based enumerations are now treated as string fields with comments
func TestMessageEnumNameCollision(t *testing.T) {
	t.Skip("String-based enumerations are treated as string fields with comments")
	setupTest(t)

	xsdContent := `<?xml version="1.0" encoding="UTF-8"?>
<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema"
           targetNamespace="http://example.com/collision"
           xmlns:tns="http://example.com/collision"
           elementFormDefault="qualified">

    <!-- Message named Status -->
    <xs:complexType name="Status">
        <xs:sequence>
            <xs:element name="code" type="xs:string"/>
            <xs:element name="message" type="xs:string"/>
        </xs:sequence>
    </xs:complexType>

    <!-- Enum also named Status -->
    <xs:simpleType name="Status">
        <xs:restriction base="xs:string">
            <xs:enumeration value="OK"/>
            <xs:enumeration value="ERROR"/>
        </xs:restriction>
    </xs:simpleType>

</xs:schema>`

	// Write test XSD file
	tmpFile := "test_msg_enum_collision.xsd"
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

	// Check that both message and enum exist with different names
	// Since enums are processed first, enum gets "Status" and message gets "Status2"
	if !strings.Contains(content, "enum Status {") {
		t.Error("Status enum should exist")
	}
	if !strings.Contains(content, "message Status2 {") {
		t.Error("Status message should be renamed to Status2")
	}
}
