package libconfig

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// errorReader is a custom reader that always returns an error
type errorReader struct{}

func (r *errorReader) Read(p []byte) (n int, err error) {
	return 0, fmt.Errorf("simulated read error")
}

// TestNewLexerWithIOError tests the error handling path in NewLexer
// when io.Copy fails, ensuring it creates an empty lexer gracefully.
func TestNewLexerWithIOError(t *testing.T) {
	// Create a reader that will cause io.Copy to fail
	errorReader := &errorReader{}

	// Call NewLexer with the error reader
	lexer := NewLexer(errorReader)

	// Verify the lexer is in the expected empty state
	if lexer.input != "" {
		t.Errorf("Expected empty input, got %q", lexer.input)
	}

	if lexer.pos != 0 {
		t.Errorf("Expected pos=0, got %d", lexer.pos)
	}

	if lexer.line != 1 {
		t.Errorf("Expected line=1, got %d", lexer.line)
	}

	if lexer.column != 1 {
		t.Errorf("Expected column=1, got %d", lexer.column)
	}

	// Verify it has exactly one EOF token
	if len(lexer.tokens) != 1 {
		t.Errorf("Expected 1 token, got %d", len(lexer.tokens))
	}

	if len(lexer.tokens) > 0 {
		token := lexer.tokens[0]
		if token.Type != TokenEOF {
			t.Errorf("Expected EOF token, got %s", token.Type)
		}
		if token.Value != "" {
			t.Errorf("Expected empty token value, got %q", token.Value)
		}
		if token.Line != 1 {
			t.Errorf("Expected token line=1, got %d", token.Line)
		}
		if token.Column != 1 {
			t.Errorf("Expected token column=1, got %d", token.Column)
		}
	}

	// Test that NextToken() works correctly with the error lexer
	token := lexer.NextToken()
	if token.Type != TokenEOF {
		t.Errorf("Expected NextToken to return EOF, got %s", token.Type)
	}

	// Test that PeekToken() works correctly
	peekedToken := lexer.PeekToken()
	if peekedToken.Type != TokenEOF {
		t.Errorf("Expected PeekToken to return EOF, got %s", peekedToken.Type)
	}
}

// TestTokenString tests the Token.String() method
func TestTokenString(t *testing.T) {
	tests := []struct {
		token    Token
		expected string
	}{
		{
			Token{Value: "test", Type: TokenString, Line: 1, Column: 5},
			"{STRING: \"test\" at 1:5}",
		},
		{
			Token{Value: "42", Type: TokenInteger, Line: 2, Column: 10},
			"{INTEGER: \"42\" at 2:10}",
		},
		{
			Token{Value: "=", Type: TokenAssign, Line: 3, Column: 1},
			"{ASSIGN: \"=\" at 3:1}",
		},
	}

	for i, tt := range tests {
		t.Run(fmt.Sprintf("token_%d", i), func(t *testing.T) {
			result := tt.token.String()
			if result != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result)
			}
		})
	}
}

// TestTokenTypeString tests all TokenType.String() cases
func TestTokenTypeString(t *testing.T) {
	tests := []struct {
		tokenType TokenType
		expected  string
	}{
		{TokenEOF, "EOF"},
		{TokenIdentifier, "IDENTIFIER"},
		{TokenString, "STRING"},
		{TokenInteger, "INTEGER"},
		{TokenFloat, "FLOAT"},
		{TokenBoolean, "BOOLEAN"},
		{TokenAssign, "ASSIGN"},
		{TokenSemicolon, "SEMICOLON"},
		{TokenComma, "COMMA"},
		{TokenLeftBrace, "LEFT_BRACE"},
		{TokenRightBrace, "RIGHT_BRACE"},
		{TokenLeftBracket, "LEFT_BRACKET"},
		{TokenRightBracket, "RIGHT_BRACKET"},
		{TokenLeftParen, "LEFT_PAREN"},
		{TokenRightParen, "RIGHT_PAREN"},
		{TokenInclude, "INCLUDE"},
		{TokenError, "ERROR"},
		{TokenType(999), "UNKNOWN"}, // Test unknown type
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := tt.tokenType.String()
			if result != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result)
			}
		})
	}
}

// TestLookupIntErrors tests error cases in LookupInt function
func TestLookupIntErrors(t *testing.T) {
	config, err := ParseString(`
		max_int64 = 9223372036854775807L;
		min_int64 = -9223372036854775808L;
		float_val = 3.14;
		string_val = "hello";
	`)
	if err != nil {
		t.Fatalf("Failed to parse config: %v", err)
	}

	// Test int64 out of range (for 32-bit systems, this may not trigger)
	_, err = config.LookupInt("max_int64")
	// On 64-bit systems, this should work fine, but we're testing the error path exists

	// Test wrong type errors
	_, err = config.LookupInt("float_val")
	if err == nil {
		t.Error("Expected error when looking up float as int")
	}

	_, err = config.LookupInt("string_val")
	if err == nil {
		t.Error("Expected error when looking up string as int")
	}

	// Test non-existent path
	_, err = config.LookupInt("nonexistent")
	if err == nil {
		t.Error("Expected error for non-existent path")
	}
}

// TestParseFile tests the ParseFile function
func TestParseFile(t *testing.T) {
	// Create a temporary config file
	tmpFile, err := os.CreateTemp("", "test_config_*.cfg")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	// Write test configuration
	configContent := `
		app_name = "TestApp";
		version = "1.0.0";
		port = 8080;
		debug = true;
	`
	if _, err := tmpFile.WriteString(configContent); err != nil {
		t.Fatalf("Failed to write to temp file: %v", err)
	}
	tmpFile.Close()

	// Test ParseFile
	config, err := ParseFile(tmpFile.Name())
	if err != nil {
		t.Fatalf("Failed to parse file: %v", err)
	}

	// Verify parsed content
	appName, err := config.LookupString("app_name")
	if err != nil || appName != "TestApp" {
		t.Errorf("Expected app_name='TestApp', got '%s'", appName)
	}

	port, err := config.LookupInt("port")
	if err != nil || port != 8080 {
		t.Errorf("Expected port=8080, got %d", port)
	}

	// Test file not found error
	_, err = ParseFile("nonexistent_file.cfg")
	if err == nil {
		t.Error("Expected error for non-existent file")
	}
}

// TestNewParserWithBaseDir tests the NewParserWithBaseDir function
func TestNewParserWithBaseDir(t *testing.T) {
	lexer := NewLexer(strings.NewReader("test = 42;"))
	baseDir := "/test/base/dir"

	parser := NewParserWithBaseDir(lexer, baseDir)

	if parser.baseDir != baseDir {
		t.Errorf("Expected baseDir='%s', got '%s'", baseDir, parser.baseDir)
	}

	if parser.includeDepth != 0 {
		t.Errorf("Expected includeDepth=0, got %d", parser.includeDepth)
	}

	if parser.lexer != lexer {
		t.Error("Expected lexer to be set correctly")
	}
}

