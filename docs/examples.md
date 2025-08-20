# Examples

This document provides various examples of XSD to Protocol Buffer conversion.

## Example 001_simple: Simple Types and Enumerations

### Input XSD (001_simple)

```xml
<?xml version="1.0" encoding="UTF-8"?>
<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema"
           targetNamespace="http://example.com/simple"
           xmlns:tns="http://example.com/simple"
           elementFormDefault="qualified">

    <xs:simpleType name="Status">
        <xs:restriction base="xs:string">
            <xs:enumeration value="ACTIVE"/>
            <xs:enumeration value="INACTIVE"/>
            <xs:enumeration value="PENDING"/>
        </xs:restriction>
    </xs:simpleType>

    <xs:complexType name="Address">
        <xs:sequence>
            <xs:element name="street" type="xs:string"/>
            <xs:element name="city" type="xs:string"/>
            <xs:element name="postalCode" type="xs:string"/>
            <xs:element name="country" type="xs:string"/>
        </xs:sequence>
    </xs:complexType>

    <xs:complexType name="Person">
        <xs:sequence>
            <xs:element name="firstName" type="xs:string"/>
            <xs:element name="lastName" type="xs:string"/>
            <xs:element name="age" type="xs:int"/>
            <xs:element name="email" type="xs:string" minOccurs="0"/>
            <xs:element name="address" type="tns:Address"/>
            <xs:element name="status" type="tns:Status"/>
            <xs:element name="tags" type="xs:string" maxOccurs="unbounded" minOccurs="0"/>
        </xs:sequence>
        <xs:attribute name="id" type="xs:string" use="required"/>
    </xs:complexType>

    <xs:element name="person" type="tns:Person"/>

</xs:schema>
```

### Output Proto

```protobuf
syntax = "proto3";

package simple;

enum Status {
  ACTIVE = 0;
  INACTIVE = 1;
  PENDING = 2;
}

message Address {
  string street = 1;
  string city = 2;
  string postal_code = 3;
  string country = 4;
}

message Person {
  string first_name = 1;
  string last_name = 2;
  int32 age = 3;
  string email = 4;
  Address address = 5;
  Status status = 6;
  repeated string tags = 7;
  string id = 8;
}
```

### Key Features Demonstrated

- Enumeration conversion
- Complex type nesting
- Optional elements (`minOccurs="0"`)
- Repeated elements (`maxOccurs="unbounded"`)
- Attributes conversion
- Package generation from namespace

## Command Examples

### Convert with CLI

```bash
# Basic conversion
xsd2proto examples/001_simple/simple.xsd

# With verbose output
xsd2proto -v examples/001_simple/simple.xsd

# With custom output
xsd2proto -o generated.proto examples/001_simple/simple.xsd
```

### Convert with Library

```go
package main

import (
    "log"
    "xsd2proto"
)

func main() {
    opts := xsd2proto.Options{
        Verbose: true,
    }
    
    converter := xsd2proto.NewConverter(opts)
    
    if err := converter.Convert("examples/001_simple/simple.xsd"); err != nil {
        log.Fatal(err)
    }
}
```

## Common Patterns

### 1. Basic Data Types

```xml
<xs:element name="name" type="xs:string"/>
<xs:element name="age" type="xs:int"/>
<xs:element name="active" type="xs:boolean"/>
```

Converts to:

```protobuf
string name = 1;
int32 age = 2;
bool active = 3;
```

### 2. Optional Fields

```xml
<xs:element name="email" type="xs:string" minOccurs="0"/>
```

Converts to:

```protobuf
string email = 1;  // Optional in proto3
```

### 3. Repeated Fields

```xml
<xs:element name="tags" type="xs:string" maxOccurs="unbounded"/>
```

Converts to:

```protobuf
repeated string tags = 1;
```

### 4. Nested Messages

```xml
<xs:complexType name="Order">
    <xs:sequence>
        <xs:element name="id" type="xs:string"/>
        <xs:element name="customer" type="tns:Customer"/>
    </xs:sequence>
</xs:complexType>
```

Converts to:

```protobuf
message Order {
  string id = 1;
  Customer customer = 2;
}
```

## Best Practices

### XSD Design for Better Proto Output

1. **Use meaningful names**: Field names should be descriptive and follow naming conventions.

2. **Avoid deep nesting**: Keep message structures reasonably flat.

3. **Use enumerations**: Define string restrictions as enumerations when possible.

4. **Explicit types**: Prefer named types over anonymous inline types.

### Example of Well-Structured XSD

```xml
<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema"
           targetNamespace="http://example.com/api"
           elementFormDefault="qualified">

    <!-- Well-defined enumerations -->
    <xs:simpleType name="UserRole">
        <xs:restriction base="xs:string">
            <xs:enumeration value="ADMIN"/>
            <xs:enumeration value="USER"/>
            <xs:enumeration value="GUEST"/>
        </xs:restriction>
    </xs:simpleType>

    <!-- Clear, reusable types -->
    <xs:complexType name="User">
        <xs:sequence>
            <xs:element name="username" type="xs:string"/>
            <xs:element name="role" type="tns:UserRole"/>
            <xs:element name="createdAt" type="xs:dateTime"/>
        </xs:sequence>
        <xs:attribute name="id" type="xs:long" use="required"/>
    </xs:complexType>

</xs:schema>
```

This produces clean, maintainable Protocol Buffer definitions.
