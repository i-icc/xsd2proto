package test

import (
	"os"
	"os/exec"
	"strings"
	"testing"
)

// setupTest changes to the project root directory for each test
func setupTest(t *testing.T) {
	// Get current working directory
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get working directory: %v", err)
	}

	// If we're in the test directory, go to parent
	if strings.HasSuffix(wd, "/test") {
		if err := os.Chdir(".."); err != nil {
			t.Fatalf("Failed to change to project root: %v", err)
		}
	}
}

// TestE2EBasicConversion tests end-to-end conversion of the sample XSD file
func TestE2EBasicConversion(t *testing.T) {
	setupTest(t)

	// Build the CLI tool first
	cmd := exec.Command("go", "build", "-o", "xsd2proto_test", "cmd/xsd2proto/main.go")
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to build CLI tool: %v", err)
	}
	defer os.Remove("xsd2proto_test")

	// Test basic conversion using the sample file
	outputFile := "test_output.proto"
	defer os.Remove(outputFile)

	cmd = exec.Command("./xsd2proto_test", "-o", outputFile, "examples/001_simple/simple.xsd")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("CLI conversion failed: %v\nOutput: %s", err, output)
	}

	// Verify output file was created
	if _, err := os.Stat(outputFile); os.IsNotExist(err) {
		t.Fatal("Output file was not created")
	}

	// Read and verify the generated proto content
	content, err := os.ReadFile(outputFile)
	if err != nil {
		t.Fatalf("Failed to read output file: %v", err)
	}

	protoContent := string(content)

	// Verify basic structure
	expectedParts := []string{
		`syntax = "proto3";`,
		"package simple;",
		"enum Status {",
		"STATUS_UNSPECIFIED = 0;",
		"STATUS_ACTIVE = 1;",
		"STATUS_INACTIVE = 2;",
		"STATUS_PENDING = 3;",
		"message Address {",
		"message Person {",
		"string first_name = 1;",
		"repeated string tags = 7;",
	}

	for _, part := range expectedParts {
		if !strings.Contains(protoContent, part) {
			t.Errorf("Generated proto should contain: %s\nActual content:\n%s", part, protoContent)
		}
	}
}

// TestE2EVerboseMode tests CLI with verbose output
func TestE2EVerboseMode(t *testing.T) {
	setupTest(t)
	cmd := exec.Command("go", "build", "-o", "xsd2proto_test", "cmd/xsd2proto/main.go")
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to build CLI tool: %v", err)
	}
	defer os.Remove("xsd2proto_test")

	outputFile := "test_verbose.proto"
	defer os.Remove(outputFile)

	cmd = exec.Command("./xsd2proto_test", "-v", "-o", outputFile, "examples/001_simple/simple.xsd")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("CLI conversion with verbose failed: %v\nOutput: %s", err, output)
	}

	outputStr := string(output)
	if !strings.Contains(outputStr, "Converting") {
		t.Error("Verbose output should contain 'Converting'")
	}

	if !strings.Contains(outputStr, "Successfully") {
		t.Error("Verbose output should contain 'Successfully'")
	}
}

// TestE2EGoPackageOption tests CLI with go_package option
func TestE2EGoPackageOption(t *testing.T) {
	setupTest(t)
	cmd := exec.Command("go", "build", "-o", "xsd2proto_test", "cmd/xsd2proto/main.go")
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to build CLI tool: %v", err)
	}
	defer os.Remove("xsd2proto_test")

	outputFile := "test_gopackage.proto"
	defer os.Remove(outputFile)

	cmd = exec.Command("./xsd2proto_test", "-p", "github.com/example/proto", "-o", outputFile, "examples/001_simple/simple.xsd")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("CLI conversion with go_package failed: %v\nOutput: %s", err, output)
	}

	content, err := os.ReadFile(outputFile)
	if err != nil {
		t.Fatalf("Failed to read output file: %v", err)
	}

	protoContent := string(content)
	if !strings.Contains(protoContent, `option go_package = "github.com/example/proto";`) {
		t.Error("Generated proto should contain go_package option")
	}
}

// TestE2EInvalidFile tests CLI with non-existent file
func TestE2EInvalidFile(t *testing.T) {
	setupTest(t)
	cmd := exec.Command("go", "build", "-o", "xsd2proto_test", "cmd/xsd2proto/main.go")
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to build CLI tool: %v", err)
	}
	defer os.Remove("xsd2proto_test")

	cmd = exec.Command("./xsd2proto_test", "nonexistent.xsd")
	output, err := cmd.CombinedOutput()
	if err == nil {
		t.Error("CLI should fail with non-existent file")
	}

	outputStr := string(output)
	if !strings.Contains(outputStr, "does not exist") {
		t.Errorf("Error message should mention file doesn't exist, got: %s", outputStr)
	}
}

// TestE2EVersionFlag tests CLI version flag
func TestE2EVersionFlag(t *testing.T) {
	setupTest(t)
	cmd := exec.Command("go", "build", "-o", "xsd2proto_test", "cmd/xsd2proto/main.go")
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to build CLI tool: %v", err)
	}
	defer os.Remove("xsd2proto_test")

	cmd = exec.Command("./xsd2proto_test", "--version")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Version command failed: %v\nOutput: %s", err, output)
	}

	outputStr := string(output)
	if !strings.Contains(outputStr, "xsd2proto version") {
		t.Errorf("Version output should contain version info, got: %s", outputStr)
	}
}