// TestNextTokenEOF tests NextToken behavior at EOF
func TestNextTokenEOF(t *testing.T) {
	lexer := NewLexer(strings.NewReader(""))

	// First call should return EOF
	token1 := lexer.NextToken()
	if token1.Type != TokenEOF {
		t.Errorf("Expected EOF, got %s", token1.Type)
	}

	// Subsequent calls should also return EOF
	token2 := lexer.NextToken()
	if token2.Type != TokenEOF {
		t.Errorf("Expected EOF on second call, got %s", token2.Type)
	}
}

// TestPeekTokenEOF tests PeekToken behavior at EOF
func TestPeekTokenEOF(t *testing.T) {
	lexer := NewLexer(strings.NewReader(""))

	// Peek should return EOF
	peeked := lexer.PeekToken()
	if peeked.Type != TokenEOF {
		t.Errorf("Expected EOF from peek, got %s", peeked.Type)
	}

	// NextToken should still return EOF
	next := lexer.NextToken()
	if next.Type != TokenEOF {
		t.Errorf("Expected EOF from next, got %s", next.Type)
	}
}

// TestIncludeFileHandling tests include file functionality with temporary files
func TestIncludeFileHandling(t *testing.T) {
	// Create temporary directory
	tmpDir, err := os.MkdirTemp("", "libconfig_test_")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create included file
	includedFile := filepath.Join(tmpDir, "included.cfg")
	includedContent := `
		included_setting = "from_include";
		included_port = 9090;
	`
	if err := os.WriteFile(includedFile, []byte(includedContent), 0o644); err != nil {
		t.Fatalf("Failed to write included file: %v", err)
	}

	// Create main file with include
	mainFile := filepath.Join(tmpDir, "main.cfg")
	mainContent := fmt.Sprintf(`
		main_setting = "from_main";
		@include "%s"
		main_port = 8080;
	`, "included.cfg") // Use relative path
	if err := os.WriteFile(mainFile, []byte(mainContent), 0o644); err != nil {
		t.Fatalf("Failed to write main file: %v", err)
	}

	// Test parsing with includes
	config, err := ParseFile(mainFile)
	if err != nil {
		t.Fatalf("Failed to parse file with includes: %v", err)
	}

	// Verify main settings
	mainSetting, err := config.LookupString("main_setting")
	if err != nil || mainSetting != "from_main" {
		t.Errorf("Expected main_setting='from_main', got '%s'", mainSetting)
	}

	// Verify included settings
	includedSetting, err := config.LookupString("included_setting")
	if err != nil || includedSetting != "from_include" {
		t.Errorf("Expected included_setting='from_include', got '%s'", includedSetting)
	}

	includedPort, err := config.LookupInt("included_port")
	if err != nil || includedPort != 9090 {
		t.Errorf("Expected included_port=9090, got %d", includedPort)
	}
}

// TestIncludeDepthLimit tests include depth limiting
func TestIncludeDepthLimit(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "libconfig_depth_test_")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a chain of include files that exceeds depth limit
	for i := 0; i <= 11; i++ {
		filename := filepath.Join(tmpDir, fmt.Sprintf("level%d.cfg", i))
		var content string
		if i < 11 {
			content = fmt.Sprintf(`
				level%d_setting = %d;
				@include "level%d.cfg"
			`, i, i, i+1)
		} else {
			content = fmt.Sprintf(`level%d_setting = %d;`, i, i)
		}

		if err := os.WriteFile(filename, []byte(content), 0o644); err != nil {
			t.Fatalf("Failed to write level%d file: %v", i, err)
		}
	}

	// Try to parse - should fail due to depth limit
	mainFile := filepath.Join(tmpDir, "level0.cfg")
	_, err = ParseFile(mainFile)
	if err == nil {
		t.Error("Expected error due to include depth limit, but parsing succeeded")
	}

	// Verify error mentions depth limit
	if !strings.Contains(err.Error(), "depth limit") {
		t.Errorf("Expected depth limit error, got: %v", err)
	}
}

// TestLexerPeekEdgeCases tests edge cases in lexer peek function
func TestLexerPeekEdgeCases(t *testing.T) {
	// Test peek at end of input
	lexer := NewLexer(strings.NewReader("a"))

	// Advance to 'a'
	lexer.advance()

	// Peek should return 0 (EOF)
	peeked := lexer.peek()
	if peeked != 0 {
		t.Errorf("Expected peek to return 0 at EOF, got %q", peeked)
	}
}

func TestParseString(t *testing.T) {
	config, err := ParseString(`name = "test"; port = 8080;`)
	if err != nil {
		t.Fatalf("Failed to parse: %v", err)
	}

	name, err := config.LookupString("name")
	if err != nil || name != "test" {
		t.Errorf("Expected name='test', got '%s'", name)
	}

	port, err := config.LookupInt("port")
	if err != nil || port != 8080 {
		t.Errorf("Expected port=8080, got %d", port)
	}
}

// Test basic integer values.
func TestParseIntegers(t *testing.T) {
	tests := []struct {
		name     string
		config   string
		key      string
		expected int
	}{
		{"positive_decimal", `value = 42;`, "value", 42},
		{"negative_decimal", `value = -123;`, "value", -123},
		{"zero", `value = 0;`, "value", 0},
		{"max_int", `value = 2147483647;`, "value", 2147483647},
		{"min_int", `value = -2147483648;`, "value", -2147483648},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config, err := ParseString(tt.config)
			if err != nil {
				t.Fatalf("Failed to parse config: %v", err)
			}

			value, err := config.LookupInt(tt.key)
			if err != nil {
				t.Fatalf("Failed to lookup int: %v", err)
			}

			if value != tt.expected {
				t.Errorf("Expected %d, got %d", tt.expected, value)
			}
		})
	}
}

// Test 64-bit integer values.
func TestParseInt64(t *testing.T) {
	tests := []struct {
		name     string
		config   string
		key      string
		expected int64
	}{
		{"small_int64", `value = 1234L;`, "value", 1234},
		{"large_int64", `value = 9223372036854775807L;`, "value", 9223372036854775807},
		{"negative_int64", `value = -9223372036854775808L;`, "value", -9223372036854775808},
		{"without_suffix", `value = 1234567890123;`, "value", 1234567890123},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config, err := ParseString(tt.config)
			if err != nil {
				t.Fatalf("Failed to parse config: %v", err)
			}

			value, err := config.LookupInt64(tt.key)
			if err != nil {
				t.Fatalf("Failed to lookup int64: %v", err)
			}

			if value != tt.expected {
				t.Errorf("Expected %d, got %d", tt.expected, value)
			}
		})
	}
}

// Test float values.
func TestParseFloats(t *testing.T) {
	tests := []struct {
		name     string
		config   string
		key      string
		expected float64
	}{
		{"simple_float", `value = 3.14;`, "value", 3.14},
		{"negative_float", `value = -2.5;`, "value", -2.5},
		{"scientific_notation", `value = 1.23e4;`, "value", 12300.0},
		{"scientific_negative_exp", `value = 1.5e-3;`, "value", 0.0015},
		{"scientific_capital_e", `value = 2.5E+2;`, "value", 250.0},
		{"zero_float", `value = 0.0;`, "value", 0.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config, err := ParseString(tt.config)
			if err != nil {
				t.Fatalf("Failed to parse config: %v", err)
			}

			value, err := config.LookupFloat(tt.key)
			if err != nil {
				t.Fatalf("Failed to lookup float: %v", err)
			}

			if value != tt.expected {
				t.Errorf("Expected %f, got %f", tt.expected, value)
			}
		})
	}
}

