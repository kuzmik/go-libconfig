package libconfig

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
)

// Predefined parser errors for better error handling and testing.
var (
	ErrUnexpectedToken            = errors.New("unexpected token")
	ErrExpectedToken              = errors.New("expected token")
	ErrIncludeDepthExceeded       = errors.New("include depth limit exceeded")
	ErrExpectedStringAfterInclude = errors.New("expected string after @include")
	ErrIncludeFileNotFound        = errors.New("include file not found")
	ErrExpectedIdentifier         = errors.New("expected identifier")
	ErrExpectedAssignment         = errors.New("expected assignment operator")
	ErrArrayTypeMismatch          = errors.New("array elements must have the same type")
)

// Parser parses libconfig tokens into a configuration.
type Parser struct {
	lexer        *Lexer
	baseDir      string // Directory of the main config file for resolving includes
	current      Token
	includeDepth int // Track include depth to prevent infinite recursion
}

// NewParser creates a new parser.
func NewParser(lexer *Lexer) *Parser {
	p := &Parser{
		lexer:        lexer,
		includeDepth: 0,
	}
	p.advance()

	return p
}

// NewParserWithBaseDir creates a new parser with a base directory for includes.
func NewParserWithBaseDir(lexer *Lexer, baseDir string) *Parser {
	p := &Parser{
		lexer:        lexer,
		baseDir:      baseDir,
		includeDepth: 0,
	}
	p.advance()

	return p
}

// advance moves to the next token.
func (p *Parser) advance() {
	p.current = p.lexer.NextToken()
}

// expect checks if the current token is of the expected type and advances.
func (p *Parser) expect(tokenType TokenType) error {
	if p.current.Type != tokenType {
		return fmt.Errorf("expected %s, got %s at line %d, column %d: %w",
			tokenType, p.current.Type, p.current.Line, p.current.Column, ErrExpectedToken)
	}

	p.advance()

	return nil
}

// Parse parses the configuration.
func (p *Parser) Parse() (*Config, error) {
	config := NewConfig()

	// Parse top-level settings
	for p.current.Type != TokenEOF {
		if p.current.Type == TokenInclude {
			// Handle @include directive
			if err := p.parseInclude(&config.Root); err != nil {
				return nil, err
			}

			continue
		}

		// Parse setting
		name, value, err := p.parseSetting()
		if err != nil {
			return nil, err
		}

		config.Root.GroupVal[name] = value

		// Optional semicolon
		if p.current.Type == TokenSemicolon {
			p.advance()
		}
	}

	return config, nil
}

// parseInclude handles @include directives by actually parsing and merging the included files.
func (p *Parser) parseInclude(target *Value) error {
	if p.includeDepth >= 10 {
		return fmt.Errorf("include depth limit exceeded (10) at line %d: %w", p.current.Line, ErrIncludeDepthExceeded)
	}

	p.advance() // consume @include

	if p.current.Type != TokenString {
		return fmt.Errorf("expected string after @include at line %d: %w", p.current.Line, ErrExpectedStringAfterInclude)
	}

	includePath := p.current.Value
	p.advance()

	// Optional semicolon after include
	if p.current.Type == TokenSemicolon {
		p.advance()
	}

	// Resolve the include path relative to the base directory
	var fullPath string
	if p.baseDir != "" {
		fullPath = filepath.Join(p.baseDir, includePath)
	} else {
		fullPath = includePath
	}

	// Try common extensions if the file doesn't exist as-is
	possiblePaths := []string{
		fullPath,
		fullPath + ".cnf",
		fullPath + ".cfg",
	}

	var existingPath string

	for _, path := range possiblePaths {
		if fileExists(path) {
			existingPath = path
			break
		}
	}

	if existingPath == "" {
		return fmt.Errorf("include file '%s' not found (tried: %v): %w", includePath, possiblePaths, ErrIncludeFileNotFound)
	}

	// Parse the included file
	includedConfig, err := parseFileWithDepth(existingPath, p.includeDepth+1)
	if err != nil {
		return fmt.Errorf("error parsing included file '%s': %w", existingPath, err)
	}

	// Merge the included configuration into the target
	mergeConfig(target, &includedConfig.Root)

	return nil
}

// parseSetting parses a name = value or name : value setting.
func (p *Parser) parseSetting() (string, Value, error) {
	if p.current.Type != TokenIdentifier {
		return "", Value{}, fmt.Errorf("expected identifier at line %d, column %d: %w",
			p.current.Line, p.current.Column, ErrExpectedIdentifier)
	}

	name := p.current.Value
	p.advance()

	if p.current.Type != TokenAssign {
		return "", Value{}, fmt.Errorf("expected assignment operator at line %d, column %d: %w",
			p.current.Line, p.current.Column, ErrExpectedAssignment)
	}

	p.advance()

	value, err := p.parseValue()
	if err != nil {
		return "", Value{}, err
	}

	return name, value, nil
}

