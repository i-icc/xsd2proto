# CLI Reference

The `xsd2proto` command-line tool provides an easy way to convert XSD files to Protocol Buffer definitions.

## Usage

```
xsd2proto [options] <input.xsd>
```

## Options

| Flag | Long Form | Description | Default |
|------|-----------|-------------|---------|
| `-o` | `--output` | Output file path | Input filename with .proto extension |
| `-p` | `--package` | Go package option for generated proto file | None |
| `-v` | `--verbose` | Enable verbose output | false |
| `-h` | `--help` | Show help message | - |
| | `--version` | Show version information | - |
| | `--no-header` | Disable auto-generation header comment | false |

## Examples

### Basic Conversion

Convert an XSD file to a protobuf file with the same base name:

```bash
xsd2proto schema.xsd
# Output: schema.proto
```

### Custom Output Path

Specify a custom output file path:

```bash
xsd2proto -o /path/to/output.proto schema.xsd
```

### With Go Package Option

Generate protobuf with a specific Go package option:

```bash
xsd2proto -p "github.com/example/proto" schema.xsd
```

This will add the following option to the generated proto file:
```protobuf
option go_package = "github.com/example/proto";
```

### Verbose Output

Enable detailed logging during conversion:

```bash
xsd2proto -v schema.xsd
```

Example verbose output:
```
Converting schema.xsd to protobuf...
Successfully parsed XSD schema with 3 elements, 5 complex types, 2 simple types
Successfully generated schema.proto
```

### Header Comment Control

By default, generated proto files include a header comment indicating they were auto-generated:

To disable the header comment:

```bash
xsd2proto --no-header schema.xsd
```