// Test boolean values.
func TestParseBooleans(t *testing.T) {
	tests := []struct {
		name     string
		config   string
		key      string
		expected bool
	}{
		{"true_lowercase", `value = true;`, "value", true},
		{"false_lowercase", `value = false;`, "value", false},
		{"true_uppercase", `value = TRUE;`, "value", true},
		{"false_uppercase", `value = FALSE;`, "value", false},
		{"true_mixed", `value = True;`, "value", true},
		{"false_mixed", `value = False;`, "value", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config, err := ParseString(tt.config)
			if err != nil {
				t.Fatalf("Failed to parse config: %v", err)
			}

			value, err := config.LookupBool(tt.key)
			if err != nil {
				t.Fatalf("Failed to lookup bool: %v", err)
			}

			if value != tt.expected {
				t.Errorf("Expected %t, got %t", tt.expected, value)
			}
		})
	}
}

// Test string values.
func TestParseStrings(t *testing.T) {
	tests := []struct {
		name     string
		config   string
		key      string
		expected string
	}{
		{"simple_string", `value = "hello";`, "value", "hello"},
		{"empty_string", `value = "";`, "value", ""},
		{"string_with_spaces", `value = "hello world";`, "value", "hello world"},
		{"string_with_numbers", `value = "test123";`, "value", "test123"},
		{"string_with_special_chars", `value = "test@#$%^&*()";`, "value", "test@#$%^&*()"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config, err := ParseString(tt.config)
			if err != nil {
				t.Fatalf("Failed to parse config: %v", err)
			}

			value, err := config.LookupString(tt.key)
			if err != nil {
				t.Fatalf("Failed to lookup string: %v", err)
			}

			if value != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, value)
			}
		})
	}
}

// Test type checking.
func TestTypeChecking(t *testing.T) {
	config, err := ParseString(`
		str_val = "hello";
		int_val = 42;
		float_val = 3.14;
		bool_val = true;
	`)
	if err != nil {
		t.Fatalf("Failed to parse config: %v", err)
	}

	// Test wrong type lookups should return errors
	_, err = config.LookupString("int_val")
	if err == nil {
		t.Error("Expected error when looking up int as string")
	}

	_, err = config.LookupInt("str_val")
	if err == nil {
		t.Error("Expected error when looking up string as int")
	}

	_, err = config.LookupFloat("bool_val")
	if err == nil {
		t.Error("Expected error when looking up bool as float")
	}

	_, err = config.LookupBool("float_val")
	if err == nil {
		t.Error("Expected error when looking up float as bool")
	}
}

// Test path lookups.
func TestPathLookup(t *testing.T) {
	config, err := ParseString(`
		database = {
			host = "localhost";
			port = 5432;
			credentials = {
				username = "admin";
				password = "secret";
			};
		};
	`)
	if err != nil {
		t.Fatalf("Failed to parse config: %v", err)
	}

	// Test nested lookups
	host, err := config.LookupString("database.host")
	if err != nil || host != "localhost" {
		t.Errorf("Expected 'localhost', got '%s'", host)
	}

	port, err := config.LookupInt("database.port")
	if err != nil || port != 5432 {
		t.Errorf("Expected 5432, got %d", port)
	}

	username, err := config.LookupString("database.credentials.username")
	if err != nil || username != "admin" {
		t.Errorf("Expected 'admin', got '%s'", username)
	}

	// Test non-existent path
	_, err = config.LookupString("database.nonexistent")
	if err == nil {
		t.Error("Expected error for non-existent path")
	}

	// Test invalid path (trying to access non-group as group)
	_, err = config.LookupString("database.host.invalid")
	if err == nil {
		t.Error("Expected error for invalid path")
	}
}

// Test assignment operators.
func TestAssignmentOperators(t *testing.T) {
	tests := []struct {
		name   string
		config string
	}{
		{"equals", `name = "test";`},
		{"colon", `name : "test";`},
		{"mixed", `name = "test"; port : 8080;`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config, err := ParseString(tt.config)
			if err != nil {
				t.Fatalf("Failed to parse config with %s: %v", tt.name, err)
			}

			name, err := config.LookupString("name")
			if err != nil || name != "test" {
				t.Errorf("Expected 'test', got '%s'", name)
			}
		})
	}
}

// Test semicolons.
func TestSemicolons(t *testing.T) {
	tests := []struct {
		name   string
		config string
	}{
		{"with_semicolons", `name = "test"; port = 8080;`},
		{"without_semicolons", `name = "test" port = 8080`},
		{"mixed", `name = "test"; port = 8080`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config, err := ParseString(tt.config)
			if err != nil {
				t.Fatalf("Failed to parse config: %v", err)
			}

			name, err := config.LookupString("name")
			if err != nil || name != "test" {
				t.Errorf("Expected 'test', got '%s'", name)
			}

			port, err := config.LookupInt("port")
			if err != nil || port != 8080 {
				t.Errorf("Expected 8080, got %d", port)
			}
		})
	}
}

// Test comments.
func TestComments(t *testing.T) {
	configStr := `
		// C++ style comment
		name = "test"; // inline comment

		/* C style comment */
		port = 8080; /* inline C comment */

		# Script style comment
		debug = true; # inline script comment

		/*
		 * Multi-line
		 * C style comment
		 */
		host = "localhost";
	`

	config, err := ParseString(configStr)
	if err != nil {
		t.Fatalf("Failed to parse config with comments: %v", err)
	}

	name, err := config.LookupString("name")
	if err != nil || name != "test" {
		t.Errorf("Expected 'test', got '%s'", name)
	}

	port, err := config.LookupInt("port")
	if err != nil || port != 8080 {
		t.Errorf("Expected 8080, got %d", port)
	}

	debug, err := config.LookupBool("debug")
	if err != nil || debug != true {
		t.Errorf("Expected true, got %t", debug)
	}

	host, err := config.LookupString("host")
	if err != nil || host != "localhost" {
		t.Errorf("Expected 'localhost', got '%s'", host)
	}
}

// Test empty configurations.
func TestEmptyConfig(t *testing.T) {
	tests := []string{
		"",
		"   ",
		"\n\n\n",
		"// just comments",
		"/* just comments */",
		"# just comments",
	}

	for _, configStr := range tests {
		config, err := ParseString(configStr)
		if err != nil {
			t.Errorf("Failed to parse empty config: %v", err)
			continue
		}

		if config == nil {
			t.Error("Config should not be nil")
			continue
		}

		if config.Root.Type != TypeGroup {
			t.Error("Root should be a group")
		}

		if len(config.Root.GroupVal) != 0 {
			t.Error("Root group should be empty")
		}
	}
}

