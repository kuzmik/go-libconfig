// Package libconfig provides a parser for libconfig configuration files.
// It supports the full libconfig specification including scalars, arrays, groups, lists,
// and various integer formats (decimal, hex, octal, binary).
package libconfig

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// ValueType represents the type of a configuration value.
type ValueType int

const (
	TypeInt ValueType = iota
	TypeInt64
	TypeFloat
	TypeBool
	TypeString
	TypeArray
	TypeGroup
	TypeList
)

// String returns the string representation of the value type.
func (vt ValueType) String() string {
	switch vt {
	case TypeInt:
		return "int"
	case TypeInt64:
		return "int64"
	case TypeFloat:
		return "float"
	case TypeBool:
		return "bool"
	case TypeString:
		return "string"
	case TypeArray:
		return "array"
	case TypeGroup:
		return "group"
	case TypeList:
		return "list"
	default:
		return "unknown"
	}
}

// Value represents a configuration value.
type Value struct {
	ArrayVal []Value
	ListVal  []Value
	StrVal   string
	GroupVal map[string]Value
	IntVal   int
	Int64Val int64
	FloatVal float64
	Type     ValueType
	BoolVal  bool
}

// Config represents a libconfig configuration.
type Config struct {
	Root Value
}

// NewConfig creates a new empty configuration.
func NewConfig() *Config {
	return &Config{
		Root: Value{
			Type:     TypeGroup,
			GroupVal: make(map[string]Value),
		},
	}
}

// ParseFile parses a libconfig file.
func ParseFile(filename string) (*Config, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}

	defer func() {
		file.Close() // Ignore close errors after successful read
	}()

	lexer := NewLexer(file)
	baseDir := filepath.Dir(filename)
	parser := NewParserWithBaseDir(lexer, baseDir)

	return parser.Parse()
}

// ParseString parses a libconfig string.
func ParseString(input string) (*Config, error) {
	return Parse(strings.NewReader(input))
}

// Parse parses libconfig data from a reader.
func Parse(reader io.Reader) (*Config, error) {
	lexer := NewLexer(reader)
	parser := NewParser(lexer)

	return parser.Parse()
}

// Lookup finds a setting by path (dot-separated).
func (c *Config) Lookup(path string) (*Value, error) {
	parts := strings.Split(path, ".")
	current := &c.Root

	for _, part := range parts {
		if part == "" {
			continue
		}

		if current.Type != TypeGroup {
			return nil, fmt.Errorf("cannot lookup '%s': %w", part, ErrCannotLookupInNonGroup)
		}

		val, exists := current.GroupVal[part]
		if !exists {
			return nil, fmt.Errorf("setting '%s': %w", part, ErrSettingNotFound)
		}

		current = &val
	}

	return current, nil
}

// LookupInt looks up an integer value by path.
func (c *Config) LookupInt(path string) (int, error) {
	val, err := c.Lookup(path)
	if err != nil {
		return 0, err
	}

	switch val.Type {
	case TypeInt:
		return val.IntVal, nil
	case TypeInt64:
		if val.Int64Val > int64(^uint(0)>>1) || val.Int64Val < int64(-1<<(64-1)) {
			return 0, fmt.Errorf("int64 value %d: %w", val.Int64Val, ErrIntegerOutOfRange)
		}

		return int(val.Int64Val), nil
	default:
		return 0, fmt.Errorf("value at '%s': %w", path, ErrNotInteger)
	}
}

// LookupInt64 looks up a 64-bit integer value by path.
func (c *Config) LookupInt64(path string) (int64, error) {
	val, err := c.Lookup(path)
	if err != nil {
		return 0, err
	}

	switch val.Type {
	case TypeInt:
		return int64(val.IntVal), nil
	case TypeInt64:
		return val.Int64Val, nil
	default:
		return 0, fmt.Errorf("value at '%s': %w", path, ErrNotInteger)
	}
}

// LookupFloat looks up a float value by path.
func (c *Config) LookupFloat(path string) (float64, error) {
	val, err := c.Lookup(path)
	if err != nil {
		return 0, err
	}

	if val.Type != TypeFloat {
		return 0, fmt.Errorf("value at '%s': %w", path, ErrNotFloat)
	}

	return val.FloatVal, nil
}