// parseValue parses a value (scalar, array, group, or list).
func (p *Parser) parseValue() (Value, error) {
	switch p.current.Type {
	case TokenString:
		value := p.current.Value
		p.advance()

		// Handle string concatenation
		for p.current.Type == TokenString {
			value += p.current.Value
			p.advance()
		}

		return NewStringValue(value), nil

	case TokenInteger:
		val, err := parseIntegerLiteral(p.current.Value)
		if err != nil {
			return Value{}, fmt.Errorf("invalid integer at line %d: %w", p.current.Line, err)
		}

		p.advance()

		return val, nil

	case TokenFloat:
		val, err := strconv.ParseFloat(p.current.Value, 64)
		if err != nil {
			return Value{}, fmt.Errorf("invalid float at line %d: %w", p.current.Line, err)
		}

		p.advance()

		return NewFloatValue(val), nil

	case TokenBoolean:
		val := p.current.Value == "true"
		p.advance()

		return NewBoolValue(val), nil

	case TokenLeftBrace:
		return p.parseGroup()

	case TokenLeftBracket:
		return p.parseArray()

	case TokenLeftParen:
		return p.parseList()

	default:
		return Value{}, fmt.Errorf("unexpected token %s at line %d, column %d: %w",
			p.current.Type, p.current.Line, p.current.Column, ErrUnexpectedToken)
	}
}

// parseGroup parses a group { ... }.
func (p *Parser) parseGroup() (Value, error) {
	if err := p.expect(TokenLeftBrace); err != nil {
		return Value{}, err
	}

	group := make(map[string]Value)

	for p.current.Type != TokenRightBrace && p.current.Type != TokenEOF {
		if p.current.Type == TokenInclude {
			// Handle @include within groups
			groupValue := Value{Type: TypeGroup, GroupVal: group}
			if err := p.parseInclude(&groupValue); err != nil {
				return Value{}, err
			}

			group = groupValue.GroupVal

			continue
		}

		name, value, err := p.parseSetting()
		if err != nil {
			return Value{}, err
		}

		group[name] = value

		// Optional semicolon
		if p.current.Type == TokenSemicolon {
			p.advance()
		}
	}

	if err := p.expect(TokenRightBrace); err != nil {
		return Value{}, err
	}

	return NewGroupValue(group), nil
}

// parseArray parses an array [ ... ].
func (p *Parser) parseArray() (Value, error) {
	if err := p.expect(TokenLeftBracket); err != nil {
		return Value{}, err
	}

	var elements []Value

	// Empty array
	if p.current.Type == TokenRightBracket {
		p.advance()
		return NewArrayValue(elements), nil
	}

	// Parse first element
	firstElement, err := p.parseValue()
	if err != nil {
		return Value{}, err
	}

	elements = append(elements, firstElement)

	// Parse remaining elements
	for p.current.Type == TokenComma {
		p.advance() // consume comma

		// Allow trailing comma
		if p.current.Type == TokenRightBracket {
			break
		}

		element, err := p.parseValue()
		if err != nil {
			return Value{}, err
		}

		// Ensure all elements have the same type (arrays are homogeneous)
		if element.Type != firstElement.Type {
			return Value{}, fmt.Errorf("array elements must have the same type, got %s and %s at line %d: %w",
				firstElement.Type, element.Type, p.current.Line, ErrArrayTypeMismatch)
		}

		elements = append(elements, element)
	}

	if err := p.expect(TokenRightBracket); err != nil {
		return Value{}, err
	}

	return NewArrayValue(elements), nil
}

// parseList parses a list ( ... ).
func (p *Parser) parseList() (Value, error) {
	if err := p.expect(TokenLeftParen); err != nil {
		return Value{}, err
	}

	var elements []Value

	// Empty list
	if p.current.Type == TokenRightParen {
		p.advance()
		return NewListValue(elements), nil
	}

	// Parse first element
	element, err := p.parseValue()
	if err != nil {
		return Value{}, err
	}

	elements = append(elements, element)

	// Parse remaining elements
	for p.current.Type == TokenComma {
		p.advance() // consume comma

		// Allow trailing comma
		if p.current.Type == TokenRightParen {
			break
		}

		element, err := p.parseValue()
		if err != nil {
			return Value{}, err
		}

		elements = append(elements, element)
	}

	if err := p.expect(TokenRightParen); err != nil {
		return Value{}, err
	}

	return NewListValue(elements), nil
}

// Helper functions

// fileExists checks if a file exists.
func fileExists(path string) bool {
	if _, err := os.Stat(path); err == nil {
		return true
	}

	return false
}

// parseFileWithDepth parses a file with include depth tracking.
func parseFileWithDepth(filename string, depth int) (*Config, error) {
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
	parser.includeDepth = depth

	return parser.Parse()
}

// mergeConfig merges source config into target config.
func mergeConfig(target, source *Value) {
	if target.Type != TypeGroup || source.Type != TypeGroup {
		return
	}

	if target.GroupVal == nil {
		target.GroupVal = make(map[string]Value)
	}

	for key, value := range source.GroupVal {
		target.GroupVal[key] = value
	}
}