// Test Parse function with reader.
func TestParseReader(t *testing.T) {
	configStr := `name = "test"; port = 8080;`
	reader := strings.NewReader(configStr)

	config, err := Parse(reader)
	if err != nil {
		t.Fatalf("Failed to parse from reader: %v", err)
	}

	name, err := config.LookupString("name")
	if err != nil || name != "test" {
		t.Errorf("Expected 'test', got '%s'", name)
	}
}

// Test arrays.
func TestParseArrays(t *testing.T) {
	tests := []struct {
		name        string
		config      string
		key         string
		expectedLen int
		valueType   ValueType
	}{
		{"string_array", `values = [ "a", "b", "c" ];`, "values", 3, TypeString},
		{"int_array", `values = [ 1, 2, 3, 4, 5 ];`, "values", 5, TypeInt},
		{"float_array", `values = [ 1.1, 2.2, 3.3 ];`, "values", 3, TypeFloat},
		{"bool_array", `values = [ true, false, true ];`, "values", 3, TypeBool},
		{"empty_array", `values = [ ];`, "values", 0, TypeString}, // Empty arrays default to string type
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config, err := ParseString(tt.config)
			if err != nil {
				t.Fatalf("Failed to parse config: %v", err)
			}

			value, err := config.Lookup(tt.key)
			if err != nil {
				t.Fatalf("Failed to lookup array: %v", err)
			}

			if value.Type != TypeArray {
				t.Errorf("Expected array type, got %s", value.Type)
			}

			if len(value.ArrayVal) != tt.expectedLen {
				t.Errorf("Expected array length %d, got %d", tt.expectedLen, len(value.ArrayVal))
			}

			// Check element types (if not empty)
			if len(value.ArrayVal) > 0 && value.ArrayVal[0].Type != tt.valueType {
				t.Errorf("Expected element type %s, got %s", tt.valueType, value.ArrayVal[0].Type)
			}
		})
	}
}

// Test array values.
func TestArrayValues(t *testing.T) {
	config, err := ParseString(`
		strings = [ "hello", "world", "test" ];
		numbers = [ 10, 20, 30 ];
		floats = [ 1.1, 2.2, 3.3 ];
		bools = [ true, false, true ];
	`)
	if err != nil {
		t.Fatalf("Failed to parse config: %v", err)
	}

	// Test string array
	strArray, err := config.Lookup("strings")
	if err != nil {
		t.Fatalf("Failed to lookup strings: %v", err)
	}

	expectedStrings := []string{"hello", "world", "test"}
	for i, str := range expectedStrings {
		if strArray.ArrayVal[i].StrVal != str {
			t.Errorf("Expected string[%d]='%s', got '%s'", i, str, strArray.ArrayVal[i].StrVal)
		}
	}

	// Test int array
	intArray, err := config.Lookup("numbers")
	if err != nil {
		t.Fatalf("Failed to lookup numbers: %v", err)
	}

	expectedInts := []int{10, 20, 30}
	for i, num := range expectedInts {
		if intArray.ArrayVal[i].IntVal != num {
			t.Errorf("Expected number[%d]=%d, got %d", i, num, intArray.ArrayVal[i].IntVal)
		}
	}

	// Test float array
	floatArray, err := config.Lookup("floats")
	if err != nil {
		t.Fatalf("Failed to lookup floats: %v", err)
	}

	expectedFloats := []float64{1.1, 2.2, 3.3}
	for i, num := range expectedFloats {
		if floatArray.ArrayVal[i].FloatVal != num {
			t.Errorf("Expected float[%d]=%f, got %f", i, num, floatArray.ArrayVal[i].FloatVal)
		}
	}

	// Test bool array
	boolArray, err := config.Lookup("bools")
	if err != nil {
		t.Fatalf("Failed to lookup bools: %v", err)
	}

	expectedBools := []bool{true, false, true}
	for i, b := range expectedBools {
		if boolArray.ArrayVal[i].BoolVal != b {
			t.Errorf("Expected bool[%d]=%t, got %t", i, b, boolArray.ArrayVal[i].BoolVal)
		}
	}
}

// Test array with trailing comma.
func TestArrayTrailingComma(t *testing.T) {
	config, err := ParseString(`values = [ "a", "b", "c", ];`)
	if err != nil {
		t.Fatalf("Failed to parse array with trailing comma: %v", err)
	}

	value, err := config.Lookup("values")
	if err != nil {
		t.Fatalf("Failed to lookup array: %v", err)
	}

	if len(value.ArrayVal) != 3 {
		t.Errorf("Expected array length 3, got %d", len(value.ArrayVal))
	}
}

// Test groups.
func TestParseGroups(t *testing.T) {
	config, err := ParseString(`
		database = {
			host = "localhost";
			port = 5432;
			name = "mydb";
		};

		server = {
			host = "webserver";
			port = 8080;
			ssl = true;
		};

		empty = { };
	`)
	if err != nil {
		t.Fatalf("Failed to parse config: %v", err)
	}

	// Test database group
	dbValue, err := config.Lookup("database")
	if err != nil {
		t.Fatalf("Failed to lookup database: %v", err)
	}

	if dbValue.Type != TypeGroup {
		t.Errorf("Expected group type, got %s", dbValue.Type)
	}

	if len(dbValue.GroupVal) != 3 {
		t.Errorf("Expected 3 group members, got %d", len(dbValue.GroupVal))
	}

	// Test nested lookups
	host, err := config.LookupString("database.host")
	if err != nil || host != "localhost" {
		t.Errorf("Expected 'localhost', got '%s'", host)
	}

	port, err := config.LookupInt("database.port")
	if err != nil || port != 5432 {
		t.Errorf("Expected 5432, got %d", port)
	}

	// Test empty group
	emptyValue, err := config.Lookup("empty")
	if err != nil {
		t.Fatalf("Failed to lookup empty group: %v", err)
	}

	if len(emptyValue.GroupVal) != 0 {
		t.Errorf("Expected empty group, got %d members", len(emptyValue.GroupVal))
	}
}

// Test nested groups.
func TestNestedGroups(t *testing.T) {
	config, err := ParseString(`
		app = {
			name = "MyApp";
			database = {
				host = "localhost";
				credentials = {
					username = "admin";
					password = "secret";
					settings = {
						timeout = 30;
						retries = 3;
					};
				};
			};
		};
	`)
	if err != nil {
		t.Fatalf("Failed to parse nested groups: %v", err)
	}

	// Test deeply nested lookups
	username, err := config.LookupString("app.database.credentials.username")
	if err != nil || username != "admin" {
		t.Errorf("Expected 'admin', got '%s'", username)
	}

	timeout, err := config.LookupInt("app.database.credentials.settings.timeout")
	if err != nil || timeout != 30 {
		t.Errorf("Expected 30, got %d", timeout)
	}
}