// LookupBool looks up a boolean value by path.
func (c *Config) LookupBool(path string) (bool, error) {
	val, err := c.Lookup(path)
	if err != nil {
		return false, err
	}

	if val.Type != TypeBool {
		return false, fmt.Errorf("value at '%s': %w", path, ErrNotBoolean)
	}

	return val.BoolVal, nil
}

// LookupString looks up a string value by path.
func (c *Config) LookupString(path string) (string, error) {
	val, err := c.Lookup(path)
	if err != nil {
		return "", err
	}

	if val.Type != TypeString {
		return "", fmt.Errorf("value at '%s': %w", path, ErrNotString)
	}

	return val.StrVal, nil
}

// Helper functions for creating values

// NewIntValue creates a new integer value.
func NewIntValue(val int) Value {
	return Value{Type: TypeInt, IntVal: val}
}

// NewInt64Value creates a new 64-bit integer value.
func NewInt64Value(val int64) Value {
	return Value{Type: TypeInt64, Int64Val: val}
}

// NewFloatValue creates a new float value.
func NewFloatValue(val float64) Value {
	return Value{Type: TypeFloat, FloatVal: val}
}

// NewBoolValue creates a new boolean value.
func NewBoolValue(val bool) Value {
	return Value{Type: TypeBool, BoolVal: val}
}

// NewStringValue creates a new string value.
func NewStringValue(val string) Value {
	return Value{Type: TypeString, StrVal: val}
}

// NewArrayValue creates a new array value.
func NewArrayValue(vals []Value) Value {
	return Value{Type: TypeArray, ArrayVal: vals}
}

// NewGroupValue creates a new group value.
func NewGroupValue(vals map[string]Value) Value {
	return Value{Type: TypeGroup, GroupVal: vals}
}

// NewListValue creates a new list value.
func NewListValue(vals []Value) Value {
	return Value{Type: TypeList, ListVal: vals}
}

// parseIntegerLiteral parses integer literals in various formats.
func parseIntegerLiteral(s string) (Value, error) {
	s = strings.TrimSpace(s)

	isLong := strings.HasSuffix(s, "L") || strings.HasSuffix(s, "l")
	if isLong {
		s = s[:len(s)-1]
	}

	var (
		val int64
		err error
	)

	switch {
	case strings.HasPrefix(s, "0x") || strings.HasPrefix(s, "0X"):
		// Hexadecimal
		val, err = strconv.ParseInt(s[2:], 16, 64)
	case strings.HasPrefix(s, "0b") || strings.HasPrefix(s, "0B"):
		// Binary
		val, err = strconv.ParseInt(s[2:], 2, 64)
	case strings.HasPrefix(s, "0o") || strings.HasPrefix(s, "0O") || strings.HasPrefix(s, "0q") || strings.HasPrefix(s, "0Q"):
		// Octal (new format)
		val, err = strconv.ParseInt(s[2:], 8, 64)
	default:
		// Decimal
		val, err = strconv.ParseInt(s, 10, 64)
	}

	if err != nil {
		return Value{}, fmt.Errorf("invalid integer literal '%s': %w", s, err)
	}

	// Determine if we should return 32-bit or 64-bit based on value and suffix
	if isLong || val > int64(^uint(0)>>1) || val < int64(-1<<(64-1)) {
		return NewInt64Value(val), nil
	}

	return NewIntValue(int(val)), nil
}

// Predefined errors for better error handling and testing.
var (
	ErrCannotLookupInNonGroup = errors.New("cannot lookup in non-group value")
	ErrSettingNotFound        = errors.New("setting not found")
	ErrNotInteger             = errors.New("value is not an integer")
	ErrNotFloat               = errors.New("value is not a float")
	ErrNotBoolean             = errors.New("value is not a boolean")
	ErrNotString              = errors.New("value is not a string")
	ErrIntegerOutOfRange      = errors.New("integer value out of range")
)
