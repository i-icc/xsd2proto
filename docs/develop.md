# Development Guide

## Project Structure

```
xsd2proto/
├── cmd/xsd2proto/          # CLI entry point
├── internal/
├── examples/              # Test cases and examples
└── docs/                 # Documentation
```

## Building

```bash
go build -o xsd2proto cmd/xsd2proto/main.go
```

## Testing

Run conversion examples:
```bash
./xsd2proto examples/001_simple/simple.xsd
```

## Code Style

- Remove obvious/unnecessary comments
- Use English for all comments and documentation
- Follow Go conventions for naming