// Test lists (heterogeneous).
func TestParseLists(t *testing.T) {
	config, err := ParseString(`
		mixed = ( "string", 42, true, 3.14 );
		nested = (
			( "inner", "list" ),
			{ name = "object"; value = 123; },
			[ 1, 2, 3 ]
		);
		empty = ( );
	`)
	if err != nil {
		t.Fatalf("Failed to parse lists: %v", err)
	}

	// Test mixed types list
	mixedValue, err := config.Lookup("mixed")
	if err != nil {
		t.Fatalf("Failed to lookup mixed list: %v", err)
	}

	if mixedValue.Type != TypeList {
		t.Errorf("Expected list type, got %s", mixedValue.Type)
	}

	if len(mixedValue.ListVal) != 4 {
		t.Errorf("Expected 4 list elements, got %d", len(mixedValue.ListVal))
	}

	// Check individual element types
	expectedTypes := []ValueType{TypeString, TypeInt, TypeBool, TypeFloat}
	for i, expectedType := range expectedTypes {
		if mixedValue.ListVal[i].Type != expectedType {
			t.Errorf("Expected element[%d] type %s, got %s", i, expectedType, mixedValue.ListVal[i].Type)
		}
	}

	// Check values
	if mixedValue.ListVal[0].StrVal != "string" {
		t.Errorf("Expected 'string', got '%s'", mixedValue.ListVal[0].StrVal)
	}

	if mixedValue.ListVal[1].IntVal != 42 {
		t.Errorf("Expected 42, got %d", mixedValue.ListVal[1].IntVal)
	}

	if mixedValue.ListVal[2].BoolVal != true {
		t.Errorf("Expected true, got %t", mixedValue.ListVal[2].BoolVal)
	}

	if mixedValue.ListVal[3].FloatVal != 3.14 {
		t.Errorf("Expected 3.14, got %f", mixedValue.ListVal[3].FloatVal)
	}

	// Test nested list
	nestedValue, err := config.Lookup("nested")
	if err != nil {
		t.Fatalf("Failed to lookup nested list: %v", err)
	}

	if len(nestedValue.ListVal) != 3 {
		t.Errorf("Expected 3 nested elements, got %d", len(nestedValue.ListVal))
	}

	// First element should be a list
	if nestedValue.ListVal[0].Type != TypeList {
		t.Errorf("Expected first element to be list, got %s", nestedValue.ListVal[0].Type)
	}

	// Second element should be a group
	if nestedValue.ListVal[1].Type != TypeGroup {
		t.Errorf("Expected second element to be group, got %s", nestedValue.ListVal[1].Type)
	}

	// Third element should be an array
	if nestedValue.ListVal[2].Type != TypeArray {
		t.Errorf("Expected third element to be array, got %s", nestedValue.ListVal[2].Type)
	}

	// Test empty list
	emptyValue, err := config.Lookup("empty")
	if err != nil {
		t.Fatalf("Failed to lookup empty list: %v", err)
	}

	if len(emptyValue.ListVal) != 0 {
		t.Errorf("Expected empty list, got %d elements", len(emptyValue.ListVal))
	}
}

// Test list with trailing comma.
func TestListTrailingComma(t *testing.T) {
	config, err := ParseString(`values = ( "a", 42, true, );`)
	if err != nil {
		t.Fatalf("Failed to parse list with trailing comma: %v", err)
	}

	value, err := config.Lookup("values")
	if err != nil {
		t.Fatalf("Failed to lookup list: %v", err)
	}

	if len(value.ListVal) != 3 {
		t.Errorf("Expected list length 3, got %d", len(value.ListVal))
	}
}

// Test complex nested structures.
func TestComplexStructures(t *testing.T) {
	config, err := ParseString(`
		servers = [
			{
				name = "web1";
				host = "192.168.1.10";
				ports = [ 80, 443 ];
				features = ( "ssl", "cache", { type = "proxy"; target = "backend"; } );
			},
			{
				name = "web2";
				host = "192.168.1.11";
				ports = [ 80, 8080 ];
				features = ( "load_balancer", { type = "failover"; backup = true; } );
			}
		];
	`)
	if err != nil {
		t.Fatalf("Failed to parse complex structure: %v", err)
	}

	serversValue, err := config.Lookup("servers")
	if err != nil {
		t.Fatalf("Failed to lookup servers: %v", err)
	}

	if serversValue.Type != TypeArray {
		t.Errorf("Expected array type, got %s", serversValue.Type)
	}

	if len(serversValue.ArrayVal) != 2 {
		t.Errorf("Expected 2 servers, got %d", len(serversValue.ArrayVal))
	}

	// Check first server
	server1 := serversValue.ArrayVal[0]
	if server1.Type != TypeGroup {
		t.Errorf("Expected server to be group, got %s", server1.Type)
	}

	if server1.GroupVal["name"].StrVal != "web1" {
		t.Errorf("Expected name 'web1', got '%s'", server1.GroupVal["name"].StrVal)
	}

	// Check ports array in first server
	ports := server1.GroupVal["ports"]
	if ports.Type != TypeArray {
		t.Errorf("Expected ports to be array, got %s", ports.Type)
	}

	if len(ports.ArrayVal) != 2 {
		t.Errorf("Expected 2 ports, got %d", len(ports.ArrayVal))
	}

	// Check features list in first server
	features := server1.GroupVal["features"]
	if features.Type != TypeList {
		t.Errorf("Expected features to be list, got %s", features.Type)
	}

	if len(features.ListVal) != 3 {
		t.Errorf("Expected 3 features, got %d", len(features.ListVal))
	}
}

// Test different number formats.
func TestNumberFormats(t *testing.T) {
	tests := []struct {
		name     string
		config   string
		key      string
		expected int64
	}{
		{"decimal", `value = 42;`, "value", 42},
		{"hex_lowercase", `value = 0xff;`, "value", 255},
		{"hex_uppercase", `value = 0XFF;`, "value", 255},
		{"hex_mixed", `value = 0xAaBbCc;`, "value", 11189196},
		{"binary_lowercase", `value = 0b1010;`, "value", 10},
		{"binary_uppercase", `value = 0B1010;`, "value", 10},
		{"octal_new_lowercase", `value = 0o755;`, "value", 493},
		{"octal_new_uppercase", `value = 0O755;`, "value", 493},
		{"octal_new_q_lowercase", `value = 0q755;`, "value", 493},
		{"octal_new_q_uppercase", `value = 0Q755;`, "value", 493},
		// Note: negative hex/binary/octal are not supported in this implementation
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config, err := ParseString(tt.config)
			if err != nil {
				t.Fatalf("Failed to parse config: %v", err)
			}

			value, err := config.LookupInt64(tt.key)
			if err != nil {
				t.Fatalf("Failed to lookup int64: %v", err)
			}

			if value != tt.expected {
				t.Errorf("Expected %d, got %d", tt.expected, value)
			}
		})
	}
}

