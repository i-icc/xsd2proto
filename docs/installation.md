# Installation Guide

This document provides detailed instructions for installing xsd2proto.

## Prerequisites

- Go 1.21 or later

## Installation

Install xsd2proto using Go modules:

```bash
go install github.com/i-icc/xsd2proto/cmd/xsd2proto@latest
```

## Verify Installation

After installation, verify that xsd2proto is working correctly:

```bash
# Check version
xsd2proto --version

# View help
xsd2proto --help
```

## Uninstalling

To remove xsd2proto:

```bash
rm $(go env GOPATH)/bin/xsd2proto
```

## Troubleshooting

### "package is not a main package" Error

If you encounter the error:
```
package github.com/i-icc/xsd2proto is not a main package
```

Make sure you're installing from the correct path that includes `/cmd/xsd2proto`:

```bash
# Incorrect (will fail)
go install github.com/i-icc/xsd2proto@latest

# Correct
go install github.com/i-icc/xsd2proto/cmd/xsd2proto@latest
```

### Clear Go Module Cache

If you experience issues with cached modules:

```bash
go clean -modcache
GOPROXY=direct go install github.com/i-icc/xsd2proto/cmd/xsd2proto@latest
```
