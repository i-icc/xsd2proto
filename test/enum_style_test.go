package test

import (
	"os"
	"strings"
	"testing"

	"github.com/i-icc/xsd2proto/internal/converter"
	"github.com/i-icc/xsd2proto/internal/generator"
	"github.com/i-icc/xsd2proto/internal/parser"
)

// TestEnumValuePrefixStyle tests that enum values follow protobuf style guide
func TestEnumValuePrefixStyle(t *testing.T) {
	setupTest(t)

	xsdContent := `<?xml version="1.0" encoding="UTF-8"?>
<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema"
           targetNamespace="http://example.com/test"
           xmlns:tns="http://example.com/test"
           elementFormDefault="qualified">

    <xs:simpleType name="FixtureType">
        <xs:restriction base="xs:string">
            <xs:enumeration value="MATCH"/>
            <xs:enumeration value="OUTRIGHT"/>
            <xs:enumeration value="AGGREGATE"/>
            <xs:enumeration value="VIRTUAL"/>
        </xs:restriction>
    </xs:simpleType>

    <xs:simpleType name="Status">
        <xs:restriction base="xs:string">
            <xs:enumeration value="ACTIVE"/>
            <xs:enumeration value="INACTIVE"/>
        </xs:restriction>
    </xs:simpleType>

</xs:schema>`

	// Write test XSD file
	tmpFile := "test_enum_style.xsd"
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

	// Check FixtureType enum values
	if !strings.Contains(content, "FIXTURE_TYPE_UNSPECIFIED = 0;") {
		t.Error("First FixtureType enum value should be FIXTURE_TYPE_UNSPECIFIED")
	}
	if !strings.Contains(content, "FIXTURE_TYPE_MATCH = 1;") {
		t.Error("Second FixtureType enum value should be FIXTURE_TYPE_MATCH")
	}
	if !strings.Contains(content, "FIXTURE_TYPE_OUTRIGHT = 2;") {
		t.Error("Third FixtureType enum value should be FIXTURE_TYPE_OUTRIGHT")
	}
	if !strings.Contains(content, "FIXTURE_TYPE_AGGREGATE = 3;") {
		t.Error("Fourth FixtureType enum value should be FIXTURE_TYPE_AGGREGATE")
	}
	if !strings.Contains(content, "FIXTURE_TYPE_VIRTUAL = 4;") {
		t.Error("Fifth FixtureType enum value should be FIXTURE_TYPE_VIRTUAL")
	}

	// Check Status enum values
	if !strings.Contains(content, "STATUS_UNSPECIFIED = 0;") {
		t.Error("First Status enum value should be STATUS_UNSPECIFIED")
	}
	if !strings.Contains(content, "STATUS_ACTIVE = 1;") {
		t.Error("Second Status enum value should be STATUS_ACTIVE")
	}
	if !strings.Contains(content, "STATUS_INACTIVE = 2;") {
		t.Error("Third Status enum value should be STATUS_INACTIVE")
	}
}

// TestCamelCaseEnumPrefix tests enum prefix generation with camelCase names
func TestCamelCaseEnumPrefix(t *testing.T) {
	setupTest(t)

	xsdContent := `<?xml version="1.0" encoding="UTF-8"?>
<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema"
           targetNamespace="http://example.com/test"
           xmlns:tns="http://example.com/test"
           elementFormDefault="qualified">

    <xs:simpleType name="userRole">
        <xs:restriction base="xs:string">
            <xs:enumeration value="ADMIN"/>
            <xs:enumeration value="USER"/>
        </xs:restriction>
    </xs:simpleType>

    <xs:simpleType name="paymentMethod">
        <xs:restriction base="xs:string">
            <xs:enumeration value="CREDIT_CARD"/>
            <xs:enumeration value="PAYPAL"/>
        </xs:restriction>
    </xs:simpleType>

</xs:schema>`

	// Write test XSD file
	tmpFile := "test_camel_enum.xsd"
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

	// Check userRole enum values
	if !strings.Contains(content, "USER_ROLE_UNSPECIFIED = 0;") {
		t.Error("First UserRole enum value should be USER_ROLE_UNSPECIFIED")
	}
	if !strings.Contains(content, "USER_ROLE_ADMIN = 1;") {
		t.Error("Second UserRole enum value should be USER_ROLE_ADMIN")
	}

	// Check paymentMethod enum values
	if !strings.Contains(content, "PAYMENT_METHOD_UNSPECIFIED = 0;") {
		t.Error("First PaymentMethod enum value should be PAYMENT_METHOD_UNSPECIFIED")
	}
	if !strings.Contains(content, "PAYMENT_METHOD_CREDIT_CARD = 1;") {
		t.Error("Second PaymentMethod enum value should be PAYMENT_METHOD_CREDIT_CARD")
	}
}