// Test large integers and long suffix.
func TestLongIntegers(t *testing.T) {
	tests := []struct {
		name     string
		config   string
		key      string
		expected int64
		isLong   bool
	}{
		{"small_with_L", `value = 42L;`, "value", 42, true},
		{"small_with_l", `value = 42l;`, "value", 42, true},
		{"max_int64_L", `value = 9223372036854775807L;`, "value", 9223372036854775807, true},
		{"min_int64_L", `value = -9223372036854775808L;`, "value", -9223372036854775808, true},
		{"hex_long", `value = 0xFFL;`, "value", 255, true},
		{"binary_long", `value = 0b111111111111111111111111111111111111111111111111111111111111111L;`, "value", 9223372036854775807, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config, err := ParseString(tt.config)
			if err != nil {
				t.Fatalf("Failed to parse config: %v", err)
			}

			value, err := config.LookupInt64(tt.key)
			if err != nil {
				t.Fatalf("Failed to lookup int64: %v", err)
			}

			if value != tt.expected {
				t.Errorf("Expected %d, got %d", tt.expected, value)
			}

			// Verify the raw value to check if it's stored as int64
			rawValue, err := config.Lookup(tt.key)
			if err != nil {
				t.Fatalf("Failed to lookup raw value: %v", err)
			}

			if tt.isLong && rawValue.Type != TypeInt64 {
				t.Errorf("Expected type Int64, got %s", rawValue.Type)
			}
		})
	}
}

// Test float scientific notation.
func TestScientificNotation(t *testing.T) {
	tests := []struct {
		name     string
		config   string
		key      string
		expected float64
	}{
		{"simple_e", `value = 1e3;`, "value", 1000.0},
		{"simple_E", `value = 1E3;`, "value", 1000.0},
		{"with_decimal_e", `value = 1.5e2;`, "value", 150.0},
		{"with_decimal_E", `value = 1.5E2;`, "value", 150.0},
		{"positive_exp", `value = 2.5e+3;`, "value", 2500.0},
		{"negative_exp", `value = 2.5e-3;`, "value", 0.0025},
		{"positive_exp_E", `value = 2.5E+3;`, "value", 2500.0},
		{"negative_exp_E", `value = 2.5E-3;`, "value", 0.0025},
		{"zero_exp", `value = 3.14e0;`, "value", 3.14},
		{"large_exp", `value = 1.23e10;`, "value", 12300000000.0},
		{"small_exp", `value = 1.23e-10;`, "value", 1.23e-10},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config, err := ParseString(tt.config)
			if err != nil {
				t.Fatalf("Failed to parse config: %v", err)
			}

			value, err := config.LookupFloat(tt.key)
			if err != nil {
				t.Fatalf("Failed to lookup float: %v", err)
			}

			if value != tt.expected {
				t.Errorf("Expected %g, got %g", tt.expected, value)
			}
		})
	}
}

// Test string escape sequences.
func TestStringEscapes(t *testing.T) {
	tests := []struct {
		name     string
		config   string
		key      string
		expected string
	}{
		{"newline", `value = "line1\nline2";`, "value", "line1\nline2"},
		{"tab", `value = "col1\tcol2";`, "value", "col1\tcol2"},
		{"carriage_return", `value = "line1\rline2";`, "value", "line1\rline2"},
		{"backspace", `value = "test\btest";`, "value", "test\btest"},
		{"form_feed", `value = "test\ftest";`, "value", "test\ftest"},
		{"bell", `value = "test\atest";`, "value", "test\atest"},
		{"vertical_tab", `value = "test\vtest";`, "value", "test\vtest"},
		{"backslash", `value = "path\\to\\file";`, "value", "path\\to\\file"},
		{"quote", `value = "He said, \"Hello\"";`, "value", "He said, \"Hello\""},
		{"hex_escape", `value = "ABC\x41\x42\x43";`, "value", "ABCABC"},
		{"multiple_escapes", `value = "line1\n\tline2\r\nline3";`, "value", "line1\n\tline2\r\nline3"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config, err := ParseString(tt.config)
			if err != nil {
				t.Fatalf("Failed to parse config: %v", err)
			}

			value, err := config.LookupString(tt.key)
			if err != nil {
				t.Fatalf("Failed to lookup string: %v", err)
			}

			if value != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, value)
			}
		})
	}
}

// Test string concatenation.
func TestStringConcatenation(t *testing.T) {
	tests := []struct {
		name     string
		config   string
		key      string
		expected string
	}{
		{
			"simple_concat",
			`value = "hello" "world";`,
			"value",
			"helloworld",
		},
		{
			"three_strings",
			`value = "one" "two" "three";`,
			"value",
			"onetwothree",
		},
		{
			"multiline_concat",
			`value = "This is a very long string that "
			        "spans multiple lines and is "
			        "automatically concatenated.";`,
			"value",
			"This is a very long string that spans multiple lines and is automatically concatenated.",
		},
		{
			"concat_with_escapes",
			`value = "line1\n" "line2\n" "line3";`,
			"value",
			"line1\nline2\nline3",
		},
		{
			"empty_strings",
			`value = "" "hello" "" "world" "";`,
			"value",
			"helloworld",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config, err := ParseString(tt.config)
			if err != nil {
				t.Fatalf("Failed to parse config: %v", err)
			}

			value, err := config.LookupString(tt.key)
			if err != nil {
				t.Fatalf("Failed to lookup string: %v", err)
			}

			if value != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, value)
			}
		})
	}
}

// Test edge cases and special values.
func TestEdgeCases(t *testing.T) {
	config, err := ParseString(`
		zero_int = 0;
		zero_float = 0.0;
		empty_string = "";
		space_string = " ";
		unicode_string = "hello";
		max_int = 2147483647;
		min_int = -2147483648;
		max_int64 = 9223372036854775807L;
		min_int64 = -9223372036854775808L;
		very_small_float = 1e-100;
		very_large_float = 1e100;
	`)
	if err != nil {
		t.Fatalf("Failed to parse edge cases config: %v", err)
	}

	// Test zero values
	zeroInt, err := config.LookupInt("zero_int")
	if err != nil || zeroInt != 0 {
		t.Errorf("Expected zero int, got %d", zeroInt)
	}

	zeroFloat, err := config.LookupFloat("zero_float")
	if err != nil || zeroFloat != 0.0 {
		t.Errorf("Expected zero float, got %f", zeroFloat)
	}

	// Test empty and space strings
	emptyStr, err := config.LookupString("empty_string")
	if err != nil || emptyStr != "" {
		t.Errorf("Expected empty string, got %q", emptyStr)
	}

	spaceStr, err := config.LookupString("space_string")
	if err != nil || spaceStr != " " {
		t.Errorf("Expected space string, got %q", spaceStr)
	}

	// Test simple string
	simpleStr, err := config.LookupString("unicode_string")
	if err != nil || simpleStr != "hello" {
		t.Errorf("Expected simple string, got %q", simpleStr)
	}

	// Test max/min values
	maxInt, err := config.LookupInt("max_int")
	if err != nil || maxInt != 2147483647 {
		t.Errorf("Expected max int, got %d", maxInt)
	}

	minInt, err := config.LookupInt("min_int")
	if err != nil || minInt != -2147483648 {
		t.Errorf("Expected min int, got %d", minInt)
	}

	maxInt64, err := config.LookupInt64("max_int64")
	if err != nil || maxInt64 != 9223372036854775807 {
		t.Errorf("Expected max int64, got %d", maxInt64)
	}

	minInt64, err := config.LookupInt64("min_int64")
	if err != nil || minInt64 != -9223372036854775808 {
		t.Errorf("Expected min int64, got %d", minInt64)
	}

	// Test very small and large floats
	verySmall, err := config.LookupFloat("very_small_float")
	if err != nil || verySmall != 1e-100 {
		t.Errorf("Expected very small float, got %g", verySmall)
	}

	veryLarge, err := config.LookupFloat("very_large_float")
	if err != nil || veryLarge != 1e100 {
		t.Errorf("Expected very large float, got %g", veryLarge)
	}
}