// TestE2EHelpFlag tests CLI help flag
func TestE2EHelpFlag(t *testing.T) {
	setupTest(t)
	cmd := exec.Command("go", "build", "-o", "xsd2proto_test", "cmd/xsd2proto/main.go")
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to build CLI tool: %v", err)
	}
	defer os.Remove("xsd2proto_test")

	cmd = exec.Command("./xsd2proto_test", "--help")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Help command failed: %v\nOutput: %s", err, output)
	}

	outputStr := string(output)
	expectedHelpTexts := []string{
		"xsd2proto - Convert XSD files",
		"Usage:",
		"Options:",
		"Examples:",
	}

	for _, text := range expectedHelpTexts {
		if !strings.Contains(outputStr, text) {
			t.Errorf("Help output should contain '%s', got: %s", text, outputStr)
		}
	}
}

// TestE2ENoArguments tests CLI with no arguments
func TestE2ENoArguments(t *testing.T) {
	setupTest(t)
	cmd := exec.Command("go", "build", "-o", "xsd2proto_test", "cmd/xsd2proto/main.go")
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to build CLI tool: %v", err)
	}
	defer os.Remove("xsd2proto_test")

	cmd = exec.Command("./xsd2proto_test")
	output, err := cmd.CombinedOutput()
	if err == nil {
		t.Error("CLI should fail when no arguments provided")
	}

	outputStr := string(output)
	if !strings.Contains(outputStr, "Please provide exactly one XSD input file") {
		t.Errorf("Error message should mention missing input file, got: %s", outputStr)
	}
}

// TestE2EDefaultOutputPath tests CLI with default output path generation
func TestE2EDefaultOutputPath(t *testing.T) {
	setupTest(t)
	cmd := exec.Command("go", "build", "-o", "xsd2proto_test", "cmd/xsd2proto/main.go")
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to build CLI tool: %v", err)
	}
	defer os.Remove("xsd2proto_test")

	// Copy the sample XSD to a temporary location to test default output
	tempXSD := "temp_test.xsd"
	expectedOutput := "temp_test.proto"
	defer os.Remove(tempXSD)
	defer os.Remove(expectedOutput)

	// Copy sample XSD
	input, err := os.ReadFile("examples/001_simple/simple.xsd")
	if err != nil {
		t.Fatalf("Failed to read sample XSD: %v", err)
	}
	if err := os.WriteFile(tempXSD, input, 0644); err != nil {
		t.Fatalf("Failed to write temp XSD: %v", err)
	}

	cmd = exec.Command("./xsd2proto_test", tempXSD)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("CLI conversion failed: %v\nOutput: %s", err, output)
	}

	// Verify output file with default name was created
	if _, err := os.Stat(expectedOutput); os.IsNotExist(err) {
		t.Fatal("Default output file was not created")
	}
}

// TestE2EComplexSchema tests conversion with a more complex XSD
func TestE2EComplexSchema(t *testing.T) {
	setupTest(t)
	cmd := exec.Command("go", "build", "-o", "xsd2proto_test", "cmd/xsd2proto/main.go")
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to build CLI tool: %v", err)
	}
	defer os.Remove("xsd2proto_test")

	// Create a more complex XSD for testing
	complexXSD := `<?xml version="1.0" encoding="UTF-8"?>
<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema"
           targetNamespace="http://example.com/complex"
           xmlns:tns="http://example.com/complex"
           elementFormDefault="qualified">

    <xs:simpleType name="Priority">
        <xs:restriction base="xs:string">
            <xs:enumeration value="LOW"/>
            <xs:enumeration value="MEDIUM"/>
            <xs:enumeration value="HIGH"/>
            <xs:enumeration value="CRITICAL"/>
        </xs:restriction>
    </xs:simpleType>

    <xs:complexType name="Task">
        <xs:sequence>
            <xs:element name="title" type="xs:string"/>
            <xs:element name="description" type="xs:string" minOccurs="0"/>
            <xs:element name="priority" type="tns:Priority"/>
            <xs:element name="assignees" type="xs:string" maxOccurs="unbounded" minOccurs="0"/>
            <xs:element name="dueDate" type="xs:dateTime" minOccurs="0"/>
        </xs:sequence>
        <xs:attribute name="id" type="xs:long" use="required"/>
        <xs:attribute name="completed" type="xs:boolean"/>
    </xs:complexType>

    <xs:element name="task" type="tns:Task"/>

</xs:schema>`

	tempXSD := "complex_test.xsd"
	outputFile := "complex_test.proto"
	defer os.Remove(tempXSD)
	defer os.Remove(outputFile)

	// Write complex XSD
	if err := os.WriteFile(tempXSD, []byte(complexXSD), 0644); err != nil {
		t.Fatalf("Failed to write complex XSD: %v", err)
	}

	cmd = exec.Command("./xsd2proto_test", "-o", outputFile, tempXSD)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Complex XSD conversion failed: %v\nOutput: %s", err, output)
	}

	// Read and verify the generated proto content
	content, err := os.ReadFile(outputFile)
	if err != nil {
		t.Fatalf("Failed to read complex output file: %v", err)
	}

	protoContent := string(content)

	// Verify complex features
	expectedComplexParts := []string{
		"package complex;",
		"enum Priority {",
		"PRIORITY_UNSPECIFIED = 0;",
		"PRIORITY_LOW = 1;",
		"PRIORITY_CRITICAL = 4;",
		"message Task {",
		"google.protobuf.Timestamp due_date",
		"repeated string assignees",
		"int64 id",
		"bool completed",
	}

	for _, part := range expectedComplexParts {
		if !strings.Contains(protoContent, part) {
			t.Errorf("Complex proto should contain: %s\nActual content:\n%s", part, protoContent)
		}
	}

	// Should have timestamp import
	if !strings.Contains(protoContent, `import "google/protobuf/timestamp.proto";`) {
		t.Error("Complex proto should import timestamp.proto")
	}
}
