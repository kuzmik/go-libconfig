# go-libconfig

A pure Go implementation of a [libconfig](https://hyperrealm.github.io/libconfig/) parser. This library allows you to parse configuration files that follow the libconfig specification, which provides a structured configuration file format that is more compact and readable than XML or JSON.

## Features

- **Full libconfig specification support**: Scalars, arrays, groups, lists
- **Multiple integer formats**: Decimal, hexadecimal (`0xFF`), binary (`0b1010`), octal (`0o755`)
- **Flexible value types**: Strings, integers, floats, booleans
- **Complex data structures**: Nested groups, arrays, and heterogeneous lists
- **String features**: Escape sequences, concatenation, Unicode support
- **Include directives**: `@include` support for modular configurations
- **Robust parsing**: Comprehensive error handling and reporting
- **Go-idiomatic API**: Type-safe value lookup methods
- **Static error types**: Predefined errors for `errors.Is()` checking
- **Production ready**: Comprehensive test suite, benchmarks, and linting

## Installation

```bash
go get github.com/kuzmik/go-libconfig
```

## Quick Start

```go
package main

import (
    "fmt"
    "log"

    "github.com/kuzmik/go-libconfig"
)

func main() {
    config, err := libconfig.ParseString(`
        name = "MyApp";
        port = 8080;
        debug = true;

        database = {
            host = "localhost";
            port = 5432;
        };

        servers = [ "web1", "web2", "web3" ];
    `)

    if err != nil {
        log.Fatal(err)
    }

    // Look up values by path
    name, _ := config.LookupString("name")
    port, _ := config.LookupInt("port")
    dbHost, _ := config.LookupString("database.host")

    fmt.Printf("App: %s, Port: %d, DB: %s\n", name, port, dbHost)
}
```

## Configuration Format

The libconfig format supports various data types and structures:

### Scalars

```libconfig
# Strings
name = "My Application";
description = "A sample application";

# Integers (various formats)
port = 8080;                    # Decimal
flags = 0xFF;                   # Hexadecimal
perms = 0o755;                  # Octal
binary = 0b1010;               # Binary
big_num = 9223372036854775807L; # 64-bit integer

# Floats
pi = 3.14159;
scientific = 1.23e-4;

# Booleans
debug = true;
production = false;
```

### Groups (Objects)

```libconfig
database = {
    host = "localhost";
    port = 5432;
    credentials = {
        username = "admin";
        password = "secret";
    };
};
```

### Arrays (Homogeneous)

```libconfig
# All elements must be the same type
servers = [ "web1", "web2", "web3" ];
ports = [ 8080, 8081, 8082 ];
weights = [ 1.0, 0.8, 0.6 ];
flags = [ true, false, true ];
```

### Lists (Heterogeneous)

```libconfig
# Mixed types allowed
mixed = ( "string", 42, true, 3.14, { key = "value"; } );
```

### String Features

```libconfig
# String concatenation
long_text = "This is a very long string that "
           "spans multiple lines automatically.";

# Escape sequences
escaped = "Line 1\nLine 2\tTabbed text\rCarriage return";
unicode = "Unicode: \x41\x42\x43";  # ABC
quotes = "He said, \"Hello there!\"";
```

### Include Directives

```libconfig
# main.cfg
@include "database.cfg"
@include "logging.cfg"

application = {
    name = "MyApp";
    @include "features.cfg"
};
```

## API Reference

### Parsing Functions

- `ParseFile(filename string) (*Config, error)` - Parse from file
- `ParseString(input string) (*Config, error)` - Parse from string
- `Parse(reader io.Reader) (*Config, error)` - Parse from io.Reader

### Lookup Methods

- `Lookup(path string) (*Value, error)` - Get raw value
- `LookupString(path string) (string, error)` - Get string value
- `LookupInt(path string) (int, error)` - Get integer value
- `LookupInt64(path string) (int64, error)` - Get 64-bit integer value
- `LookupFloat(path string) (float64, error)` - Get float value
- `LookupBool(path string) (bool, error)` - Get boolean value

### Working with Complex Types

```go
// Access array elements
serversVal, err := config.Lookup("servers")
if err == nil && serversVal.Type == libconfig.TypeArray {
    for i, server := range serversVal.ArrayVal {
        fmt.Printf("Server %d: %s\n", i, server.StrVal)
    }
}

// Access group members
dbVal, err := config.Lookup("database")
if err == nil && dbVal.Type == libconfig.TypeGroup {
    for key, value := range dbVal.GroupVal {
        fmt.Printf("DB config %s: %v\n", key, value)
    }
}

// Access list elements (mixed types)
listVal, err := config.Lookup("mixed_list")
if err == nil && listVal.Type == libconfig.TypeList {
    for i, item := range listVal.ListVal {
        fmt.Printf("Item %d (type %s): %v\n", i, item.Type, item)
    }
}
```

## Error Handling

The library provides detailed error messages with line and column information:

```go
config, err := libconfig.ParseString(`invalid syntax here`)
if err != nil {
    fmt.Printf("Parse error: %v\n", err)
    // Output: Parse error: expected identifier at line 1, column 15
}
```

### Static Error Types

The library defines static error types that can be checked with `errors.Is()`:

```go
import "errors"

_, err := config.LookupString("nonexistent.path")
if errors.Is(err, libconfig.ErrSettingNotFound) {
    fmt.Println("Setting not found")
}

_, err = config.LookupInt("string_value")
if errors.Is(err, libconfig.ErrNotInteger) {
    fmt.Println("Value is not an integer")
}
```

**Available error types:**
- `ErrCannotLookupInNonGroup` - Trying to lookup in non-group value
- `ErrSettingNotFound` - Setting path doesn't exist
- `ErrNotInteger` - Value is not an integer
- `ErrNotFloat` - Value is not a float
- `ErrNotBoolean` - Value is not a boolean
- `ErrNotString` - Value is not a string
- `ErrIntegerOutOfRange` - Integer value out of range for target type

## Value Types

The library supports the following value types:

- `TypeInt` - 32-bit integers
- `TypeInt64` - 64-bit integers
- `TypeFloat` - 64-bit floating point
- `TypeBool` - Boolean values
- `TypeString` - String values
- `TypeArray` - Homogeneous arrays
- `TypeGroup` - Objects/maps
- `TypeList` - Heterogeneous lists

## Examples

See the [examples](examples/) directory for complete working examples including:

- Basic usage with all data types
- Complex nested structures
- File parsing with includes
- Error handling patterns

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Specification

This library implements the [libconfig specification](https://hyperrealm.github.io/libconfig/). For detailed format documentation, refer to the official libconfig manual.

## Development

This project includes comprehensive development tooling:

### Testing

```bash
# Run tests
make test

# Run tests with race detection
make race

# Run benchmarks
make bench

# Generate test coverage report
make coverage

# Open coverage report in browser
make coverage-html
```

### Code Quality

```bash
# Run linting
make lint

# Format code
make fmt

# Run all checks (format + lint + test)
make check
```

### Project Structure

- Comprehensive test suite with >95% coverage
- Benchmark tests for performance monitoring
- Golangci-lint configuration for code quality
- Makefile for common development tasks
- Examples demonstrating all features