// Test array type homogeneity enforcement.
func TestArrayTypeHomogeneity(t *testing.T) {
	// These should fail because arrays must be homogeneous
	invalidArrays := []string{
		`values = [ "string", 42 ];`,             // string and int
		`values = [ 1, 2.5 ];`,                   // int and float
		`values = [ true, "false" ];`,            // bool and string
		`values = [ 1, true ];`,                  // int and bool
		`values = [ 1.5, true ];`,                // float and bool
		`values = [ "test", { key = "val"; } ];`, // string and group
	}

	for i, configStr := range invalidArrays {
		t.Run(fmt.Sprintf("invalid_array_%d", i), func(t *testing.T) {
			_, err := ParseString(configStr)
			if err == nil {
				t.Error("Expected error for heterogeneous array, but parsing succeeded")
			}
		})
	}
}

// Test @include directive (basic functionality).
func TestIncludeDirective(t *testing.T) {
	// Test that @include fails when file doesn't exist
	configStr := `
		name = "main";
		@include "nonexistent.cfg"
		port = 8080;
	`

	_, err := ParseString(configStr)
	if err == nil {
		t.Fatal("Expected error for missing include file, but got none")
	}

	// Should contain information about the missing file
	if !strings.Contains(err.Error(), "include file 'nonexistent.cfg' not found") {
		t.Errorf("Expected include file error, got: %v", err)
	}
}

// Test @include within groups.
func TestIncludeInGroups(t *testing.T) {
	configStr := `
		database = {
			host = "localhost";
			@include "db_settings.cfg"
			port = 5432;
		};
	`

	_, err := ParseString(configStr)
	if err == nil {
		t.Fatal("Expected error for missing include file in group, but got none")
	}

	// Should contain information about the missing file
	if !strings.Contains(err.Error(), "include file 'db_settings.cfg' not found") {
		t.Errorf("Expected include file error, got: %v", err)
	}
}

// Test error cases.
func TestErrorCases(t *testing.T) {
	errorTests := []struct {
		name   string
		config string
	}{
		// Note: unterminated strings are handled gracefully by the lexer
		{"invalid_assignment", `name "missing equals";`},
		{"missing_identifier", `= "value";`},
		{"unterminated_array", `values = [ "one", "two"`},
		{"unterminated_group", `group = { key = "value"`},
		{"unterminated_list", `list = ( "one", "two"`},
		{"invalid_number", `value = 12.34.56;`},
		{"invalid_hex", `value = 0xGGG;`},
		{"invalid_binary", `value = 0b123;`},
		{"invalid_token", `value = @invalid;`},
		{"duplicate_decimal", `value = 12..34;`},
		{"empty_hex", `value = 0x;`},
		{"empty_binary", `value = 0b;`},
		{"empty_octal", `value = 0o;`},
	}

	for _, tt := range errorTests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ParseString(tt.config)
			if err == nil {
				t.Errorf("Expected error for %s, but parsing succeeded", tt.name)
			}
		})
	}
}

// Test syntax error reporting.
func TestSyntaxErrorReporting(t *testing.T) {
	// Test that error messages include line and column information
	configStr := `
		name = "test";
		port = invalid_value;
		debug = true;
	`

	_, err := ParseString(configStr)
	if err == nil {
		t.Fatal("Expected syntax error")
	}

	// Error should mention the line number (3 in this case)
	errMsg := err.Error()
	if !strings.Contains(errMsg, "line") {
		t.Errorf("Error message should contain line information: %s", errMsg)
	}
}

// Test various whitespace handling.
func TestWhitespaceHandling(t *testing.T) {
	tests := []struct {
		name   string
		config string
	}{
		{
			"tabs_and_spaces",
			"\tname\t=\t\"test\";\n  port  =  8080  ;",
		},
		{
			"multiple_newlines",
			"name = \"test\";\n\n\nport = 8080;",
		},
		{
			"windows_line_endings",
			"name = \"test\";\r\nport = 8080;\r\n",
		},
		{
			"mixed_whitespace",
			" \t name \t = \t \"test\" \t ; \r\n \t port \t = \t 8080 \t ; \r\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config, err := ParseString(tt.config)
			if err != nil {
				t.Fatalf("Failed to parse config with whitespace: %v", err)
			}

			name, err := config.LookupString("name")
			if err != nil || name != "test" {
				t.Errorf("Expected 'test', got '%s'", name)
			}

			port, err := config.LookupInt("port")
			if err != nil || port != 8080 {
				t.Errorf("Expected 8080, got %d", port)
			}
		})
	}
}

// Test value construction helpers.
func TestValueConstructors(t *testing.T) {
	// Test the New*Value helper functions
	intVal := NewIntValue(42)
	if intVal.Type != TypeInt || intVal.IntVal != 42 {
		t.Errorf("NewIntValue failed: got type %s, value %d", intVal.Type, intVal.IntVal)
	}

	int64Val := NewInt64Value(9223372036854775807)
	if int64Val.Type != TypeInt64 || int64Val.Int64Val != 9223372036854775807 {
		t.Errorf("NewInt64Value failed: got type %s, value %d", int64Val.Type, int64Val.Int64Val)
	}

	floatVal := NewFloatValue(3.14159)
	if floatVal.Type != TypeFloat || floatVal.FloatVal != 3.14159 {
		t.Errorf("NewFloatValue failed: got type %s, value %f", floatVal.Type, floatVal.FloatVal)
	}

	boolVal := NewBoolValue(true)
	if boolVal.Type != TypeBool || boolVal.BoolVal != true {
		t.Errorf("NewBoolValue failed: got type %s, value %t", boolVal.Type, boolVal.BoolVal)
	}

	strVal := NewStringValue("hello")
	if strVal.Type != TypeString || strVal.StrVal != "hello" {
		t.Errorf("NewStringValue failed: got type %s, value %q", strVal.Type, strVal.StrVal)
	}

	arrayVal := NewArrayValue([]Value{intVal, NewIntValue(84)})
	if arrayVal.Type != TypeArray || len(arrayVal.ArrayVal) != 2 {
		t.Errorf("NewArrayValue failed: got type %s, length %d", arrayVal.Type, len(arrayVal.ArrayVal))
	}

	groupVal := NewGroupValue(map[string]Value{"key": strVal})
	if groupVal.Type != TypeGroup || len(groupVal.GroupVal) != 1 {
		t.Errorf("NewGroupValue failed: got type %s, length %d", groupVal.Type, len(groupVal.GroupVal))
	}

	listVal := NewListValue([]Value{intVal, strVal, boolVal})
	if listVal.Type != TypeList || len(listVal.ListVal) != 3 {
		t.Errorf("NewListValue failed: got type %s, length %d", listVal.Type, len(listVal.ListVal))
	}
}

// Test type string representations.
func TestTypeStringRepresentation(t *testing.T) {
	tests := []struct {
		valueType ValueType
		expected  string
	}{
		{TypeInt, "int"},
		{TypeInt64, "int64"},
		{TypeFloat, "float"},
		{TypeBool, "bool"},
		{TypeString, "string"},
		{TypeArray, "array"},
		{TypeGroup, "group"},
		{TypeList, "list"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := tt.valueType.String()
			if result != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result)
			}
		})
	}

	// Test unknown type
	unknownType := ValueType(999)
	if unknownType.String() != "unknown" {
		t.Errorf("Expected 'unknown' for invalid type, got %q", unknownType.String())
	}
}

// Test int to int64 conversion edge cases.
func TestIntInt64Conversion(t *testing.T) {
	config, err := ParseString(`
		regular_int = 42;
		large_int = 2147483647;         // max int32
		too_large = 9223372036854775807L; // explicit int64
		int64_val = 9223372036854775807L;
	`)
	if err != nil {
		t.Fatalf("Failed to parse config: %v", err)
	}

	// Test regular int can be retrieved as int64
	val, err := config.LookupInt64("regular_int")
	if err != nil || val != 42 {
		t.Errorf("Expected 42 as int64, got %d", val)
	}

	// Test large int
	val, err = config.LookupInt64("large_int")
	if err != nil || val != 2147483647 {
		t.Errorf("Expected 2147483647 as int64, got %d", val)
	}

	// Test explicit int64
	val, err = config.LookupInt64("too_large")
	if err != nil || val != 9223372036854775807 {
		t.Errorf("Expected 9223372036854775807 as int64, got %d", val)
	}

	// Check that explicit int64 is stored correctly
	rawValue, err := config.Lookup("too_large")
	if err != nil {
		t.Fatalf("Failed to lookup raw value: %v", err)
	}

	if rawValue.Type != TypeInt64 {
		t.Error("Explicit int64 value should be stored as int64")
	}
}

// Test complex real-world configuration.
func TestRealWorldConfig(t *testing.T) {
	complexConfig := `
		# Application configuration
		app = {
			name = "MyApp";
			version = "1.2.3";
			debug = false;

			# Server settings
			server = {
				host = "0.0.0.0";
				port = 8080;
				ssl = {
					enabled = true;
					cert_file = "/etc/ssl/cert.pem";
					key_file = "/etc/ssl/key.pem";
				};
			};

			# Database configuration
			database = {
				driver = "postgresql";
				connection = {
					host = "localhost";
					port = 5432;
					database = "myapp_db";
					username = "myapp_user";
					password = "secure_password_123";
				};
				pool = {
					min_connections = 5;
					max_connections = 100;
					idle_timeout = 300.0;
				};
			};

			# Feature flags
			features = {
				new_ui = true;
				analytics = false;
				beta_features = [ "feature_a", "feature_b", "feature_c" ];
			};

			# Log levels for different components
			logging = {
				level = "INFO";
				components = (
					{ name = "database"; level = "DEBUG"; },
					{ name = "auth"; level = "WARN"; },
					{ name = "api"; level = "INFO"; }
				);
			};
		};

		# External service configurations
		services = [
			{
				name = "payment_gateway";
				url = "https://api.payment.com";
				timeout = 30;
				retries = 3;
				api_key = "secret_key_12345";
			},
			{
				name = "email_service";
				url = "https://api.email.com";
				timeout = 15;
				retries = 2;
				api_key = "email_key_67890";
			}
		];

		# Monitoring and metrics
		monitoring = {
			enabled = true;
			interval = 60;  // seconds
			metrics = [ "cpu", "memory", "disk", "network" ];
			thresholds = {
				cpu_usage = 80.0;
				memory_usage = 85.0;
				disk_usage = 90.0;
			};
		};
	`

	config, err := ParseString(complexConfig)
	if err != nil {
		t.Fatalf("Failed to parse complex config: %v", err)
	}

	// Test various lookups
	appName, err := config.LookupString("app.name")
	if err != nil || appName != "MyApp" {
		t.Errorf("Expected app name 'MyApp', got '%s'", appName)
	}

	serverPort, err := config.LookupInt("app.server.port")
	if err != nil || serverPort != 8080 {
		t.Errorf("Expected server port 8080, got %d", serverPort)
	}

	sslEnabled, err := config.LookupBool("app.server.ssl.enabled")
	if err != nil || sslEnabled != true {
		t.Errorf("Expected SSL enabled true, got %t", sslEnabled)
	}

	dbHost, err := config.LookupString("app.database.connection.host")
	if err != nil || dbHost != "localhost" {
		t.Errorf("Expected DB host 'localhost', got '%s'", dbHost)
	}

	idleTimeout, err := config.LookupFloat("app.database.pool.idle_timeout")
	if err != nil || idleTimeout != 300.0 {
		t.Errorf("Expected idle timeout 300.0, got %f", idleTimeout)
	}

	// Test array access
	betaFeatures, err := config.Lookup("app.features.beta_features")
	if err != nil {
		t.Fatalf("Failed to lookup beta features: %v", err)
	}

	if len(betaFeatures.ArrayVal) != 3 {
		t.Errorf("Expected 3 beta features, got %d", len(betaFeatures.ArrayVal))
	}

	// Test list access
	loggingComponents, err := config.Lookup("app.logging.components")
	if err != nil {
		t.Fatalf("Failed to lookup logging components: %v", err)
	}

	if len(loggingComponents.ListVal) != 3 {
		t.Errorf("Expected 3 logging components, got %d", len(loggingComponents.ListVal))
	}

	// Test services array
	services, err := config.Lookup("services")
	if err != nil {
		t.Fatalf("Failed to lookup services: %v", err)
	}

	if len(services.ArrayVal) != 2 {
		t.Errorf("Expected 2 services, got %d", len(services.ArrayVal))
	}

	// Test monitoring metrics array
	metrics, err := config.Lookup("monitoring.metrics")
	if err != nil {
		t.Fatalf("Failed to lookup metrics: %v", err)
	}

	if len(metrics.ArrayVal) != 4 {
		t.Errorf("Expected 4 metrics, got %d", len(metrics.ArrayVal))
	}

	// Test threshold values
	cpuThreshold, err := config.LookupFloat("monitoring.thresholds.cpu_usage")
	if err != nil || cpuThreshold != 80.0 {
		t.Errorf("Expected CPU threshold 80.0, got %f", cpuThreshold)
	}
}

func TestEscapeSequenceInRegex(t *testing.T) {
	input := `test_pattern = "\/\*\s*dde='([^*]|\*[^\/]|)*\*\/\s*$";`

	config, err := ParseString(input)
	if err != nil {
		t.Fatalf("Failed to parse: %v", err)
	}

	pattern, err := config.LookupString("test_pattern")
	if err != nil {
		t.Fatalf("Failed to lookup test_pattern: %v", err)
	}

	expected := "/\\*\\s*dde='([^*]|\\*[^/]|)*\\*/\\s*$"
	if pattern != expected {
		t.Errorf("Expected %q, got %q", expected, pattern)
	}
}
